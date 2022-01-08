package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/jeffrosenberg/random-notion/internal/pageselection"
	"github.com/jeffrosenberg/random-notion/internal/persistence"
	"github.com/jeffrosenberg/random-notion/pkg/notion"

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

type AwsSecret struct {
	Token      string `json:"token"`
	DatabaseId string `json:"database_id"`
}

func exec(api notion.PageGetter, selector pageselection.PageSelector,
	db dynamodbiface.DynamoDBAPI, databaseId string) (string, error) {
	// 1. Get cached pages from DynamoDb
	dto, err := persistence.GetPages(db, &databaseId, api.GetLogger())
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
	apiPages, err := api.GetPages(dto.NextCursor)
	if err != nil {
		api.GetLogger().Err(err).Msg("Unable to read pages from Notion API")
		// We could still read from the API, so set apiPages to a stub and keep going
		apiPages = []notion.Page{}
	}

	if len(dto.Pages) == 0 && len(apiPages) == 0 {
		if err != nil {
			return "No records found", err
		} else {
			// No error, but no pages available
			return "No records found", nil
		}
	}

	// 3. Dedup and combine both sources of pages
	pagesAdded := pageselection.UnionPages(dto, apiPages, api.GetLogger())
	if pagesAdded {
		persistence.PutPages(db, dto, api.GetLogger())
	}
	selectedPage := selector.SelectPage(dto.Pages)

	return selectedPage.Url, nil
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
	// Parse command-line arguments and create a config object
	url := flag.String("url", notion.API_URI, "Base URL of the Notion API")
	databaseId := flag.String("databaseId", "", "Notion Database ID")
	secret := flag.String("secret", "", "Notion API secret token")
	pageSize := flag.Uint("pageSize", uint(notion.DEFAULT_PAGE_SIZE), "Pages to retrieve per Notion API call")
	flag.Parse()

	// Initialize interfaces
	api := &notion.ApiConfig{
		Url:         *url,
		DatabaseId:  *databaseId,
		SecretToken: *secret,
		PageSize:    uint8(*pageSize),
		Logger:      &log.Logger,
	}
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	selector := &pageselection.RandomPage{}
	sess := session.Must(session.NewSession())
	if api.DatabaseId == "" || api.SecretToken == "" {
		setApiSecrets(api, sess)
	}
	db := dynamodb.New(sess)

	output, err := exec(api, selector, db, api.DatabaseId)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	fmt.Fprintln(os.Stdout, output)
}
