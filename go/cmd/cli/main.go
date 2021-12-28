package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/jeffrosenberg/random-notion/internal/pageselection"
	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

const (
	SecretName   = "random-notion/notion-api"
	SecretRegion = "us-west-2"
)

func execGetPage(api notion.PageGetter, selector pageselection.PageSelector) (string, error) {
	pages, err := api.GetPages()
	if err != nil {
		return "", err
	}

	selectedPage := selector.SelectPage(pages)
	return selectedPage.Url, nil
}

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
	selector := &pageselection.RandomPage{}

	output, err := execGetPage(api, selector)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	fmt.Fprintln(os.Stdout, output)
}
