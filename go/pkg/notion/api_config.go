package notion

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const API_URI = "https://api.notion.com/v1"
const ISO_TIME = "2006-01-02T15:04:05-0700"
const DEFAULT_PAGE_SIZE = uint8(100)

type ApiConfig struct {
	Url         string
	DatabaseId  string
	SecretToken string
	PageSize    uint8
	Logger      *zerolog.Logger
}

type Logger interface {
	GetLogger() *zerolog.Logger
	SetLogger(*zerolog.Logger)
}

func (api *ApiConfig) GetLogger() *zerolog.Logger {
	return api.Logger
}

func (api *ApiConfig) SetLogger(logger *zerolog.Logger) {
	api.Logger = logger
}

func NewApiConfig() *ApiConfig {
	return &ApiConfig{
		Url:         API_URI,
		DatabaseId:  "",
		SecretToken: "",
		PageSize:    DEFAULT_PAGE_SIZE,
		Logger:      &log.Logger,
	}
}
