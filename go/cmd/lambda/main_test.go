package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/jeffrosenberg/random-notion/pkg/notion"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type TestApiConfig struct {
	mock.Mock
	pages []notion.Page
}

type TestSelector struct {
	mock.Mock
}

type TestDynamoDb struct {
	mock.Mock
	dynamodbiface.DynamoDBAPI
	outputMap map[string]*dynamodb.AttributeValue
}

type TestPanic struct {
	mock.Mock
	dynamodbiface.DynamoDBAPI
}

const mockDatabaseId = "99999999abcdefgh1234000000000000"
const mockPageId = "3350ba04-48b1-43e3-8726-1b1e9828b2b3"
const mockPageUrl = "https://www.notion.so/Initial-goals-3350ba0448b143e387261b1e9828b2b3"
const mockTime = "2021-11-05T12:54:00.000Z"
const nextCursor string = "5331da24-6597-4f2d-a684-fd94a0f3278a"

func (api *TestApiConfig) GetPages(cursor string) ([]notion.Page, error) {
	api.MethodCalled("GetPages", cursor)

	if api.pages == nil {
		return nil, fmt.Errorf("No pages found")
	}

	// Slice api.pages depending on cursor input; if no cursor, take all
	index := 0
	if cursor != "" {
		for i, v := range api.pages {
			if v.Id == cursor {
				index = i
				break
			}
		}
	}

	return api.pages[index:], nil
}

func (api *TestApiConfig) GetAllPages() ([]notion.Page, error) {
	api.MethodCalled("GetAllPages")

	if api.pages == nil {
		return nil, fmt.Errorf("No pages found")
	}

	return api.pages, nil
}

func (api *TestApiConfig) GetLogger() *zerolog.Logger {
	api.MethodCalled("GetLogger")
	return &log.Logger
}

func (api *TestApiConfig) SetLogger(logger *zerolog.Logger) {
	return // no action for tests
}

func (selector *TestSelector) SelectPage(pages []notion.Page) *notion.Page {
	selector.MethodCalled("SelectPage")
	if len(pages) == 0 {
		return nil
	} else if len(pages) == 1 {
		return &pages[0]
	}
	return &pages[len(pages)-1]
}

func (db *TestDynamoDb) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	db.MethodCalled("GetItem", input)
	return &dynamodb.GetItemOutput{
		Item: db.outputMap,
	}, nil
}

func (db *TestDynamoDb) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	db.MethodCalled("PutItem", input)
	return &dynamodb.PutItemOutput{}, nil
}

func (db *TestPanic) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	db.MethodCalled("GetItem", input)
	panic("A panic happened!")
}

func TestRetrieveAllRecordsFromNotionApi(t *testing.T) {
	// Arrange
	api := &TestApiConfig{
		pages: []notion.Page{
			{
				Id:             mockPageId,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl,
			},
		},
	}
	selector := &TestSelector{}
	db := &TestDynamoDb{}
	event := events.APIGatewayV2HTTPRequest{}
	api.Mock.On("GetPages", "") // Set expectations for mock methods
	api.Mock.On("GetLogger")
	selector.Mock.On("SelectPage")
	db.Mock.On("GetItem", mock.Anything)
	db.Mock.On("PutItem", mock.Anything)
	handler := handleRequestForApi(api, selector, db, mockDatabaseId)

	// Act
	result, err := handler(context.Background(), event)

	// Assert
	require.NoError(t, err)
	expected := events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("{\"id\":\"%s\", \"url\":\"%s\"}", mockPageId, mockPageUrl),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
	assert.EqualValues(t, expected, result)
	api.AssertExpectations(t)
	selector.AssertExpectations(t)
	db.AssertExpectations(t)
}

func TestFailGracefullyWhenNoPagesExist(t *testing.T) {
	// Arrange
	api := &TestApiConfig{
		pages: []notion.Page{},
	}
	selector := &TestSelector{}
	db := &TestDynamoDb{
		outputMap: make(map[string]*dynamodb.AttributeValue),
	}
	event := events.APIGatewayV2HTTPRequest{}
	api.Mock.On("GetPages", "") // Set expectations for mock methods
	api.Mock.On("GetLogger")
	// selector.Mock.On("SelectPage") // PageSelector methods should NOT be called
	db.Mock.On("GetItem", mock.Anything)
	handler := handleRequestForApi(api, selector, db, mockDatabaseId)

	// Act
	result, err := handler(context.Background(), event)

	// Assert
	require.NoError(t, err)
	expected := events.APIGatewayV2HTTPResponse{
		StatusCode: 204,
		Body:       "",
		Headers:    map[string]string(nil),
	}
	assert.EqualValues(t, expected, result)
	api.AssertExpectations(t)
	db.AssertExpectations(t)
}

