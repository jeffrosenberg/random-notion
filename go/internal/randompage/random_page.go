package randompage

import (
	"math/rand"
	"time"

	"github.com/jeffrosenberg/random-notion/configs"
	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

func GetRandomPage(config *configs.NotionConfig) (*notion.Page, error) {
	pages, err := notion.GetPages(config)
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	randomPage := (*pages)[rand.Intn(len(*pages))]
	return &randomPage, nil
}
