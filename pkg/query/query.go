package query

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const API_URI = "https://api.notion.com/v1"
const TEMP_TOKEN = "secret_jdINX4JHB9LSHbImH0zQUzsEmYaBHjCn8XcagrHmWau"
const TEMP_DATABASE_ID = "45d3242e5c6d4a3bb99e4aa4db83f015"

type Database struct {
	Id             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Url            string `json:"url"`
}

type Page struct { // TODO: more properties
	Id             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Url            string `json:"url"`
}

type PageResponse struct {
	Object  string `json:"object"`
	Results []Page `json:"results"`
}

func GetDatabase(notionApiUrl string) (*Database, error) {
	url, err := url.Parse(fmt.Sprintf("%s/databases/%s", notionApiUrl, TEMP_DATABASE_ID))
	if err != nil {
		return nil, fmt.Errorf("Unable to parse URL: %w", err)
	}

	res, err := http.Get(url.String())
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received invalid status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Enable to read response body: %w", err)
	}

	var db Database
	json.Unmarshal(body, &db)

	return &db, nil
}

func GetPages(notionApiUrl string) (*[]Page, error) {
	url, err := url.Parse(fmt.Sprintf("%s/databases/%s/query", notionApiUrl, TEMP_DATABASE_ID))
	if err != nil {
		return nil, fmt.Errorf("Unable to parse URL: %w", err)
	}

	postBody := map[string]string{"page_size": "10"}
	jsonValue, _ := json.Marshal(postBody)

	res, err := http.Post(url.String(), "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received invalid status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Enable to read response body: %w", err)
	}

	var pages PageResponse
	json.Unmarshal(body, &pages)

	return &pages.Results, nil
}
