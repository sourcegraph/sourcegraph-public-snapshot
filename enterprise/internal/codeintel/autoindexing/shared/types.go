package shared

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
)

type GetIndexesOptions struct {
	RepositoryID  int
	State         string
	States        []string
	Term          string
	WithoutUpload bool
	Limit         int
	Offset        int
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
	States        []string
	Term          string
	RepositoryID  int
	WithoutUpload bool
}

type ReindexIndexesOptions struct {
	States        []string
	Term          string
	RepositoryID  int
	WithoutUpload bool
}
