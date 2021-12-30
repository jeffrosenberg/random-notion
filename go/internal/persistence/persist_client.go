package persistence

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

const NOTION_TABLE_NAME string = "TODO"

type NotionDTO struct {
	DatabaseId string
	Pages      []notion.Page
	NextCursor string
}

func GetPages(client dynamodbiface.DynamoDBAPI, databaseId *string) (dto *NotionDTO, err error) {
	dto = &NotionDTO{}

	req := &dynamodb.GetItemInput{
		TableName: aws.String(NOTION_TABLE_NAME),
		Key:       map[string]*dynamodb.AttributeValue{"DatabaseId": {S: databaseId}},
	}

	// TODO: Add logging for DynamoDb call?
	output, err := client.GetItem(req)
	if err != nil {
		return
	}

	if len(output.Item) == 0 {
		return
	}

	err = dynamodbattribute.UnmarshalMap(output.Item, dto)
	return
}

func PutPages(client dynamodbiface.DynamoDBAPI, dto *NotionDTO) (err error) {
	inputItem, err := dynamodbattribute.MarshalMap(dto)
	if err != nil {
		return fmt.Errorf("Unable to generate DynamoDb input: %w", err)
	}

	req := &dynamodb.PutItemInput{
		Item:         inputItem,
		ReturnValues: aws.String("NONE"),
		TableName:    aws.String(NOTION_TABLE_NAME),
	}

	_, err = client.PutItem(req)
	if err != nil {
		return fmt.Errorf("Error inserting to DynamoDb: %w", err)
	}

	return nil
}
