package main

import (
	"fmt"
	"testing"

	"github.com/jeffrosenberg/random-notion/pkg/notion"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type TestApiConfig struct {
	mock.Mock
	pages []notion.Page
}

type TestSelector struct {
	mock.Mock
}

const mockDatabaseId = "99999999abcdefgh1234000000000000"

func (api *TestApiConfig) GetPages(cursor string) ([]notion.Page, error) {
	api.MethodCalled("GetPages", cursor)

	if api.pages == nil {
		return nil, fmt.Errorf("No pages found")
	}

	if cursor == "" {
		return api.pages, nil
	}

	return nil, fmt.Errorf("Test method not implemented")
}

func (api *TestApiConfig) GetAllPages() ([]notion.Page, error) {
	api.MethodCalled("GetAllPages")

	if api.pages == nil {
		return nil, fmt.Errorf("No pages found")
	}

	return api.pages, nil
}

func (api *TestApiConfig) GetLogger() *zerolog.Logger {
	api.MethodCalled("GetLogger")
	return &log.Logger
}

func (api *TestApiConfig) SetLogger(logger *zerolog.Logger) {
	return // no action for tests
}

func (selector *TestSelector) SelectPage(pages []notion.Page) *notion.Page {
	selector.MethodCalled("SelectPage")
	return &pages[0]
}

func TestHandleRequest_Success(t *testing.T) {
	api := &TestApiConfig{
		pages: []notion.Page{
			{
				Id:             "3350ba04-48b1-43e3-8726-1b1e9828b2b3",
				CreatedTime:    "2021-11-05T12:54:00.000Z",
				LastEditedTime: "2021-11-05T12:55:00.000Z",
				Url:            "https://www.notion.so/Initial-goals-3350ba0448b143e387261b1e9828b2b3",
			},
		},
	}
	selector := &TestSelector{}
	api.Mock.On("GetPages", "")    // Assert that PageGetter methods are called
	selector.Mock.On("SelectPage") // Assert that PageSelector methods are called

	result, err := execGetPage(api, selector)
	require.NoError(t, err)
	assert.EqualValues(t, api.pages[0].Url, result)
	api.AssertExpectations(t)
	selector.AssertExpectations(t)
}

func TestHandleRequest_Error(t *testing.T) {
	api := &TestApiConfig{}
	selector := &TestSelector{}
	api.Mock.On("GetPages", "") // Assert that PageGetter methods are called
	// selector.Mock.On("SelectPage") // PageSelector methods should NOT be called

	result, err := execGetPage(api, selector)
	require.Error(t, err)
	assert.EqualValues(t, "", result)
	api.AssertExpectations(t)
	selector.AssertExpectations(t)
}
