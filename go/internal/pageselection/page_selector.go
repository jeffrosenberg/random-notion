package pageselection

import (
	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

type PageSelector interface {
	SelectPage([]notion.Page) *notion.Page
}
