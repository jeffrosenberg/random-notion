package configs

const API_URI = "https://api.notion.com/v1"
const DEFAULT_PAGE_SIZE = uint8(10)

type NotionConfig struct {
	ApiUrl      string
	DatabaseId  string
	SecretToken string
	PageSize    uint8
}

func NewNotionConfig() *NotionConfig {
	return &NotionConfig{
		ApiUrl:      API_URI,
		DatabaseId:  "",
		SecretToken: "",
		PageSize:    DEFAULT_PAGE_SIZE,
	}
}
