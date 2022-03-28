package result

import (
	"net/url"

	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type NotebookMatch struct {
	ID int64

	Name          string
	NamespaceName string
	Private       bool
	Stars         int
}

func (n NotebookMatch) RepoName() types.MinimalRepo {
	// This result type is not associated with any repository.
	return types.MinimalRepo{}
}

func (n NotebookMatch) Limit(limit int) int {
	// Always represents one result and limit > 0 so we just return limit - 1.
	return limit - 1
}

func (n *NotebookMatch) ResultCount() int {
	return 1
}

func (n *NotebookMatch) Select(path filter.SelectPath) Match {
	return nil
}

func (n *NotebookMatch) URL() *url.URL {
	return &url.URL{Path: "/notebooks/" + n.marshalNotebookID()}
}

func (n *NotebookMatch) Key() Key {
	return Key{
		ID: n.marshalNotebookID(),

		// TODO: Could represent date created, or maybe date updated, to improve relevance
		// AuthorDate: n.AuthorDate,

		// Use same rank as repos
		TypeRank: rankRepoMatch,

		// TODO: Key appears to be used for ranking results, maybe we should extend it to
		// include star count, or a more generic rank weight. This might be duplicative of
		// repo.Stars, but that appears only to be used in
		// (*searchIndexerServer).serveConfiguration so maybe that is okay?
		// WeightRank: n.Stars,
	}
}

func (n *NotebookMatch) searchResultMarker() {}

// from enterprise/cmd/frontend/internal/notebooks/resolvers/resolvers.go
func (n *NotebookMatch) marshalNotebookID() string {
	return string(relay.MarshalID("Notebook", 1))
}
