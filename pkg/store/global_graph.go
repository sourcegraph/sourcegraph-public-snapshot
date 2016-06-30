package store

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// GlobalRefs defines the interface for getting and listing global ref locations.
type GlobalRefs interface {
	// Get returns the names and ref counts of all repos and files within those repos
	// that refer the given def.
	Get(ctx context.Context, op *sourcegraph.DefsListRefLocationsOp) (*sourcegraph.RefLocationsList, error)

	// Update takes the graph output of a repo at the latest commit and
	// updates the set of refs in the global ref store that originate from
	// it.
	Update(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) error
}

// DefExamples defines the interface for getting and listing def usage examples.
type DefExamples interface {
	// Get returns the list of example locations for a given def.
	Get(ctx context.Context, op *sourcegraph.DefsListExamplesOp) (*sourcegraph.RefLocationsList, error)
}

type RepoUnit struct {
	Repo     int32
	Unit     string
	UnitType string
}

type GlobalDeps interface {
	Upsert(ctx context.Context, resolutions []*unit.Resolution) error
	Resolve(ctx context.Context, raw *unit.Key) ([]unit.Key, error)
}

type Defs interface {
	Search(ctx context.Context, op DefSearchOp) (*sourcegraph.SearchResultsList, error)
	UpdateFromSrclibStore(ctx context.Context, op DefUpdateOp) error
	Update(ctx context.Context, op DefUpdateOp) error
}

type DefSearchOp struct {
	// TokQuery is a list of tokens that describe the user's text
	// query. Order matter, as the last token is given especial weight.
	TokQuery []string
	Opt      *sourcegraph.SearchOptions
}

type DefUpdateOp struct {
	Repo     int32
	CommitID string
	Defs     []*graph.Def

	RefreshCounts bool

	// Latest is true if and only if the data imported in this update should be
	// treated as the latest version of the default branch (e.g., tip of master)
	Latest bool
}
