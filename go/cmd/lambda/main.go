package main

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/jeffrosenberg/random-notion/configs"
	"github.com/jeffrosenberg/random-notion/internal/randompage"
	"github.com/jeffrosenberg/random-notion/pkg/notion"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

var CommitID string

type HandlerFn func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error)

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
	logger := log.Logger
	lc, ok := lambdacontext.FromContext(ctx) // request context
	if ok && lc != nil {
		logger = log.With().
			Str("commit_id", CommitID).
			Str("request_id", lc.AwsRequestID).
			Str("function_name", lambdacontext.FunctionName).
			Logger().
			Level(zerolog.DebugLevel)
		// zerolog usage note: must use Msg() or Send() to trigger logs to actually send
		logger.Info().Str("log_level", logger.GetLevel().String()).Msg("Logging initialized")
	} else {
		log.Warn().Msg("Lambda context not found")
	}
	api.SetLogger(&logger)
}

func main() {
	api := &notion.ApiConfig{
		Url:         configs.API_URI,
		DatabaseId:  configs.TEMP_DATABASE_ID,
		SecretToken: configs.TEMP_TOKEN,
		PageSize:    configs.PAGE_SIZE,
		Logger:      nil,
	}
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	lambda.Start(handleRequestForApi(api))
}
