package notion

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/jeffrosenberg/random-notion/configs"
)

const mockDatabaseId = "99999999abcdefgh1234000000000000"
const mockApiToken = "secret_token"
const mockApiVersion = "2021-08-16"

func mockNotionServer(mockData string, status int) (*httptest.Server, *configs.NotionConfig) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate Notion headers
		if contains(r.Header.Values("Authorization"), fmt.Sprintf("Bearer %s", mockApiToken)) == false {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"object": "error","status": 400,"code": "unauthorized","message": "API token is invalid."}`))
		} else if contains(r.Header.Values("Notion-Version"), mockApiVersion) == false {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"object": "error","status": 401,"code": "missing_version","message": "Notion-Version header should be defined..."}`))
		} else {
			w.WriteHeader(status)
			w.Write([]byte(mockData))
		}
	}))

	config := &configs.NotionConfig{
		ApiUrl:      server.URL,
		DatabaseId:  mockDatabaseId,
		SecretToken: mockApiToken,
	}

	return server, config
}

func contains(input []string, expected string) bool {
	for _, val := range input {
		if val == expected {
			return true
		}
	}
	return false
}
