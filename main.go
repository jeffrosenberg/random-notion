package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/jeffrosenberg/random-notion/configs"
	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

const API_URI = "https://api.notion.com/v1"
const PAGE_SIZE = uint(100)
const TEMP_TOKEN = "secret_jdINX4JHB9LSHbImH0zQUzsEmYaBHjCn8XcagrHmWau"
const TEMP_DATABASE_ID = "45d3242e5c6d4a3bb99e4aa4db83f015"

func main() {
	// Parse command-line arguments and create a config object
	url := flag.String("url", API_URI, "Base URL of the Notion API")
	databaseId := flag.String("databaseId", TEMP_DATABASE_ID, "Notion Databse ID")
	secret := flag.String("secret", TEMP_TOKEN, "Notion API secret token")
	pageSize := flag.Uint("pageSize", PAGE_SIZE, "Pages to retrieve per Notion API call")
	flag.Parse()

	config := &configs.NotionConfig{
		ApiUrl:      *url,
		DatabaseId:  *databaseId,
		SecretToken: *secret,
		PageSize:    uint8(*pageSize),
	}

	pages, err := notion.GetPages(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Get pages failed")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	rand.Seed(time.Now().UnixNano())
	randomPage := (*pages)[rand.Intn(len(*pages))]
	fmt.Fprintln(os.Stdout, randomPage.Url)
}
