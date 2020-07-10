package backend

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/schema"
)

// zoektIndexOptions are options which change what we index for a
// repository. Everytime a repository is indexed by zoekt this structure is
// fetched. See getIndexOptions in the zoekt codebase.
//
// We only specify a subset of the fields.
type zoektIndexOptions struct {
	// LargeFiles is a slice of glob patterns where matching file paths should
	// be indexed regardless of their size. The pattern syntax can be found
	// here: https://golang.org/pkg/path/filepath/#Match.
	LargeFiles []string

	// Symbols if true will make zoekt index the output of ctags.
	Symbols bool
}

// GetIndexOptions returns a json blob for consumption by
// sourcegraph-zoekt-indexserver.
func GetIndexOptions(c *schema.SiteConfiguration) ([]byte, error) {
	o := &zoektIndexOptions{
		LargeFiles: c.SearchLargeFiles,
		Symbols:    getBoolPtr(c.SearchIndexSymbolsEnabled, true),
	}
	return json.Marshal(o)
}

func getBoolPtr(b *bool, default_ bool) bool {
	if b == nil {
		return default_
	}
	return *b
}
