package notion

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Database struct {
	Id             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Url            string `json:"url"`
}

func (api *ApiConfig) GetDatabase() (*Database, error) {
	api.Logger.Info().Msg("Getting database")
	url, err := url.Parse(fmt.Sprintf("%s/databases/%s", api.Url, api.DatabaseId))
	if err != nil {
		api.Logger.Err(err).Msg("Unable to parse URL")
		return nil, fmt.Errorf("Unable to parse URL: %w", err)
	}

	client := &http.Client{}
	api.Logger.Trace().
		Str("request_verb", "GET").
		Str("request_url", url.String()).
		Msg("Prepared Notion API request")
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		api.Logger.Err(err).Msg("Unable to create request")
		return nil, fmt.Errorf("Unable to create request: %w", err)
	}
	req.Header.Set("Notion-Version", "2021-08-16")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.SecretToken))

	res, err := client.Do(req)
	if err != nil {
		api.Logger.Err(err).Msg("Unable to retrieve response")
		return nil, fmt.Errorf("Unable to retrieve response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Received invalid status: %s", res.Status)
		api.Logger.Err(err).Msg("Received invalid status")
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		api.Logger.Err(err).Msg("Unable to read response body")
		return nil, fmt.Errorf("Unable to read response body: %w", err)
	}
	api.Logger.Trace().RawJSON("db_response_json", body).Msg("Receieved Notion API response")

	var db Database
	json.Unmarshal(body, &db)

	return &db, nil
}
