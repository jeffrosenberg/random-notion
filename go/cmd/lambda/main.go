package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/jeffrosenberg/random-notion/internal/pageselection"
	"github.com/jeffrosenberg/random-notion/internal/persistence"
	"github.com/jeffrosenberg/random-notion/pkg/notion"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"

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

// Inject via CDK configuration
var (
	CommitID string
	LogLevel string = "1" // must be convertable to zerolog.Level - Debug = 0, Info = 1, Trace = -1
)

type HandlerFn func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error)

type AwsSecret struct {
	Token      string `json:"token"`
	DatabaseId string `json:"database_id"`
}

// Closure for injection of notion.PageGetter interface
// TODO: Revisit whether there's a more elegant way to get databaseId
func handleRequestForApi(api notion.PageGetter, selector pageselection.PageSelector,
	db dynamodbiface.DynamoDBAPI, databaseId string) HandlerFn {
	return func(ctx context.Context, e events.APIGatewayV2HTTPRequest) (event events.APIGatewayV2HTTPResponse, err error) {
		var dto *persistence.NotionDTO
		var apiPages []notion.Page

		createLogger(ctx, api)
		api.GetLogger().Debug().
			Str("function", "handleRequestForApi").
			Str("log_level", api.GetLogger().GetLevel().String()).
			Msg("Random Notion handler triggered")

		// Recover from a panic and return a 500 error
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%v", r)
				api.GetLogger().
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
		api.GetLogger().Debug().Msg("Getting pages from DynamoDb")
		dto, err = persistence.GetPages(db, &databaseId, api.GetLogger())
		if dto == nil {
			if err != nil {
				api.GetLogger().Err(err).Msg("Unable to read cached data from DynamoDb")
			}
			// We could still read from the API, so set dto to a stub and keep going
			dto = &persistence.NotionDTO{
				DatabaseId: databaseId,
				Pages:      []notion.Page{},
				NextCursor: "",
			}
		}

		// 2. Get additional pages from the Notion API
		api.GetLogger().Debug().
			Str("function", "handleRequestForApi").
			Msg("Getting pages from Notion API")
		apiPages, err = api.GetPages(dto.NextCursor)
		if err != nil {
			api.GetLogger().Err(err).Msg("Unable to read pages from Notion API")
			// We could still read from the API, so set apiPages to a stub and keep going
			apiPages = []notion.Page{}
		}

		api.GetLogger().Info().
			Int("pages_cached", len(dto.Pages)).
			Int("pages_api", len(apiPages)).
			Msg("Retrieved pages")

		if len(dto.Pages) == 0 && len(apiPages) == 0 {
			if err != nil {
				api.GetLogger().Err(err).Send()
				return events.APIGatewayV2HTTPResponse{
					StatusCode: 400,
					Body:       err.Error(),
				}, err
			} else {
				// No error, but no pages available: Return 204 No Content
				// This is a tough scenario to pick a status code for,
				// but should also be encountered rarely
				api.GetLogger().Warn().Msg("No records found")
				return events.APIGatewayV2HTTPResponse{
					StatusCode: 204,
				}, nil
			}
		}

		// 3. Dedup and combine both sources of pages
		api.GetLogger().Debug().
			Str("function", "handleRequestForApi").
			Msg("Unioning pages")
		pagesAdded := pageselection.UnionPages(dto, apiPages, api.GetLogger())
		if pagesAdded {
			persistence.PutPages(db, dto, api.GetLogger())
		}
		selectedPage := selector.SelectPage(dto.Pages)

		api.GetLogger().Debug().
			Str("page_id", selectedPage.Id).
			Str("page_url", selectedPage.Url).
			Send()
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Body:       fmt.Sprintf("{\"id\":\"%s\", \"url\":\"%s\"}", selectedPage.Id, selectedPage.Url),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}
}

func createLogger(ctx context.Context, api notion.PageGetter) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Read LogLevel from linker value
	var level zerolog.Level
	convertedLevel, err := strconv.ParseInt(LogLevel, 0, 8)
	if err != nil {
		level = zerolog.InfoLevel
		log.Warn().Str("LogLevel", LogLevel).Msg("Unable to convert LogLevel to zerolog.Level")
	} else {
		level = zerolog.Level(convertedLevel)
	}

	logger := log.Logger
	lc, ok := lambdacontext.FromContext(ctx) // request context
	if ok && lc != nil {
		logger = log.With().
			Str("commit_id", CommitID).
			Str("request_id", lc.AwsRequestID).
			Str("function_name", lambdacontext.FunctionName).
			Logger().
			Level(zerolog.Level(level))
		// zerolog usage note: must use Msg() or Send() to trigger logs to actually send
		logger.Info().Str("log_level", logger.GetLevel().String()).Msg("Logging initialized")
	} else {
		log.Warn().Msg("Lambda context not found")
	}
	api.SetLogger(&logger)
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
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				api.GetLogger().Err(aerr).Msg(secretsmanager.ErrCodeDecryptionFailure)

			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				api.GetLogger().Err(aerr).Msg(secretsmanager.ErrCodeInternalServiceError)

			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				api.GetLogger().Err(aerr).Msg(secretsmanager.ErrCodeInvalidParameterException)

			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				api.GetLogger().Err(aerr).Msg(secretsmanager.ErrCodeInvalidRequestException)

			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				api.GetLogger().Err(aerr).Msg(secretsmanager.ErrCodeResourceNotFoundException)
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			api.GetLogger().Err(aerr).Send()
		}
		return
	}

	if result.SecretString != nil {
		var secret AwsSecret
		json.Unmarshal([]byte(*result.SecretString), &secret)
		api.SecretToken = secret.Token
		api.DatabaseId = secret.DatabaseId
		api.GetLogger().Info().Msg("Retrieved API secrets")
	} else {
		panic("Unable to retrieve API secrets")
	}
}

func main() {
	// Initialize interfaces
	api := notion.NewApiConfig()
	selector := &pageselection.RandomPage{}
	sess := session.Must(session.NewSession())
	setApiSecrets(api, sess)
	db := dynamodb.New(sess)

	lambda.Start(handleRequestForApi(api, selector, db, api.DatabaseId))
}
