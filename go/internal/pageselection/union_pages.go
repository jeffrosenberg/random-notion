package pageselection

import (
	"time"

	"github.com/jeffrosenberg/random-notion/internal/persistence"
	"github.com/jeffrosenberg/random-notion/pkg/logging"
	"github.com/jeffrosenberg/random-notion/pkg/notion"
)

func UnionPages(dto *persistence.NotionDTO, addl []notion.Page) (pagesAdded bool) {
	defer logging.LogFunction(
		"pageselection.UnionPages", time.Now(), "Unioning pages",
		map[string]interface{}{
			"pages_cached": len(dto.Pages),
			"pages_api":    len(addl),
		},
	)

	// Short circuit if no addition pages found beyond those cached
	if len(addl) == 0 {
		return
	}

	// Store IDs in a map for deduping
	ids := make(map[string]struct{})
	for _, page := range dto.Pages {
		ids[page.Id] = struct{}{}
	}

	// Dedup incoming slice of pages
	pagesToAppend := make([]notion.Page, 0, len(addl))
	for _, page := range addl {
		_, exists := ids[page.Id]
		if !exists {
			pagesToAppend = append(pagesToAppend, page)
			pagesAdded = true
		}
	}

	// Update NotionPages DTO with deduped additional pages
	dto.Pages = append(dto.Pages, pagesToAppend...)
	return
}
