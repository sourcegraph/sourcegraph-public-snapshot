package shared

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
)

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

// IndexConfiguration stores the index configuration for a repository.
type IndexConfiguration struct {
	ID           int    `json:"id"`
	RepositoryID int    `json:"repository_id"`
	Data         []byte `json:"data"`
}

type IndexesWithRepositoryNamespace struct {
	Root    string
	Indexer string
	Indexes []types.Index
}

type RepositoryWithCount struct {
	RepositoryID int
	Count        int
}

type RepositoryWithAvailableIndexers struct {
	RepositoryID      int
	AvailableIndexers map[string]AvailableIndexer
}
