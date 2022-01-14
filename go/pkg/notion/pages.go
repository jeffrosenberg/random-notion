package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jeffrosenberg/random-notion/pkg/logging"
	"github.com/rs/zerolog"
)

type Page struct {
	Id             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Url            string `json:"url"`
}

type PageGetter interface {
	GetPages() ([]Page, error)
	GetPagesSinceTime(time.Time) ([]Page, error)
	GetDatabaseId() string
}

// ===============================================================
// Notion query request body,
// per https://developers.notion.com/reference/post-database-query
// There are other properties, I've only defined what I need
// ---------------------------------------------------------------
type filterDateClause struct {
	After      string `json:"after,omitempty"`
	Before     string `json:"before,omitempty"`
	Equals     string `json:"equals,omitempty"`
	OnOrAfter  string `json:"on_or_after,omitempty"`
	OnOrBefore string `json:"on_or_before,omitempty"`
}

type filterDef struct {
	Property string           `json:"property"`
	Date     filterDateClause `json:"date"`
}

type pageRequest struct {
	Filter      filterDef `json:"filter,omitempty"`
	PageSize    uint8     `json:"page_size"`
	StartCursor string    `json:"start_cursor,omitempty"`
}

// ===============================================================

type pageResponse struct {
	Object  string `json:"object"`
	Results []Page `json:"results"`
	Next    string `json:"next_cursor"`
	HasMore bool   `json:"has_more"`
}

var logger *zerolog.Logger

// Return pages from the Notion API, filtered by time and starting at an optional cursor string
func (api *ApiConfig) getPages(sinceTime *time.Time, cursor string) ([]Page, error) {
	defer logging.LogFunction(
		"pages.getPages", time.Now(), "Getting pages from API",
		map[string]interface{}{
			"since_time": sinceTime,
			"cursor":     cursor,
		},
	)
	pages := []Page{}
	hasMore := true
	logger = logging.GetLogger()

	for hasMore == true {
		response, err := api.queryPages(sinceTime, cursor)
		if err != nil {
			logger.Err(err).Send()
			return nil, err // More robust error handling would be nice, but skipping as this is a hobby project
		}
		pages = append(pages, response.Results...)
		hasMore = response.HasMore
		cursor = response.Next
		logger.Debug().
			Int("pages_retrieved", len(response.Results)).
			Bool("has_more", response.HasMore).
			Str("cursor", response.Next).
			Msg("Processed Notion API response")
	}

	return pages, nil
}

// Return all pages from the Notion API
func (api *ApiConfig) GetPages() ([]Page, error) {
	return api.getPages(nil, "")
}

// Return pages from the Notion API, filtered by time
func (api *ApiConfig) GetPagesSinceTime(sinceTime time.Time) ([]Page, error) {
	if sinceTime.IsZero() {
		return api.GetPages()
	}
	return api.getPages(&sinceTime, "")
}

func (api *ApiConfig) GetDatabaseId() string {
	return api.DatabaseId
}

func (api *ApiConfig) queryPages(sinceTime *time.Time, cursor string) (pageResponse, error) {
	url, err := url.Parse(fmt.Sprintf("%s/databases/%s/query", api.Url, api.DatabaseId))
	if err != nil {
		return pageResponse{}, fmt.Errorf("Unable to parse URL: %w", err)
	}

	client := &http.Client{}
	postBody := pageRequest{
		PageSize:    api.PageSize,
		StartCursor: cursor,
	}
	if sinceTime != nil {
		postBody.Filter = filterDef{
			Property: "Created",
			Date: filterDateClause{
				After: sinceTime.Format(ISO_TIME),
			},
		}
	}
	jsonValue, _ := json.Marshal(postBody)
	logger.Trace().
		Str("request_verb", "POST").
		Str("request_url", url.String()).
		RawJSON("request_json", jsonValue).
		Msg("Prepared Notion API request")
	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(jsonValue))
	if err != nil {
		return pageResponse{}, fmt.Errorf("Unable to create request: %w", err)
	}
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
	logger.Trace().RawJSON("page_response_json", body).Msg("Receieved Notion API response")

	var pageResponse pageResponse
	json.Unmarshal(body, &pageResponse)
	return pageResponse, nil
}
