package result

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type NotebookBlocks []NotebookBlock

func (blocks NotebookBlocks) Value() (driver.Value, error) {
	return json.Marshal(blocks)
}

func (blocks *NotebookBlocks) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &blocks)
}

type NotebookMatch struct {
	ID int64

	Title     string
	Namespace string
	Private   bool
	Stars     int

	Blocks NotebookBlocks `json:"-"`
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
	if path.Root() != filter.Notebook {
		return nil
	}

	switch len(path) {
	case 1:
		return n
	case 2, 3:
		if path[1] == "block" {
			if len(n.Blocks) == 0 {
				return nil
			}

			return (&NotebookBlocksMatch{
				Notebook: *n,
				Blocks:   n.Blocks,
			}).Select(path)
		}
	}
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
	return string(relay.MarshalID("Notebook", n.ID))
}