func TestRetrieveAllRecordsFromDynamoDb(t *testing.T) {
	// Arrange
	api := &TestApiConfig{
		pages: []notion.Page{},
	}
	selector := &TestSelector{}
	db := &TestDynamoDb{
		outputMap: map[string]*dynamodb.AttributeValue{
			"database_id": {S: aws.String(mockDatabaseId)},
			"pages": {
				L: []*dynamodb.AttributeValue{
					{
						M: map[string]*dynamodb.AttributeValue{
							"id":               {S: aws.String(mockPageId)},
							"created_time":     {S: aws.String(mockTime)},
							"last_edited_time": {S: aws.String(mockTime)},
							"url":              {S: aws.String(mockPageUrl)},
						},
					},
				},
			},
		},
	}
	event := events.APIGatewayV2HTTPRequest{}
	api.Mock.On("GetPages", "") // Set expectations for mock methods
	api.Mock.On("GetLogger")
	selector.Mock.On("SelectPage")
	db.Mock.On("GetItem", mock.Anything)
	// db.Mock.On("PutItem", mock.Anything) // PutItem should NOT be called
	handler := handleRequestForApi(api, selector, db, mockDatabaseId)

	// Act
	result, err := handler(context.Background(), event)

	// Assert
	require.NoError(t, err)
	expected := events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("{\"id\":\"%s\", \"url\":\"%s\"}", mockPageId, mockPageUrl),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
	assert.EqualValues(t, expected, result)
	api.AssertExpectations(t)
	selector.AssertExpectations(t)
	db.AssertExpectations(t)
}

func TestRetrieveRecordsFromMultipleSources(t *testing.T) {
	// Arrange
	api := &TestApiConfig{
		pages: []notion.Page{
			{
				Id:             mockPageId,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl,
			},
			{
				Id:             "5331da24-6597-4f2d-a684-fd94a0f3278a",
				CreatedTime:    "2021-11-01T01:01:00.000Z",
				LastEditedTime: "2021-11-01T13:24:00.000Z",
				Url:            "https://www.notion.so/Chicken-korma-recipe-How-to-make-chicken-korma-Swasthi-s-Recipes-5331da2465974f2da684fd94a0f3278a",
			},
		},
	}
	selector := &TestSelector{}
	db := &TestDynamoDb{
		outputMap: map[string]*dynamodb.AttributeValue{
			"database_id": {S: aws.String(mockDatabaseId)},
			"pages": {
				L: []*dynamodb.AttributeValue{
					{
						M: map[string]*dynamodb.AttributeValue{
							"id":               {S: aws.String(mockPageId)},
							"created_time":     {S: aws.String(mockTime)},
							"last_edited_time": {S: aws.String(mockTime)},
							"url":              {S: aws.String(mockPageUrl)},
						},
					},
				},
			},
			"next_cursor": {S: aws.String("5331da24-6597-4f2d-a684-fd94a0f3278a")},
		},
	}
	event := events.APIGatewayV2HTTPRequest{}
	api.Mock.On("GetPages", "5331da24-6597-4f2d-a684-fd94a0f3278a") // Set expectations for mock methods
	api.Mock.On("GetLogger")
	selector.Mock.On("SelectPage")
	db.Mock.On("GetItem", mock.Anything)
	db.Mock.On("PutItem", mock.Anything)
	handler := handleRequestForApi(api, selector, db, mockDatabaseId)

	// Act
	result, err := handler(context.Background(), event)

	// Assert
	require.NoError(t, err)
	expected := events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("{\"id\":\"%s\", \"url\":\"%s\"}", "5331da24-6597-4f2d-a684-fd94a0f3278a", "https://www.notion.so/Chicken-korma-recipe-How-to-make-chicken-korma-Swasthi-s-Recipes-5331da2465974f2da684fd94a0f3278a"),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
	assert.EqualValues(t, expected, result)
	api.AssertExpectations(t)
	selector.AssertExpectations(t)
	db.AssertExpectations(t)
}

