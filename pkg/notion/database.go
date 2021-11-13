package notion

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/jeffrosenberg/random-notion/configs"
)

type Database struct {
	Id             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Url            string `json:"url"`
}

func GetDatabase(config *configs.NotionConfig) (*Database, error) {
	url, err := url.Parse(fmt.Sprintf("%s/databases/%s", config.ApiUrl, config.DatabaseId))
	if err != nil {
		return nil, fmt.Errorf("Unable to parse URL: %w", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url.String(), nil)
	req.Header.Set("Notion-Version", "2021-08-16")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SecretToken))

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
