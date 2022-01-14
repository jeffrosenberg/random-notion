package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	selection "github.com/jeffrosenberg/random-notion/internal/pageselection"
	"github.com/jeffrosenberg/random-notion/internal/persistence"
	"github.com/jeffrosenberg/random-notion/pkg/logging"
	"github.com/jeffrosenberg/random-notion/pkg/notion"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

const (
	SecretName   = "random-notion/notion-api"
	SecretRegion = "us-west-2"
)

type HandlerFn func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error)

type AwsSecret struct {
	Token      string `json:"token"`
	DatabaseId string `json:"database_id"`
}

// Closure for injection of notion.PageGetter interface
func handleRequestForApi(api notion.PageGetter, selector selection.PageSelector,
	db dynamodbiface.DynamoDBAPI) HandlerFn {
	return func(ctx context.Context, e events.APIGatewayV2HTTPRequest) (event events.APIGatewayV2HTTPResponse, err error) {
		var dto *persistence.NotionDTO
		var apiPages []notion.Page
		execStartTime := time.Now().Unix()
		databaseId := api.GetDatabaseId()

		logger := logging.GetLoggerWithContext(ctx)
		logger.Trace().
			Str("function", "handleRequestForApi").
			Str("log_level", logger.GetLevel().String()).
			Msg("Random Notion handler triggered")

		if databaseId == "" {
			logger.Warn().Msg("No DatabaseId provided")
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 400,
				Body:       "No DatabaseId provided",
			}, nil
		}

		// Recover from a panic and return a 500 error
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%v", r)
				logger.
					Err(err).
					Interface("dto", dto).
					Interface("api_pages", apiPages).
					Interface("pagegetter", api).
					Interface("pageselector", selector).
					Str("database_id", databaseId).
					Msg("Recovered from a panic")
				event = events.APIGatewayV2HTTPResponse{
					StatusCode: 500,
					Body:       "Internal server error",
				}
			}
		}()

		// 1. Get cached pages from DynamoDb
		logger.Trace().Msg("Getting pages from DynamoDb")
		dto, err = persistence.GetPages(db, &databaseId)
		if dto == nil {
			if err != nil {
				logger.Err(err).Msg("Unable to read cached data from DynamoDb")
			}
			// We could still read from the API, so set dto to a stub and keep going
			dto = &persistence.NotionDTO{
				DatabaseId: databaseId,
				Pages:      []notion.Page{},
				LastQuery:  execStartTime,
			}
		}

		// 2. Get additional pages from the Notion API
		logger.Trace().Msg("Getting pages from Notion API")
		apiPages, err = api.GetPagesSinceTime(time.Unix(dto.LastQuery, 0))
		if err != nil {
			logger.Err(err).Msg("Unable to read pages from Notion API")
			// We could still read from the API, so set apiPages to a stub and keep going
			apiPages = []notion.Page{}
		}

		logger.Debug().
			Int("pages_cached", len(dto.Pages)).
			Int("pages_api", len(apiPages)).
			Msg("Retrieved pages")

		if len(dto.Pages) == 0 && len(apiPages) == 0 {
			if err != nil {
				logger.Err(err).Send()
				return events.APIGatewayV2HTTPResponse{
					StatusCode: 400,
					Body:       err.Error(),
				}, err
			} else {
				// No error, but no pages available: Return 204 No Content
				// This is a tough scenario to pick a status code for,
				// but should also be encountered rarely
				logger.Warn().Msg("No records found")
				return events.APIGatewayV2HTTPResponse{
					StatusCode: 204,
				}, nil
			}
		}

		// 3. Dedup and combine both sources of pages
		logger.Trace().Msg("Unioning pages")
		pagesAdded := selection.UnionPages(dto, apiPages)
		if pagesAdded {
			dto.LastQuery = execStartTime
			persistence.PutPages(db, dto)
		}
		selectedPage := selector.SelectPage(dto.Pages)

		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Body:       fmt.Sprintf("{\"id\":\"%s\", \"url\":\"%s\"}", selectedPage.Id, selectedPage.Url),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}
}

// Code snippet via AWS docs:
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html
func setApiSecrets(api *notion.ApiConfig, sess *session.Session) {
	//Create a Secrets Manager client
	svc := secretsmanager.New(sess, aws.NewConfig().WithRegion(SecretRegion))
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(SecretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	// In this sample we only handle the specific exceptions for the 'GetSecretValue' API.
	// See https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
	result, err := svc.GetSecretValue(input)
	if err != nil {
		logger := logging.GetLogger()
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				logger.Err(aerr).Msg(secretsmanager.ErrCodeDecryptionFailure)

			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				logger.Err(aerr).Msg(secretsmanager.ErrCodeInternalServiceError)

			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				logger.Err(aerr).Msg(secretsmanager.ErrCodeInvalidParameterException)

			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				logger.Err(aerr).Msg(secretsmanager.ErrCodeInvalidRequestException)

			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				logger.Err(aerr).Msg(secretsmanager.ErrCodeResourceNotFoundException)
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			logger.Err(aerr).Send()
		}
	}

	if result.SecretString != nil {
		var secret AwsSecret
		json.Unmarshal([]byte(*result.SecretString), &secret)
		api.SecretToken = secret.Token
		api.DatabaseId = secret.DatabaseId
	} else {
		panic("Unable to retrieve API secrets")
	}
}

func main() {
	// Initialize interfaces
	api := notion.NewApiConfig()
	selector := &selection.RandomPage{}
	sess := session.Must(session.NewSession())
	setApiSecrets(api, sess)
	db := dynamodb.New(sess)

	lambda.Start(handleRequestForApi(api, selector, db))
}
