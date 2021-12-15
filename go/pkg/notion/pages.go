package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Page struct { // TODO: pull in more properties?
	Id             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Url            string `json:"url"`
}

type PageGetter interface {
	GetPages() (*[]Page, error)
}

type pageRequest struct {
	PageSize    uint8  `json:"page_size"`
	StartCursor string `json:"start_cursor,omitempty"`
}

type pageResponse struct {
	Object  string `json:"object"`
	Results []Page `json:"results"`
	Next    string `json:"next_cursor"`
	HasMore bool   `json:"has_more"`
}

func (api *ApiConfig) GetPages() (*[]Page, error) {
	pages := []Page{}
	hasMore := true
	cursor := ""

	for hasMore == true {
		response, err := api.queryPages(cursor)
		if err != nil {
			return nil, err // TODO: More robust error handling
		}
		pages = append(pages, response.Results...)
		hasMore = response.HasMore
		cursor = response.Next
	}

	return &pages, nil
}

func (api *ApiConfig) queryPages(cursor string) (pageResponse, error) {
	url, err := url.Parse(fmt.Sprintf("%s/databases/%s/query", api.Url, api.DatabaseId))
	if err != nil {
		return pageResponse{}, fmt.Errorf("Unable to parse URL: %w", err)
	}

	client := &http.Client{}
	postBody := pageRequest{
		PageSize:    api.PageSize,
		StartCursor: cursor,
	}
	jsonValue, _ := json.Marshal(postBody)
	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(jsonValue))
	req.Header.Set("Notion-Version", "2021-08-16")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.SecretToken))
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return pageResponse{}, fmt.Errorf("Unable to retrieve response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return pageResponse{}, fmt.Errorf("Received invalid status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return pageResponse{}, fmt.Errorf("Enable to read response body: %w", err)
	}

	var pageResponse pageResponse
	json.Unmarshal(body, &pageResponse)
	return pageResponse, nil
}
