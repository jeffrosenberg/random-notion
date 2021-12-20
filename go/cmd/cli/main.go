package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/jeffrosenberg/random-notion/internal/randompage"
	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

func main() {
	// Parse command-line arguments and create a config object
	url := flag.String("url", notion.API_URI, "Base URL of the Notion API")
	databaseId := flag.String("databaseId", "", "Notion Databse ID")
	secret := flag.String("secret", "", "Notion API secret token")
	pageSize := flag.Uint("pageSize", uint(notion.DEFAULT_PAGE_SIZE), "Pages to retrieve per Notion API call")
	flag.Parse()

	if *databaseId == "" {
		fmt.Fprintln(os.Stderr, "Database ID required")
		flag.Usage()
		os.Exit(1)
	} else if *secret == "" {
		fmt.Fprintln(os.Stderr, "Notion API secret token required")
		flag.Usage()
		os.Exit(1)
	}

	api := &notion.ApiConfig{
		Url:         *url,
		DatabaseId:  *databaseId,
		SecretToken: *secret,
		PageSize:    uint8(*pageSize),
		Logger:      &log.Logger,
	}
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	randomPage, err := randompage.GetRandomPage(api)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Get page failed")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Fprintln(os.Stdout, randomPage.Url)
}
