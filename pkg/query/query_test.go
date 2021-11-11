package query

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func initMockServer(mockData string, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(mockData))
	}))
}

func TestRetrieveDatabase(t *testing.T) {
	mockData := `{
    "object": "database",
    "id": "45d3242e-5c6d-4a3b-b99e-4aa4db83f015",
    "cover": null,
    "icon": {
        "type": "emoji",
        "emoji": "📚"
    },
    "created_time": "2021-02-27T04:04:00.000Z",
    "last_edited_time": "2021-11-10T00:58:00.000Z",
    "title": [
        {
            "type": "text",
            "text": {
                "content": "Content",
                "link": null
            },
            "annotations": {
                "bold": false,
                "italic": false,
                "strikethrough": false,
                "underline": false,
                "code": false,
                "color": "default"
            },
            "plain_text": "Content",
            "href": null
        }
    ],
    "properties": {
        "Created": {
            "id": "MEdb",
            "name": "Created",
            "type": "created_time",
            "created_time": {}
        },
        "Sort Order": {
            "id": "v:yF",
            "name": "Sort Order",
            "type": "formula",
            "formula": {
                "expression": "empty(prop(\"Tag Sort Order\")) ? 100 : prop(\"Tag Sort Order\")"
            }
        },
        "Name": {
            "id": "title",
            "name": "Name",
            "type": "title",
            "title": {}
        },
        "URL": {
            "id": "80d87f08-e3b5-41cd-9166-abd3604ea26f",
            "name": "URL",
            "type": "url",
            "url": {}
        }
    },
    "parent": {
        "type": "workspace",
        "workspace": true
    },
    "url": "https://www.notion.so/45d3242e5c6d4a3bb99e4aa4db83f015"
	}`

	expected := Database{
		Id:             "45d3242e-5c6d-4a3b-b99e-4aa4db83f015",
		CreatedTime:    "2021-02-27T04:04:00.000Z",
		LastEditedTime: "2021-11-10T00:58:00.000Z",
		Url:            "https://www.notion.so/45d3242e5c6d4a3bb99e4aa4db83f015",
	}

	ts := initMockServer(mockData, http.StatusOK)
	defer ts.Close()

	db, err := GetDatabase(ts.URL)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, *db)
}

func TestRetrieveDatabaseWithMissingUrl(t *testing.T) {
	mockData := `{
    "object": "database",
    "id": "45d3242e-5c6d-4a3b-b99e-4aa4db83f015",
    "cover": null,
    "icon": {
        "type": "emoji",
        "emoji": "📚"
    },
    "created_time": "2021-02-27T04:04:00.000Z",
    "last_edited_time": "2021-11-10T00:58:00.000Z",
    "title": [
        {
            "type": "text",
            "text": {
                "content": "Content",
                "link": null
            },
            "annotations": {
                "bold": false,
                "italic": false,
                "strikethrough": false,
                "underline": false,
                "code": false,
                "color": "default"
            },
            "plain_text": "Content",
            "href": null
        }
    ],
    "properties": {
        "Created": {
            "id": "MEdb",
            "name": "Created",
            "type": "created_time",
            "created_time": {}
        }
    },
    "parent": {
        "type": "workspace",
        "workspace": true
    }
	}`

	expected := Database{
		Id:             "45d3242e-5c6d-4a3b-b99e-4aa4db83f015",
		CreatedTime:    "2021-02-27T04:04:00.000Z",
		LastEditedTime: "2021-11-10T00:58:00.000Z",
		Url:            "",
	}

	ts := initMockServer(mockData, http.StatusOK)
	defer ts.Close()

	db, err := GetDatabase(ts.URL)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, *db)
}