func TestNoNewRecordsInApi(t *testing.T) {
	// Arrange
	api := &TestApiConfig{
		pages: []notion.Page{
			{
				Id:             mockPageId,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl,
			},
		},
	}
	selector := &TestSelector{}
	db := &TestDynamoDb{
		outputMap: map[string]*dynamodb.AttributeValue{
			"database_id": {S: aws.String(mockDatabaseId)},
			"pages": {
				L: []*dynamodb.AttributeValue{
					{
						M: map[string]*dynamodb.AttributeValue{
							"id":               {S: aws.String(mockPageId)},
							"created_time":     {S: aws.String(mockTime)},
							"last_edited_time": {S: aws.String(mockTime)},
							"url":              {S: aws.String(mockPageUrl)},
						},
					},
				},
			},
			"next_cursor": {S: aws.String(mockPageId)},
		},
	}
	event := events.APIGatewayV2HTTPRequest{}
	api.Mock.On("GetPages", mockPageId) // Set expectations for mock methods
	api.Mock.On("GetLogger")
	selector.Mock.On("SelectPage")
	db.Mock.On("GetItem", mock.Anything)
	// db.Mock.On("PutItem", mock.Anything) // PutItem should NOT be called
	handler := handleRequestForApi(api, selector, db, mockDatabaseId)

	// Act
	result, err := handler(context.Background(), event)

	// Assert
	require.NoError(t, err)
	expected := events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("{\"id\":\"%s\", \"url\":\"%s\"}", mockPageId, mockPageUrl),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
	assert.EqualValues(t, expected, result)
	api.AssertExpectations(t)
	selector.AssertExpectations(t)
	db.AssertExpectations(t)
}

func TestRecoverFromPanic(t *testing.T) {
	// Arrange
	api := &TestApiConfig{
		pages: []notion.Page{
			{
				Id:             mockPageId,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl,
			},
		},
	}
	selector := &TestSelector{}
	db := &TestPanic{}
	event := events.APIGatewayV2HTTPRequest{}
	api.Mock.On("GetLogger")
	db.Mock.On("GetItem", mock.Anything)
	handler := handleRequestForApi(api, selector, db, mockDatabaseId)

	// Act
	result, err := handler(context.Background(), event)

	// Assert
	require.NoError(t, err)
	expected := events.APIGatewayV2HTTPResponse{
		StatusCode: 500,
		Body:       "Internal server error",
	}
	assert.EqualValues(t, expected, result)
	api.AssertExpectations(t)
	db.AssertExpectations(t)
}

func TestHandleNotionApiError(t *testing.T) {
	// Arrange
	api := &TestApiConfig{}
	selector := &TestSelector{}
	db := &TestDynamoDb{}
	event := events.APIGatewayV2HTTPRequest{}
	api.Mock.On("GetPages", "") // Set expectations for mock methods
	api.Mock.On("GetLogger")
	// selector.Mock.On("SelectPage") // PageSelector methods should NOT be called
	db.Mock.On("GetItem", mock.Anything)
	handler := handleRequestForApi(api, selector, db, mockDatabaseId)

	// Act
	result, err := handler(context.Background(), event)

	// Assert
	require.Error(t, err)
	expected := events.APIGatewayV2HTTPResponse{
		StatusCode: 400,
		Body:       "No pages found",
	}
	assert.EqualValues(t, expected, result)
	api.AssertExpectations(t)
	selector.AssertExpectations(t)
}

// TODO: remove if not using this
// var testDataDynamoDbAttribute map[string]*dynamodb.AttributeValue = map[string]*dynamodb.AttributeValue{
// 	"DatabaseId": {S: aws.String(mockDatabaseId)},
// 	"Pages": {
// 		L: []*dynamodb.AttributeValue{
// 			{
// 				M: map[string]*dynamodb.AttributeValue{
// 					"id":               {S: aws.String("3350ba04-48b1-43e3-8726-1b1e9828b2b3")},
// 					"created_time":     {S: aws.String("2021-11-05T12:54:00.000Z")},
// 					"last_edited_time": {S: aws.String("2021-11-05T12:55:00.000Z")},
// 					"url":              {S: aws.String("https://www.notion.so/Initial-goals-3350ba0448b143e387261b1e9828b2b3")},
// 				},
// 			},
// 			{
// 				M: map[string]*dynamodb.AttributeValue{
// 					"id":               {S: aws.String("5331da24-6597-4f2d-a684-fd94a0f3278a")},
// 					"created_time":     {S: aws.String("2021-11-01T01:01:00.000Z")},
// 					"last_edited_time": {S: aws.String("2021-11-01T13:24:00.000Z")},
// 					"url":              {S: aws.String("https://www.notion.so/Chicken-korma-recipe-How-to-make-chicken-korma-Swasthi-s-Recipes-5331da2465974f2da684fd94a0f3278a")},
// 				},
// 			},
// 		},
// 	},
// 	"NextCursor": {S: aws.String(nextCursor)},
// }
