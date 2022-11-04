package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
)

type GetIndexesOptions struct {
	RepositoryID int
	State        string
	Term         string
	Limit        int
	Offset       int
}

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

type DeleteIndexesOptions struct {
	State        string
	Term         string
	RepositoryID int
}

type ReindexIndexesOptions struct {
	State        string
	Term         string
	RepositoryID int
}
