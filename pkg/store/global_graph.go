package store

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// GlobalDefs defines the interface for searching global defs.
type GlobalDefs interface {
	// Search performs a global search for defs that match the given repo, unit and def
	// query, and ranks the defs by a combination of bag of words similarity and global ref count.
	Search(ctx context.Context, op *GlobalDefSearchOp) (*sourcegraph.SearchResultsList, error)

	// Update takes the graph output of a list of source units and
	// updates the set of defs in the global def store that originate
	// from these source units. If a repository is specified in op
	// with an empty source unit, then Update updates the data for
	// the entire repository.
	Update(ctx context.Context, op GlobalDefUpdateOp) error

	// RefreshRefCounts computes and sets the global ref counts of all defs in the
	// specified source units. If a repository is specified in op with
	// an empty source unit, then RefreshRefCounts updates ref counts
	// for the entire repository.
	RefreshRefCounts(ctx context.Context, op GlobalDefUpdateOp) error
}

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
	Get(ctx context.Context, op *sourcegraph.DefsListExamplesOp) (*sourcegraph.ExamplesList, error)
}

type GlobalDefSearchOp struct {
	UnitQuery     string
	UnitTypeQuery string

	// TokQuery is a list of tokens that describe the user's text
	// query. Order matter, as the last token is given especial weight.
	TokQuery []string

	Opt *sourcegraph.SearchOptions
}

type GlobalDefUpdateOp struct {
	RepoUnits []RepoUnit
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
