package persistence

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/jeffrosenberg/random-notion/pkg/logging"

	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

const DEFAULT_TABLE_NAME string = "random-notion-cache"

var tableName string

type NotionDTO struct {
	DatabaseId string        `dynamodbav:"database_id"`
	Pages      []notion.Page `dynamodbav:"pages"`
	LastQuery  int64         `dynamodbav:"last_query,omitempty"`
}

func GetPages(client dynamodbiface.DynamoDBAPI, databaseId *string) (dto *NotionDTO, err error) {
	defer logging.LogFunction(
		"persistence.GetPages", time.Now(), "Getting pages from DynamoDb",
		map[string]interface{}{
			"table_name":  getTableName(),
			"database_id": *databaseId,
		},
	)

	dto = &NotionDTO{
		DatabaseId: *databaseId,
	}
	req := &dynamodb.GetItemInput{
		TableName: aws.String(getTableName()),
		Key:       map[string]*dynamodb.AttributeValue{"database_id": {S: databaseId}},
	}

	output, err := client.GetItem(req)
	if err != nil {
		logging.GetLogger().Err(err).Send()
		return
	}

	if len(output.Item) == 0 {
		return
	}

	err = dynamodbattribute.UnmarshalMap(output.Item, dto)
	return
}

func PutPages(client dynamodbiface.DynamoDBAPI, dto *NotionDTO) (err error) {
	defer logging.LogFunction(
		"persistence.PutPages", time.Now(), "Putting pages to DynamoDb",
		map[string]interface{}{
			"table_name": getTableName(),
			"pages":      len(dto.Pages),
			"notion_dto": *dto,
		},
	)

	inputItem, err := dynamodbattribute.MarshalMap(dto)
	if err != nil {
		logging.GetLogger().Err(err)
		return fmt.Errorf("Unable to generate DynamoDb input: %w", err)
	}

	req := &dynamodb.PutItemInput{
		Item:         inputItem,
		ReturnValues: aws.String("NONE"),
		TableName:    aws.String(getTableName()),
	}

	_, err = client.PutItem(req)
	if err != nil {
		logging.GetLogger().Err(err)
		return fmt.Errorf("Error inserting to DynamoDb: %w", err)
	}

	return nil
}

func getTableName() string {
	if tableName == "" {
		t := os.Getenv("CACHE_TABLE_NAME")
		if t == "" {
			t = DEFAULT_TABLE_NAME
		}
		tableName = t
	}
	return tableName
}
