package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type AvailableIndexer struct {
	Roots   []string
	Indexer types.CodeIntelIndexer
}

type JobsOrHints interface {
	config.IndexJob | config.IndexJobHint
	GetIndexerName() string
	GetRoot() string
}

func PopulateInferredAvailableIndexers[J JobsOrHints](jobsOrHints []J, blocklist map[string]struct{}, inferredAvailableIndexers map[string]AvailableIndexer) map[string]AvailableIndexer {
	for _, job := range jobsOrHints {
		indexer := job.GetIndexerName()
		key := GetKeyForLookup(indexer, job.GetRoot())
		// Only add them to the inferred jobs map if they're not already in the recent uploads
		// blocklist. This is to avoid hinting at an available index if we've already indexed it.
		if _, ok := blocklist[key]; !ok {
			ai := inferredAvailableIndexers[key]
			ai.Roots = append(ai.Roots, job.GetRoot())
			if p, ok := types.PreferredIndexers[indexer]; ok {
				ai.Indexer = p
			}

			inferredAvailableIndexers[indexer] = ai
		}
	}

	return inferredAvailableIndexers
}

// GetKeyForLookup creates a quick unique key for a map lookup.
func GetKeyForLookup(indexer, root string) string {
	return fmt.Sprintf("%s:%s", indexer, root)
}
