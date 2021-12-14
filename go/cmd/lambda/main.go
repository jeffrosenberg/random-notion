package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jeffrosenberg/random-notion/configs"
	"github.com/jeffrosenberg/random-notion/internal/randompage"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

var config *configs.NotionConfig

func main() {
	config = &configs.NotionConfig{
		ApiUrl:      configs.API_URI,
		DatabaseId:  configs.TEMP_DATABASE_ID,
		SecretToken: configs.TEMP_TOKEN,
		PageSize:    configs.DEFAULT_PAGE_SIZE,
	}
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, e events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// request context
	lc, _ := lambdacontext.FromContext(ctx)
	log.Println("Random Notion GoLang function triggered")
	log.Printf("REQUEST ID: %s", lc.AwsRequestID)
	log.Printf("FUNCTION NAME: %s", lambdacontext.FunctionName)

	randomPage, err := randompage.GetRandomPage(config)
	if err != nil {
		log.Printf("Encountered an error")
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       err.Error(),
		}, err
	} else {
		log.Printf("Returned Page ID: %s", randomPage.Id)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Body:       fmt.Sprintf("{\"id\":\"%s\", \"url\":\"%s\"}", randomPage.Id, randomPage.Url),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}
}
