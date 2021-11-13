package configs

const API_URI = "https://api.notion.com/v1"

type NotionConfig struct {
	ApiUrl      string
	DatabaseId  string
	SecretToken string
}

func NewNotionConfig() *NotionConfig {
	return &NotionConfig{
		ApiUrl:      API_URI,
		DatabaseId:  "",
		SecretToken: "",
	}
}
