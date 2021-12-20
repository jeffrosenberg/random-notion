package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/jeffrosenberg/random-notion/internal/randompage"
	"github.com/jeffrosenberg/random-notion/pkg/notion"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

const (
	SecretName   = "random-notion/notion-api"
	SecretRegion = "us-west-2"
)

// Inject via CDK configuration
var (
	CommitID string
	LogLevel string // must be convertable to zerolog.Level
)

type HandlerFn func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error)

type AwsSecret struct {
	Token      string `json:"token"`
	DatabaseId string `json:"database_id"`
}

// Closure for injection of `api`
func handleRequestForApi(api notion.PageGetter) HandlerFn {
	return func(ctx context.Context, e events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		createLogger(ctx, api)
		api.GetLogger().Info().
			Str("log_level", api.GetLogger().GetLevel().String()).
			Msg("Random Notion handler triggered")

		randomPage, err := randompage.GetRandomPage(api)
		if err != nil {
			api.GetLogger().Err(err)
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 400,
				Body:       err.Error(),
			}, err
		} else {
			api.GetLogger().Debug().
				Str("page_id", randomPage.Id).
				Str("page_url", randomPage.Url).
				Send()
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Body:       fmt.Sprintf("{\"id\":\"%s\", \"url\":\"%s\"}", randomPage.Id, randomPage.Url),
				Headers:    map[string]string{"Content-Type": "application/json"},
			}, nil
		}
	}
}

func createLogger(ctx context.Context, api notion.PageGetter) {
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
func setApiSecrets(api *notion.ApiConfig) {
	//Create a Secrets Manager client
	sess, err := session.NewSession()
	if err != nil {
		// Handle session creation error
		fmt.Println(err.Error())
		return
	}
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
	api := notion.NewApiConfig()
	setApiSecrets(api)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	lambda.Start(handleRequestForApi(api))
}
