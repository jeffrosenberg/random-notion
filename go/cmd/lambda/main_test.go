package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/jeffrosenberg/random-notion/pkg/notion"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestApiConfig struct {
	page *notion.Page
}

func (api *TestApiConfig) GetPages() (*[]notion.Page, error) {
	if api.page == nil {
		return nil, fmt.Errorf("No pages found")
	}

	return &[]notion.Page{
		*api.page,
	}, nil
}

func TestHandleRequest_Success(t *testing.T) {
	api := &TestApiConfig{
		page: &notion.Page{
			Id:             "3350ba04-48b1-43e3-8726-1b1e9828b2b3",
			CreatedTime:    "2021-11-05T12:54:00.000Z",
			LastEditedTime: "2021-11-05T12:55:00.000Z",
			Url:            "https://www.notion.so/Initial-goals-3350ba0448b143e387261b1e9828b2b3",
		},
	}
	event := events.APIGatewayV2HTTPRequest{}
	handler := handleRequestForApi(api)

	result, err := handler(context.Background(), event)
	require.NoError(t, err)

	expected := events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("{\"id\":\"%s\", \"url\":\"%s\"}", api.page.Id, api.page.Url),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
	assert.EqualValues(t, expected, result)
}

func TestHandleRequest_Error(t *testing.T) {
	api := &TestApiConfig{}
	event := events.APIGatewayV2HTTPRequest{}
	handler := handleRequestForApi(api)

	result, err := handler(context.Background(), event)
	require.Error(t, err)

	expected := events.APIGatewayV2HTTPResponse{
		StatusCode: 400,
		Body:       "No pages found",
	}
	assert.EqualValues(t, expected, result)
}
