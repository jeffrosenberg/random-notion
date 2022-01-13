package persistence

import (
	"strconv"
	"testing"
	"time"

	"github.com/jeffrosenberg/random-notion/pkg/notion"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var databaseId string = "99999999-abcd-efgh-1234-000000000000"
var nextCursor string = "5331da24-6597-4f2d-a684-fd94a0f3278a"

const mockTimestring string = "2021-12-10 01:00:00 CST"
const mockTimestamp int64 = 1639119600

func mockTime() *time.Time {
	result, _ := time.Parse("2006-01-02 15:04:05 MST", mockTimestring)
	return &result
}

func mockLogger() *zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	return &zerolog.Logger{}
}

type MockDynamoDb struct {
	mock.Mock
	dynamodbiface.DynamoDBAPI
	// Set in unit test arrangement to "populate" the database
	MockDbContents map[string]*dynamodb.AttributeValue
}

func (mock MockDynamoDb) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return &dynamodb.GetItemOutput{
		ConsumedCapacity: nil,
		Item:             mock.MockDbContents,
	}, nil
}

func (mock MockDynamoDb) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	mock.MethodCalled("PutItem", input)
	return &dynamodb.PutItemOutput{}, nil
}

func TestGetNoPagesFoundReturnsDefault(t *testing.T) {
	// Arrange
	mockClient := MockDynamoDb{}
	mockClient.MockDbContents = make(map[string]*dynamodb.AttributeValue)

	// Act
	result, err := GetPages(mockClient, &databaseId, mockLogger())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, databaseId, result.DatabaseId)
	assert.Nil(t, result.Pages)
	assert.Equal(t, int64(0), result.LastQuery)
}

func TestGetReturnsNotionPages(t *testing.T) {
	// Arrange
	mockClient := MockDynamoDb{}
	mockClient.MockDbContents = testDataDynamoDbAttribute

	expected := testDataStruct

	// Act
	result, err := GetPages(mockClient, &databaseId, mockLogger())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, databaseId, result.DatabaseId)
	assert.Equal(t, expected.Pages, result.Pages)
	assert.Equal(t, mockTimestamp, result.LastQuery)
}

func TestPutPagesCalled(t *testing.T) {
	// Arrange
	mockClient := MockDynamoDb{}
	mockClient.Mock.On("PutItem", mock.Anything) // Assert that PutItem is called

	// Act
	err := PutPages(mockClient, testDataStruct, mockLogger())

	// Assert
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

var testDataDynamoDbAttribute map[string]*dynamodb.AttributeValue = map[string]*dynamodb.AttributeValue{
	"database_id": {S: aws.String(databaseId)},
	"pages": {
		L: []*dynamodb.AttributeValue{
			{
				M: map[string]*dynamodb.AttributeValue{
					"id":               {S: aws.String("3350ba04-48b1-43e3-8726-1b1e9828b2b3")},
					"created_time":     {S: aws.String("2021-11-05T12:54:00.000Z")},
					"last_edited_time": {S: aws.String("2021-11-05T12:55:00.000Z")},
					"url":              {S: aws.String("https://www.notion.so/Initial-goals-3350ba0448b143e387261b1e9828b2b3")},
				},
			},
			{
				M: map[string]*dynamodb.AttributeValue{
					"id":               {S: aws.String("5331da24-6597-4f2d-a684-fd94a0f3278a")},
					"created_time":     {S: aws.String("2021-11-01T01:01:00.000Z")},
					"last_edited_time": {S: aws.String("2021-11-01T13:24:00.000Z")},
					"url":              {S: aws.String("https://www.notion.so/Chicken-korma-recipe-How-to-make-chicken-korma-Swasthi-s-Recipes-5331da2465974f2da684fd94a0f3278a")},
				},
			},
		},
	},
	"last_query": {N: aws.String(strconv.FormatInt(mockTimestamp, 10))},
}

var testDataStruct *NotionDTO = &NotionDTO{
	DatabaseId: databaseId,
	Pages: []notion.Page{
		{
			Id:             "3350ba04-48b1-43e3-8726-1b1e9828b2b3",
			CreatedTime:    "2021-11-05T12:54:00.000Z",
			LastEditedTime: "2021-11-05T12:55:00.000Z",
			Url:            "https://www.notion.so/Initial-goals-3350ba0448b143e387261b1e9828b2b3",
		},
		{
			Id:             "5331da24-6597-4f2d-a684-fd94a0f3278a",
			CreatedTime:    "2021-11-01T01:01:00.000Z",
			LastEditedTime: "2021-11-01T13:24:00.000Z",
			Url:            "https://www.notion.so/Chicken-korma-recipe-How-to-make-chicken-korma-Swasthi-s-Recipes-5331da2465974f2da684fd94a0f3278a",
		},
	},
	LastQuery: mockTimestamp,
}
