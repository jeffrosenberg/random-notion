package notion

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetrieveDatabase(t *testing.T) {
	mockData := `{
    "object": "database",
    "id": "99999999-abcd-efgh-1234-000000000000",
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
		Id:             "99999999-abcd-efgh-1234-000000000000",
		CreatedTime:    "2021-02-27T04:04:00.000Z",
		LastEditedTime: "2021-11-10T00:58:00.000Z",
		Url:            "https://www.notion.so/45d3242e5c6d4a3bb99e4aa4db83f015",
	}

	ts, api := mockNotionServer(mockData, http.StatusOK)
	defer ts.Close()

	db, err := api.GetDatabase()
	if assert.NoError(t, err) {
		assert.NotNil(t, db)
		assert.EqualValues(t, expected, *db)
	}
}

func TestRetrieveDatabaseWithMissingUrl(t *testing.T) {
	mockData := `{
    "object": "database",
    "id": "99999999-abcd-efgh-1234-000000000000",
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
		Id:             "99999999-abcd-efgh-1234-000000000000",
		CreatedTime:    "2021-02-27T04:04:00.000Z",
		LastEditedTime: "2021-11-10T00:58:00.000Z",
		Url:            "",
	}

	ts, api := mockNotionServer(mockData, http.StatusOK)
	defer ts.Close()

	db, err := api.GetDatabase()
	if assert.NoError(t, err) {
		assert.NotNil(t, db)
		assert.EqualValues(t, expected, *db)
	}
}

func TestRetrieveDatabaseError(t *testing.T) {
	mockData := `{
		"object": "error",
		"status": 500,
		"code": "mock",
    "message": "Badly mocked data that should probably be refactored"
	}`

	ts, api := mockNotionServer(mockData, http.StatusInternalServerError)
	defer ts.Close()

	db, err := api.GetDatabase()
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestRetrieveDatabaseNotFound(t *testing.T) {
	mockData := `{
		"object": "error",
		"status": 404,
		"code": "object_not_found",
		"message": "Could not find database with ID: 99999999-abcd-efgh-1234-000000000000."
	}`

	ts, api := mockNotionServer(mockData, http.StatusNotFound)
	defer ts.Close()

	db, err := api.GetDatabase()
	assert.Error(t, err)
	assert.Nil(t, db)
}
