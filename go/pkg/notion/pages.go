package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Page struct {
	Id             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Url            string `json:"url"`
}

type PageGetter interface {
	GetPages(string) ([]Page, error)
	GetAllPages() ([]Page, error)
	Logger
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

// Return pages from the Notion API, starting at an optional cursor string
func (api *ApiConfig) GetPages(cursor string) ([]Page, error) {
	api.Logger.Info().Str("function", "GetPages").Msg("Getting pages from API")
	pages := []Page{}
	hasMore := true

	for hasMore == true {
		response, err := api.queryPages(cursor)
		if err != nil {
			api.Logger.Err(err).Send()
			return nil, err // More robust error handling would be nice, but skipping as this is a hobby project
		}
		pages = append(pages, response.Results...)
		hasMore = response.HasMore
		cursor = response.Next
		api.Logger.Debug().
			Int("pages_retrieved", len(response.Results)).
			Bool("has_more", response.HasMore).
			Str("cursor", response.Next).
			Msg("Processed Notion API response")
	}

	return pages, nil
}

// Return all pages from the Notion API
func (api *ApiConfig) GetAllPages() ([]Page, error) {
	return api.GetPages("")
}

func (api *ApiConfig) queryPages(cursor string) (pageResponse, error) {
	url, err := url.Parse(fmt.Sprintf("%s/databases/%s/query", api.Url, api.DatabaseId))
	if err != nil {
		api.Logger.Err(err).Send() // Must use Msg() or Send() to trigger logs to actually send
		return pageResponse{}, fmt.Errorf("Unable to parse URL: %w", err)
	}

	client := &http.Client{}
	postBody := pageRequest{
		PageSize:    api.PageSize,
		StartCursor: cursor,
	}
	jsonValue, _ := json.Marshal(postBody)
	api.Logger.Trace().
		Str("request_verb", "POST").
		Str("request_url", url.String()).
		RawJSON("request_json", jsonValue).
		Msg("Prepared Notion API request")
	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(jsonValue))
	if err != nil {
		api.Logger.Err(err).Msg("Unable to create request")
		return pageResponse{}, fmt.Errorf("Unable to create request: %w", err)
	}
	req.Header.Set("Notion-Version", "2021-08-16")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.SecretToken))
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		api.Logger.Err(err).Msg("Unable to retrieve response")
		return pageResponse{}, fmt.Errorf("Unable to retrieve response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		api.Logger.Err(err).Msg("Received invalid status")
		return pageResponse{}, fmt.Errorf("Received invalid status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		api.Logger.Err(err).Msg("Enable to read response body")
		return pageResponse{}, fmt.Errorf("Enable to read response body: %w", err)
	}
	api.Logger.Trace().RawJSON("page_response_json", body).Msg("Receieved Notion API response")

	var pageResponse pageResponse
	json.Unmarshal(body, &pageResponse)
	return pageResponse, nil
}
