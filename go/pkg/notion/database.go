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
	url, err := url.Parse(fmt.Sprintf("%s/databases/%s", api.Url, api.DatabaseId))
	if err != nil {
		return nil, fmt.Errorf("Unable to parse URL: %w", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url.String(), nil)
	req.Header.Set("Notion-Version", "2021-08-16")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.SecretToken))

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

	var db Database
	json.Unmarshal(body, &db)

	return &db, nil
}
