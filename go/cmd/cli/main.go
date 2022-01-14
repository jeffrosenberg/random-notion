package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	selection "github.com/jeffrosenberg/random-notion/internal/pageselection"
	"github.com/jeffrosenberg/random-notion/internal/persistence"
	"github.com/jeffrosenberg/random-notion/pkg/notion"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

const (
	SecretName   = "random-notion/notion-api"
	SecretRegion = "us-west-2"
)

type AwsSecret struct {
	Token      string `json:"token"`
	DatabaseId string `json:"database_id"`
}

func exec(api notion.PageGetter, selector selection.PageSelector,
	db dynamodbiface.DynamoDBAPI) (string, error) {
	execStartTime := time.Now().Unix()
	databaseId := api.GetDatabaseId()

	// 1. Get cached pages from DynamoDb
	dto, err := persistence.GetPages(db, &databaseId)
	if dto == nil {
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to read cached data from DynamoDb")
		}
		// We could still read from the API, so set dto to a stub and keep going
		dto = &persistence.NotionDTO{
			DatabaseId: databaseId,
			Pages:      []notion.Page{},
			LastQuery:  execStartTime,
		}
	}

	// 2. Get additional pages from the Notion API
	apiPages, err := api.GetPagesSinceTime(time.Unix(dto.LastQuery, 0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to read pages from Notion API")
		// We could still read from the API, so set apiPages to a stub and keep going
		apiPages = []notion.Page{}
	}

	if len(dto.Pages) == 0 && len(apiPages) == 0 {
		if err != nil {
			return "No records found", err
		} else {
			// No error, but no pages available
			return "No records found", nil
		}
	}

	// 3. Dedup and combine both sources of pages
	pagesAdded := selection.UnionPages(dto, apiPages)
	if pagesAdded {
		dto.LastQuery = execStartTime
		persistence.PutPages(db, dto)
	}
	selectedPage := selector.SelectPage(dto.Pages)

	return selectedPage.Url, nil
}

// Code snippet via AWS docs:
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html
func setApiSecrets(api *notion.ApiConfig, sess *session.Session) {
	//Create a Secrets Manager client
	svc := secretsmanager.New(sess, aws.NewConfig().WithRegion(SecretRegion))
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(SecretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	// In this sample we only handle the specific exceptions for the 'GetSecretValue' API.
	// See https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
	result, err := svc.GetSecretValue(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	if result.SecretString != nil {
		var secret AwsSecret
		json.Unmarshal([]byte(*result.SecretString), &secret)
		api.SecretToken = secret.Token
		api.DatabaseId = secret.DatabaseId
	} else {
		panic("Unable to retrieve API secrets")
	}
}

func main() {
	// Parse command-line arguments and create a config object
	url := flag.String("url", notion.API_URI, "Base URL of the Notion API")
	databaseId := flag.String("databaseId", "", "Notion Database ID")
	secret := flag.String("secret", "", "Notion API secret token")
	pageSize := flag.Uint("pageSize", uint(notion.DEFAULT_PAGE_SIZE), "Pages to retrieve per Notion API call")
	flag.Parse()

	// Initialize interfaces
	api := &notion.ApiConfig{
		Url:         *url,
		DatabaseId:  *databaseId,
		SecretToken: *secret,
		PageSize:    uint8(*pageSize),
	}
	selector := &selection.RandomPage{}
	sess := session.Must(session.NewSession())
	if api.DatabaseId == "" || api.SecretToken == "" {
		setApiSecrets(api, sess)
	}
	db := dynamodb.New(sess)

	output, err := exec(api, selector, db)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	fmt.Fprintln(os.Stdout, output)
}
