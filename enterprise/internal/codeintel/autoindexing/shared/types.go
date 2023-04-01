package shared

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
)

//
// TODO - make invisible

type SourcedCommits struct {
	RepositoryID   int
	RepositoryName string
	Commits        []string
}

// IndexConfiguration stores the index configuration for a repository.
type IndexConfiguration struct {
	ID           int
	RepositoryID int
	Data         []byte
}

//
// TODO - move to uploads?

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
