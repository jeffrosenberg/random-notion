package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/jeffrosenberg/random-notion/configs"
)

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

func GetPages(config *configs.NotionConfig) (*[]Page, error) {
	url, err := url.Parse(fmt.Sprintf("%s/databases/%s/query", config.ApiUrl, config.DatabaseId))
	if err != nil {
		return nil, fmt.Errorf("Unable to parse URL: %w", err)
	}

	client := &http.Client{}
	postBody := map[string]string{"page_size": "10"}
	jsonValue, _ := json.Marshal(postBody)
	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(jsonValue))
	req.Header.Set("Notion-Version", "2021-08-16")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SecretToken))
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
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
