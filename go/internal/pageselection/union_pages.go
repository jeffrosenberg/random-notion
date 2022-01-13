package pageselection

import (
	"github.com/jeffrosenberg/random-notion/internal/persistence"
	"github.com/jeffrosenberg/random-notion/pkg/notion"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func UnionPages(dto *persistence.NotionDTO, addl []notion.Page, logger *zerolog.Logger) (pagesAdded bool) {
	if logger == nil {
		logger = &log.Logger
	}
	logger.Info().Str("function", "UnionPages").Msg("Unioning pages")
	logger.Debug().
		Int("pages_cached", len(dto.Pages)).
		Int("pages_api", len(addl)).
		Msg("Unioning pages")

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
