package notion

const API_URI = "https://api.notion.com/v1"
const DEFAULT_PAGE_SIZE = uint8(10)

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
