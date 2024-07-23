package shared

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type AvailableIndexer struct {
	Roots   []string
	Indexer CodeIntelIndexer
}

func PopulateInferredAvailableIndexers(indexJobs []config.AutoIndexJobSpec, blocklist map[string]struct{}, inferredAvailableIndexers map[string]AvailableIndexer) map[string]AvailableIndexer {
	for _, job := range indexJobs {
		indexer := job.GetIndexerName()
		key := GetKeyForLookup(indexer, job.GetRoot())
		// Only add them to the inferred jobs map if they're not already in the recent uploads
		// blocklist. This is to avoid hinting at an available index if we've already indexed it.
		if _, ok := blocklist[key]; !ok {
			ai := inferredAvailableIndexers[key]
			ai.Roots = append(ai.Roots, job.GetRoot())
			if p, ok := PreferredIndexers[indexer]; ok {
				ai.Indexer = p
			}

			inferredAvailableIndexers[key] = ai
		}
	}

	return inferredAvailableIndexers
}

// GetKeyForLookup creates a quick unique key for a map lookup.
func GetKeyForLookup(indexer, root string) string {
	return fmt.Sprintf("%s:%s", sanitizeIndexer(indexer), root)
}

func sanitizeIndexer(indexer string) string {
	return strings.TrimPrefix(strings.Split(strings.Split(indexer, "@")[0], ":")[0], "sourcegraph/")
}
