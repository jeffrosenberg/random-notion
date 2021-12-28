package persistence

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	_ "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

const NOTION_TABLE_NAME string = "TODO"

type NotionPages struct { // TODO: flesh out properties
	DatabaseId string
	Pages      []notion.Page
}

func GetPages(client dynamodbiface.DynamoDBAPI, databaseId *string) (data *NotionPages, err error) {
	data = &NotionPages{}

	req := &dynamodb.GetItemInput{
		TableName: aws.String(NOTION_TABLE_NAME),
		Key:       map[string]*dynamodb.AttributeValue{"DatabaseId": {S: databaseId}},
	}

	output, err := client.GetItem(req)
	if err != nil {
		return
	}

	if len(output.Item) == 0 {
		return
	}

	err = dynamodbattribute.UnmarshalMap(output.Item, data)
	return
}

func PutPages(client dynamodbiface.DynamoDBAPI, pages *[]notion.Page, databaseId *string) (err error) {
	pagesAttr, err := dynamodbattribute.MarshalList(pages)
	if err != nil {
		return fmt.Errorf("Unable to generate DynamoDb input: %w", err)
	}
	inputItem := map[string]*dynamodb.AttributeValue{
		"DatabaseId": {S: databaseId},
		"Pages":      {L: pagesAttr},
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
