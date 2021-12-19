package notion

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const API_URI = "https://api.notion.com/v1"
const DEFAULT_PAGE_SIZE = uint8(10)

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
