package notion

const API_URI = "https://api.notion.com/v1"
const ISO_TIME = "2006-01-02T15:04:05-0700"
const DEFAULT_PAGE_SIZE = uint8(100)

type ApiConfig struct {
	Url         string
	DatabaseId  string
	SecretToken string
	PageSize    uint8
}

func NewApiConfig() *ApiConfig {
	return &ApiConfig{
		Url:         API_URI,
		DatabaseId:  "",
		SecretToken: "",
		PageSize:    DEFAULT_PAGE_SIZE,
	}
}
