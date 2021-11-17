package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jeffrosenberg/random-notion/configs"
	"github.com/jeffrosenberg/random-notion/pkg/get_random"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

const API_URI = "https://api.notion.com/v1"
const TEMP_TOKEN = "secret_jdINX4JHB9LSHbImH0zQUzsEmYaBHjCn8XcagrHmWau"
const TEMP_DATABASE_ID = "45d3242e5c6d4a3bb99e4aa4db83f015"

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, e events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	config := &configs.NotionConfig{
		ApiUrl:      API_URI,
		DatabaseId:  TEMP_DATABASE_ID,
		SecretToken: TEMP_TOKEN,
		PageSize:    configs.DEFAULT_PAGE_SIZE,
	}

	// request context
	lc, _ := lambdacontext.FromContext(ctx)
	log.Println("Random Notion GoLang function triggered")
	log.Printf("REQUEST ID: %s", lc.AwsRequestID)
	log.Printf("FUNCTION NAME: %s", lambdacontext.FunctionName)

	randomPage, err := get_random.GetRandomPage(config)
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