func TestRetrieveDatabaseError(t *testing.T) {
	mockData := `{
		"object": "error",
		"status": 500,
		"code": "mock",
    "message": "Badly mocked data that should probably be refactored"
	}`

	ts := initMockServer(mockData, http.StatusInternalServerError)
	defer ts.Close()

	db, err := GetDatabase(ts.URL)
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestRetrieveDatabaseNotFound(t *testing.T) {
	mockData := `{
		"object": "error",
		"status": 404,
		"code": "object_not_found",
		"message": "Could not find database with ID: 8d4c8be5-668e-4cab-bdcd-9d4c3add9999."
	}`

	ts := initMockServer(mockData, http.StatusNotFound)
	defer ts.Close()

	db, err := GetDatabase(ts.URL)
	assert.Error(t, err) // TODO: More specific error assertions
	assert.Nil(t, db)
}

func TestRetrievePages(t *testing.T) {
	mockData := `{
		"object": "list",
		"results": [
				{
						"object": "page",
						"id": "3350ba04-48b1-43e3-8726-1b1e9828b2b3",
						"created_time": "2021-11-05T12:54:00.000Z",
						"last_edited_time": "2021-11-05T12:55:00.000Z",
						"parent": {
								"type": "database_id",
								"database_id": "45d3242e-5c6d-4a3b-b99e-4aa4db83f015"
						},
						"archived": false,
						"properties": {
								"Created": {
										"id": "MEdb",
										"type": "created_time",
										"created_time": "2021-11-05T12:54:00.000Z"
								}
						},
						"url": "https://www.notion.so/Initial-goals-3350ba0448b143e387261b1e9828b2b3"
				},
				{
						"object": "page",
						"id": "5331da24-6597-4f2d-a684-fd94a0f3278a",
						"created_time": "2021-11-01T01:01:00.000Z",
						"last_edited_time": "2021-11-01T13:24:00.000Z",
						"parent": {
								"type": "database_id",
								"database_id": "45d3242e-5c6d-4a3b-b99e-4aa4db83f015"
						},
						"archived": false,
						"properties": {
								"Created": {
										"id": "MEdb",
										"type": "created_time",
										"created_time": "2021-11-01T01:01:00.000Z"
								}
						},
						"url": "https://www.notion.so/Chicken-korma-recipe-How-to-make-chicken-korma-Swasthi-s-Recipes-5331da2465974f2da684fd94a0f3278a"
				}
		],
		"next_cursor": null,
		"has_more": false
	}`

	expected := []Page{
		{
			Id:             "3350ba04-48b1-43e3-8726-1b1e9828b2b3",
			CreatedTime:    "2021-11-05T12:54:00.000Z",
			LastEditedTime: "2021-11-05T12:55:00.000Z",
			Url:            "https://www.notion.so/Initial-goals-3350ba0448b143e387261b1e9828b2b3",
		},
		{
			Id:             "5331da24-6597-4f2d-a684-fd94a0f3278a",
			CreatedTime:    "2021-11-01T01:01:00.000Z",
			LastEditedTime: "2021-11-01T13:24:00.000Z",
			Url:            "https://www.notion.so/Chicken-korma-recipe-How-to-make-chicken-korma-Swasthi-s-Recipes-5331da2465974f2da684fd94a0f3278a",
		},
	}

	ts := initMockServer(mockData, http.StatusOK)
	defer ts.Close()

	pages, err := GetPages(ts.URL)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, *pages)
}

func TestRetrievePagesEmpty(t *testing.T) {
	mockData := `{
		"object": "list",
		"results": [],
		"next_cursor": null,
		"has_more": false
	}`

	expected := []Page{}

	ts := initMockServer(mockData, http.StatusOK)
	defer ts.Close()

	pages, err := GetPages(ts.URL)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, *pages)
}

func TestRetrievePagesError(t *testing.T) {
	mockData := `{
		"object": "error",
		"status": 500,
		"code": "mock",
    "message": "Badly mocked data that should probably be refactored"
	}`

	ts := initMockServer(mockData, http.StatusInternalServerError)
	defer ts.Close()

	pages, err := GetPages(ts.URL)
	assert.Error(t, err)
	assert.Nil(t, pages)
}

func TestRetrievePagesDatabaseNotFound(t *testing.T) {
	mockData := `{
		"object": "error",
		"status": 404,
		"code": "object_not_found",
		"message": "Could not find database with ID: 8d4c8be5-668e-4cab-bdcd-9d4c3add9999."
	}`

	ts := initMockServer(mockData, http.StatusNotFound)
	defer ts.Close()

	pages, err := GetPages(ts.URL)
	assert.Error(t, err) // TODO: More specific error assertions
	assert.Nil(t, pages)
}
