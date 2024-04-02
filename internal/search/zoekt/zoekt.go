package zoekt

import (
	"github.com/sourcegraph/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

// repoRevFunc is a function which maps repository names returned from Zoekt
// into the Sourcegraph's resolved repository revisions for the search.
type repoRevFunc func(file *zoekt.FileMatch) (repo types.MinimalRepo, revs []string)
