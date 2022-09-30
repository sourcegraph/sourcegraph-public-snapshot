package sharedresolvers

import "github.com/sourcegraph/sourcegraph/internal/codeintel/types"

type IndexesWithRepositoryNamespace struct {
	Root    string
	Indexer string
	Indexes []types.Index
}
