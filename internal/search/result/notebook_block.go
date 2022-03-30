package result

import (
	"fmt"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// TODO these types are all in enterprise, a bit much to copy-paste, so we keep it
// untyped for ease of use for now.
type NotebookBlock map[string]interface{}

type NotebookBlocksMatch struct {
	Notebook NotebookMatch

	Blocks []NotebookBlock
}

func (n NotebookBlocksMatch) RepoName() types.MinimalRepo {
	// This result type is not associated with any repository.
	return types.MinimalRepo{}
}

func (n NotebookBlocksMatch) Limit(limit int) int {
	return limit - len(n.Blocks)
}

func (n *NotebookBlocksMatch) ResultCount() int {
	return len(n.Blocks)
}

func (n *NotebookBlocksMatch) Select(path filter.SelectPath) Match {
	if path.Root() != filter.Notebook {
		return nil
	}
	switch len(path) {
	case 2:
		if path[1] == "block" {
			return n
		}
	case 3:
		blockType := path[2]
		var blocks NotebookBlocks
		for _, b := range n.Blocks {
			if b["type"] == blockType {
				blocks = append(blocks, b)
			}
		}
		return &NotebookBlocksMatch{
			Notebook: n.Notebook,
			Blocks:   blocks,
		}
	}

	return nil
}

func (n *NotebookBlocksMatch) URL() *url.URL {
	// Cannot link directly to blocks yet
	return &url.URL{Path: "/notebooks/" + n.Notebook.marshalNotebookID()}
}

func (n *NotebookBlocksMatch) Key() Key {
	return Key{
		ID: fmt.Sprintf("%s-blocks", n.Notebook.marshalNotebookID()),
		// Use same rank as files
		TypeRank: rankFileMatch,
	}
}

func (n *NotebookBlocksMatch) searchResultMarker() {}
