package randompage

import (
	"math/rand"
	"time"

	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

func GetRandomPage(getter notion.PageGetter) (*notion.Page, error) {
	pages, err := getter.GetPages()
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	randomPage := (*pages)[rand.Intn(len(*pages))]
	return &randomPage, nil
}
