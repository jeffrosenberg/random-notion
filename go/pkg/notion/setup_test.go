package notion

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/rs/zerolog/log"
)

const mockDatabaseId = "99999999abcdefgh1234000000000000"
const mockApiToken = "secret_token"
const mockApiVersion = "2021-08-16"
const mockCursor = "240c0dcf-8334-43e5-9a01-a914c21de7e4"

func mockNotionServer(mockData string, status int) (*httptest.Server, *ApiConfig) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if notionHeaderIsValid(w, r) {
			w.WriteHeader(status)
			w.Write([]byte(mockData))
		}
	}))

	api := &ApiConfig{
		Url:         server.URL,
		DatabaseId:  mockDatabaseId,
		SecretToken: mockApiToken,
		Logger:      &log.Logger,
	}

	return server, api
}

func mockNotionServerWithPaging(mockData []string, status int) (*httptest.Server, *ApiConfig) {
	i := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if notionHeaderIsValid(w, r) {
			w.WriteHeader(status)
			if cursorValuePassed(r) {
				i++
			}
			w.Write([]byte(mockData[i]))
		}
	}))

	api := &ApiConfig{
		Url:         server.URL,
		DatabaseId:  mockDatabaseId,
		SecretToken: mockApiToken,
		Logger:      &log.Logger,
	}

	return server, api
}

func notionHeaderIsValid(w http.ResponseWriter, r *http.Request) bool {
	if contains(r.Header.Values("Authorization"), fmt.Sprintf("Bearer %s", mockApiToken)) == false {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"object": "error","status": 400,"code": "unauthorized","message": "API token is invalid."}`))
		return false
	} else if contains(r.Header.Values("Notion-Version"), mockApiVersion) == false {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"object": "error","status": 401,"code": "missing_version","message": "Notion-Version header should be defined..."}`))
		return false
	} else {
		return true
	}
}

func cursorValuePassed(r *http.Request) bool {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var pageRequest pageRequest
	json.Unmarshal(body, &pageRequest)
	return pageRequest.StartCursor == mockCursor
}

func contains(input []string, expected string) bool {
	for _, val := range input {
		if val == expected {
			return true
		}
	}
	return false
}
