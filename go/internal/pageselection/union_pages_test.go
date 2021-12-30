package pageselection

import (
	"testing"

	"github.com/jeffrosenberg/random-notion/internal/persistence"
	"github.com/jeffrosenberg/random-notion/pkg/notion"
	"github.com/stretchr/testify/assert"
)

const mockDatabaseId = "99999999abcdefgh1234000000000000"
const mockPageId = "3350ba04-48b1-43e3-8726-1b1e9828b2b3"
const mockPageId2 string = "5331da24-6597-4f2d-a684-fd94a0f3278a"
const mockPageId3 string = "240c0dcf-8334-43e5-9a01-a914c21de7e4"
const mockPageUrl = "https://www.notion.so/Initial-goals-3350ba0448b143e387261b1e9828b2b3"
const mockPageUrl2 = "https://www.notion.so/Chicken-korma-recipe-How-to-make-chicken-korma-Swasthi-s-Recipes-5331da2465974f2da684fd94a0f3278a"
const mockPageUrl3 = "https://www.notion.so/Tampa-s-Best-Shuttle-Taxi-Service-Express-Transportation-240c0dcf833443e59a01a914c21de7e4"
const mockTime = "2021-11-05T12:54:00.000Z"

func TestNonOverlappingUnion(t *testing.T) {
	// Arrange
	input := persistence.NotionDTO{
		DatabaseId: mockDatabaseId,
		Pages: []notion.Page{
			{
				Id:             mockPageId,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl,
			},
		},
		NextCursor: mockPageId2,
	}
	append := []notion.Page{
		{
			Id:             mockPageId2,
			CreatedTime:    mockTime,
			LastEditedTime: mockTime,
			Url:            mockPageUrl2,
		},
	}
	expected := persistence.NotionDTO{
		DatabaseId: mockDatabaseId,
		Pages: []notion.Page{
			{
				Id:             mockPageId,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl,
			},
			{
				Id:             mockPageId2,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl2,
			},
		},
		NextCursor: mockPageId2,
	}

	// Act
	pagesAdded := UnionPages(&input, append)

	// Assert
	assert.Equal(t, true, pagesAdded)
	assert.Equal(t, expected, input)
}

func TestOverlappingUnion(t *testing.T) {
	// Arrange
	input := persistence.NotionDTO{
		DatabaseId: mockDatabaseId,
		Pages: []notion.Page{
			{
				Id:             mockPageId,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl,
			},
			{
				Id:             mockPageId2,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl2,
			},
		},
		NextCursor: mockPageId2,
	}
	append := []notion.Page{
		{
			Id:             mockPageId2,
			CreatedTime:    mockTime,
			LastEditedTime: mockTime,
			Url:            mockPageUrl2,
		},
		{
			Id:             mockPageId3,
			CreatedTime:    mockTime,
			LastEditedTime: mockTime,
			Url:            mockPageUrl3,
		},
	}
	expected := persistence.NotionDTO{
		DatabaseId: mockDatabaseId,
		Pages: []notion.Page{
			{
				Id:             mockPageId,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl,
			},
			{
				Id:             mockPageId2,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl2,
			},
			{
				Id:             mockPageId3,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl3,
			},
		},
		NextCursor: mockPageId3,
	}

	// Act
	pagesAdded := UnionPages(&input, append)

	// Assert
	assert.Equal(t, true, pagesAdded)
	assert.Equal(t, expected, input)
}

func TestNoPagesAdded(t *testing.T) {
	// Arrange
	input := persistence.NotionDTO{
		DatabaseId: mockDatabaseId,
		Pages: []notion.Page{
			{
				Id:             mockPageId,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl,
			},
		},
		NextCursor: mockPageId,
	}
	append := []notion.Page{}
	expected := persistence.NotionDTO{
		DatabaseId: mockDatabaseId,
		Pages: []notion.Page{
			{
				Id:             mockPageId,
				CreatedTime:    mockTime,
				LastEditedTime: mockTime,
				Url:            mockPageUrl,
			},
		},
		NextCursor: mockPageId,
	}

	// Act
	pagesAdded := UnionPages(&input, append)

	// Assert
	assert.Equal(t, false, pagesAdded)
	assert.Equal(t, expected, input)
}
