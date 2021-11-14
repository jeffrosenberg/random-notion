package notion

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
								"database_id": "99999999-abcd-efgh-1234-000000000000"
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
								"database_id": "99999999-abcd-efgh-1234-000000000000"
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

	ts, config := mockNotionServer(mockData, http.StatusOK)
	defer ts.Close()

	pages, err := GetPages(config)
	if assert.NoError(t, err) {
		assert.NotNil(t, pages)
		assert.EqualValues(t, expected, *pages)
	}
}

func TestRetrievePagesEmpty(t *testing.T) {
	mockData := `{
		"object": "list",
		"results": [],
		"next_cursor": null,
		"has_more": false
	}`

	expected := []Page{}

	ts, config := mockNotionServer(mockData, http.StatusOK)
	defer ts.Close()

	pages, err := GetPages(config)
	if assert.NoError(t, err) {
		assert.NotNil(t, pages)
		assert.EqualValues(t, expected, *pages)
	}
}

func TestRetrievePagesError(t *testing.T) {
	mockData := `{
		"object": "error",
		"status": 500,
		"code": "mock",
    "message": "Badly mocked data that should probably be refactored"
	}`

	ts, config := mockNotionServer(mockData, http.StatusInternalServerError)
	defer ts.Close()

	pages, err := GetPages(config)
	assert.Error(t, err)
	assert.Nil(t, pages)
}

func TestRetrievePagesDatabaseNotFound(t *testing.T) {
	mockData := `{
		"object": "error",
		"status": 404,
		"code": "object_not_found",
		"message": "Could not find database with ID: 99999999-abcd-efgh-1234-000000000000."
	}`

	ts, config := mockNotionServer(mockData, http.StatusNotFound)
	defer ts.Close()

	pages, err := GetPages(config)
	assert.Error(t, err) // TODO: More specific error assertions
	assert.Nil(t, pages)
}

func TestRetrievePageCollection(t *testing.T) {
	mockData1 := `{
		"object": "list",
		"results": [
				{
						"object": "page",
						"id": "3350ba04-48b1-43e3-8726-1b1e9828b2b3",
						"created_time": "2021-11-05T12:54:00.000Z",
						"last_edited_time": "2021-11-05T12:55:00.000Z",
						"parent": {
								"type": "database_id",
								"database_id": "99999999-abcd-efgh-1234-000000000000"
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
								"database_id": "99999999-abcd-efgh-1234-000000000000"
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
		"has_more": true
	}`

	mockData2 := `{
		"object": "list",
		"results": [
				{
					"object": "page",
					"id": "240c0dcf-8334-43e5-9a01-a914c21de7e4",
					"created_time": "2021-10-12T23:51:00.000Z",
					"last_edited_time": "2021-10-25T13:47:00.000Z",
					"parent": {
							"type": "database_id",
							"database_id": "45d3242e-5c6d-4a3b-b99e-4aa4db83f015"
					},
					"archived": false,
					"properties": {
							"Created": {
									"id": "MEdb",
									"type": "created_time",
									"created_time": "2021-10-12T23:51:00.000Z"
							}
					},
					"url": "https://www.notion.so/Tampa-s-Best-Shuttle-Taxi-Service-Express-Transportation-240c0dcf833443e59a01a914c21de7e4"
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
		{
			Id:             "240c0dcf-8334-43e5-9a01-a914c21de7e4",
			CreatedTime:    "2021-10-12T23:51:00.000Z",
			LastEditedTime: "2021-10-25T13:47:00.000Z",
			Url:            "https://www.notion.so/Tampa-s-Best-Shuttle-Taxi-Service-Express-Transportation-240c0dcf833443e59a01a914c21de7e4",
		},
	}

	ts, config := mockNotionServerWithPaging([]string{mockData1, mockData2}, http.StatusOK)
	defer ts.Close()

	pages, err := GetPages(config)
	if assert.NoError(t, err) {
		assert.NotNil(t, pages)
		assert.EqualValues(t, expected, *pages)
	}
}
