package persistence

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

const DEFAULT_TABLE_NAME string = "random-notion-cache"

var tableName string

type NotionDTO struct {
	DatabaseId string        `dynamodbav:"database_id"`
	Pages      []notion.Page `dynamodbav:"pages"`
	NextCursor string        `dynamodbav:"next_cursor,omitempty"`
}

func GetPages(client dynamodbiface.DynamoDBAPI, databaseId *string, logger *zerolog.Logger) (dto *NotionDTO, err error) {
	if logger == nil {
		logger = &log.Logger
	}
	logger.Info().Str("function", "GetPages").Msg("Getting pages from DynamoDb")

	dto = &NotionDTO{
		DatabaseId: *databaseId,
	}

	logger.Trace().
		Str("table_name", getTableName()).
		Str("database_id", *databaseId).
		Msg("Preparing DynamoDb get item request")
	req := &dynamodb.GetItemInput{
		TableName: aws.String(getTableName()),
		Key:       map[string]*dynamodb.AttributeValue{"database_id": {S: databaseId}},
	}

	output, err := client.GetItem(req)
	if err != nil {
		logger.Err(err).Send()
		return
	}

	if len(output.Item) == 0 {
		return
	}

	err = dynamodbattribute.UnmarshalMap(output.Item, dto)
	return
}

func PutPages(client dynamodbiface.DynamoDBAPI, dto *NotionDTO, logger *zerolog.Logger) (err error) {
	if logger == nil {
		logger = &log.Logger
	}
	logger.Info().Str("function", "PutPages").Msg("Putting pages to DynamoDb")

	inputItem, err := dynamodbattribute.MarshalMap(dto)
	if err != nil {
		logger.Err(err)
		return fmt.Errorf("Unable to generate DynamoDb input: %w", err)
	}

	logger.Trace().
		Str("table_name", getTableName()).
		Int("pages", len(dto.Pages)).
		Interface("input_item", inputItem).
		Msg("Preparing DynamoDb put item request")
	req := &dynamodb.PutItemInput{
		Item:         inputItem,
		ReturnValues: aws.String("NONE"),
		TableName:    aws.String(getTableName()),
	}

	_, err = client.PutItem(req)
	if err != nil {
		logger.Err(err)
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
