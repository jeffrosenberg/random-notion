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
	page *notion.Page
}

type TestSelector struct {
	mock.Mock
}

func (api *TestApiConfig) GetPages() ([]notion.Page, error) {
	api.MethodCalled("GetPages")

	if api.page == nil {
		return nil, fmt.Errorf("No pages found")
	}

	return []notion.Page{
		*api.page,
	}, nil
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
		page: &notion.Page{
			Id:             "3350ba04-48b1-43e3-8726-1b1e9828b2b3",
			CreatedTime:    "2021-11-05T12:54:00.000Z",
			LastEditedTime: "2021-11-05T12:55:00.000Z",
			Url:            "https://www.notion.so/Initial-goals-3350ba0448b143e387261b1e9828b2b3",
		},
	}
	selector := &TestSelector{}
	api.Mock.On("GetPages")        // Assert that PageGetter methods are called
	selector.Mock.On("SelectPage") // Assert that PageSelector methods are called

	result, err := execGetPage(api, selector)
	require.NoError(t, err)
	assert.EqualValues(t, api.page.Url, result)
	api.AssertExpectations(t)
	selector.AssertExpectations(t)
}

func TestHandleRequest_Error(t *testing.T) {
	api := &TestApiConfig{}
	selector := &TestSelector{}
	api.Mock.On("GetPages") // Assert that PageGetter methods are called
	// selector.Mock.On("SelectPage") // PageSelector methods should NOT be called

	result, err := execGetPage(api, selector)
	require.Error(t, err)
	assert.EqualValues(t, "", result)
	api.AssertExpectations(t)
	selector.AssertExpectations(t)
}
