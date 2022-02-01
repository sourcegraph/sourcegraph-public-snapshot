package runner

import "github.com/sourcegraph/sourcegraph/internal/database/migration/definition"

func extractIDs(definitions []definition.Definition) []int {
	ids := make([]int, 0, len(definitions))
	for _, definition := range definitions {
		ids = append(ids, definition.ID)
	}

	return ids
}
