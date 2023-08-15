package search

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type Backend interface {
	// NewSearch parses and minimally resolves the search query q. The
	// expectation is that this method is always fast and is deterministic, such
	// that calling this again in the future should return the same Searcher. IE
	// it can speak to the DB, but maybe not gitserver.
	//
	// I expect this to be roughly equivalent to creation of a search plan in
	// our search codes job creator.
	//
	// Note: I expect things like feature flags for the user behind ctx could
	// affect what is returned. Alternatively as we release new versions of
	// Sourcegraph what is returned could change. This means we are not exactly
	// safe across repeated calls.
	NewSearch(ctx context.Context, q string) (Searcher, error)
}

// RepositoryRefSpec represents zero or more revisions we need to search in a
// repository for a refspec. This can be inferred relatively cheaply from
// parsing a query and the repos table.
//
// This type needs to be serializable so that we can persist it to a database
// or queue.
type RepositoryRefSpec struct {
	// Repository is the repository to search.
	Repository api.RepoID

	// RefSpec is something that still needs to be resolved on gitserver.
	RefSpec string
}

// RepositoryRevision represents the smallest unit we can search over, a
// specific repository and revision.
//
// This type needs to be serializable so that we can persist it to a database
// or queue.
type RepositoryRevision struct {
	// RepositoryRefSpec is where this RepositoryRevision got resolved from.
	RepositoryRefSpec

	// Revision is the resolved revision.
	Revision api.CommitID
}

// Searcher represents a search in a way we can break up the work. The flow is
// something like:
//
//  1. RepositoryRefSpecs -> just speak to the DB to find the list of repos we need to search.
//  2. ResolveRepositoryRefSpec -> speak to gitserver to find out which commits to search.
//  3. Search -> actually do a search.
//
// This does mean that things like searching a commit in a monorepo are
// expected to run over a reasonable time frame (eg within a minute?).
//
// For example doing a diff search in an old repo may not be fast enough, but
// I'm not sure if we should design that in?
//
// We expect each step can be retried, but with the expectation it isn't
// idempotent due to backend state changing. The main purpose of breaking it
// out like this is so we can report progress, do retries, and spread out the
// work over time.
type Searcher interface {
	RepositoryRefSpecs(context.Context) ([]RepositoryRefSpec, error)

	ResolveRepositoryRefSpec(context.Context, RepositoryRefSpec) ([]RepositoryRevision, error)

	Search(context.Context, RepositoryRevision, CSVWriter) error
}

// CSVWriter makes it so we can avoid caring about search types and leave it
// up to the search job to decide the shape of data.
//
// Note: I expect the implementation of this to handle things like chunking up
// the CSV/etc. EG once we hit 100MB of data it can write the data out then
// start a new file. It takes care of remembering the header for the new file.
type CSVWriter interface {
	// WriteHeader should be called first and only once.
	WriteHeader(...string) error

	// WriteRow should have the same number of values as WriteHeader and can be
	// called zero or more times.
	WriteRow(...string) error
}
