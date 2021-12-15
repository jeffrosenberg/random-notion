package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jeffrosenberg/random-notion/configs"
	"github.com/jeffrosenberg/random-notion/internal/randompage"
	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

func main() {
	// Parse command-line arguments and create a config object
	url := flag.String("url", configs.API_URI, "Base URL of the Notion API")
	databaseId := flag.String("databaseId", configs.TEMP_DATABASE_ID, "Notion Databse ID")
	secret := flag.String("secret", configs.TEMP_TOKEN, "Notion API secret token")
	pageSize := flag.Uint("pageSize", uint(configs.PAGE_SIZE), "Pages to retrieve per Notion API call")
	flag.Parse()

	api := &notion.ApiConfig{
		Url:         *url,
		DatabaseId:  *databaseId,
		SecretToken: *secret,
		PageSize:    uint8(*pageSize),
	}

	randomPage, err := randompage.GetRandomPage(api)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Get page failed")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Fprintln(os.Stdout, randomPage.Url)
}
