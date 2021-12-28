package pageselection

import (
	"math/rand"
	"time"

	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

type RandomPage struct{} // Random page selection strategy

func (RandomPage) SelectPage(pages []notion.Page) *notion.Page {
	rand.Seed(time.Now().UnixNano())
	return &(pages[rand.Intn(len(pages))]) // Return a random index of pages
}
