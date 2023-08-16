package search

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type NewSearcher interface {
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
	NewSearch(ctx context.Context, q string) (SearchQuery, error)
}

// RepositoryRevSpec represents zero or more revisions we need to search in a
// repository for a revision specifier. This can be inferred relatively
// cheaply from parsing a query and the repos table.
//
// This type needs to be serializable so that we can persist it to a database
// or queue.
//
// Note: this is like a search/repos.RepoRevSpecs except we store 1 revision
// specifier per repository. It may be worth updating this to instead store a
// slice of RevisionSpecifiers.
type RepositoryRevSpec struct {
	// Repository is the repository to search.
	Repository api.RepoID

	// RevisionSpecifier is something that still needs to be resolved on gitserver. This
	// is a serialiazed version of query.RevisionSpecifier.
	RevisionSpecifier string
}

func (r RepositoryRevSpec) String() string {
	return fmt.Sprintf("RepositoryRevSpec{%d@%s}", r.Repository, r.RevisionSpecifier)
}

// RepositoryRevision represents the smallest unit we can search over, a
// specific repository and revision.
//
// This type needs to be serializable so that we can persist it to a database
// or queue.
type RepositoryRevision struct {
	// RepositoryRevSpec is where this RepositoryRevision got resolved from.
	RepositoryRevSpec

	// Revision is a resolved revision specifier. eg HEAD, branch-name,
	// commit-hash, etc.
	Revision string
}

func (r RepositoryRevision) String() string {
	return fmt.Sprintf("RepositoryRevision{%d@%s}", r.Repository, r.Revision)
}

// SearchQuery represents a search in a way we can break up the work. The flow is
// something like:
//
//  1. RepositoryRevSpecs -> just speak to the DB to find the list of repos we need to search.
//  2. ResolveRepositoryRevSpec -> speak to gitserver to find out which commits to search.
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
//
// Commentary on exhaustive worker jobs added in
// https://github.com/sourcegraph/sourcegraph/pull/55587:
//
//   - ExhaustiveSearchJob uses RepositoryRevSpecs to create ExhaustiveSearchRepoJob
//   - ExhaustiveSearchRepoJob uses ResolveRepositoryRevSpec to create ExhaustiveSearchRepoRevisionJob
//   - ExhaustiveSearchRepoRevisionJob uses Search
//
// In each case I imagine NewSearcher.NewSearch(query) to get hold of the
// SearchQuery. NewSearch is envisioned as being cheap to do. The only IO it
// does is maybe reading featureflags/site configuration/etc. This does mean
// it is possible for things to change over time, but this should be rare and
// will result in a well defined error. The alternative is a way to serialize
// a SearchQuery, but this makes it harder to make changes to search going
// forward for what should be rare errors.
type SearchQuery interface {
	RepositoryRevSpecs(context.Context) ([]RepositoryRevSpec, error)

	ResolveRepositoryRevSpec(context.Context, RepositoryRevSpec) ([]RepositoryRevision, error)

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

// NewSearcherFake is a convenient working implementation of SearchQuery which
// always will write results generated from the repoRevs. It expects a query
// string which looks like
//
//	 1@rev1 1@rev2 2@rev3
//
//	This is a space separated list of {repoid}@{revision}.
//
//	- RepositoryRevSpecs will return one RepositoryRevSpec per unique repository.
//	- ResolveRepositoryRevSpec returns the repoRevs for that repository.
//	- Search will write one result which is just the repo and revision.
func NewSearcherFake() NewSearcher {
	return backendFake{}
}

type backendFake struct{}

func (backendFake) NewSearch(ctx context.Context, q string) (SearchQuery, error) {
	var repoRevs []RepositoryRevision
	for _, part := range strings.Fields(q) {
		var r RepositoryRevision
		if n, err := fmt.Sscanf(part, "%d@%s", &r.Repository, &r.Revision); n != 2 || err != nil {
			return nil, errors.Errorf("failed to parse repository revision %q", part)
		}
		r.RepositoryRevSpec.Repository = r.Repository
		r.RepositoryRevSpec.RevisionSpecifier = "spec"
		repoRevs = append(repoRevs, r)
	}
	return searcherFake{repoRevs: repoRevs}, nil
}

type searcherFake struct {
	repoRevs []RepositoryRevision
}

func (s searcherFake) RepositoryRevSpecs(context.Context) ([]RepositoryRevSpec, error) {
	seen := map[RepositoryRevSpec]bool{}
	var repoRevSpecs []RepositoryRevSpec
	for _, r := range s.repoRevs {
		if seen[r.RepositoryRevSpec] {
			continue
		}
		seen[r.RepositoryRevSpec] = true
		repoRevSpecs = append(repoRevSpecs, r.RepositoryRevSpec)
	}
	return repoRevSpecs, nil
}

func (s searcherFake) ResolveRepositoryRevSpec(_ context.Context, repoRevSpec RepositoryRevSpec) ([]RepositoryRevision, error) {
	var repoRevs []RepositoryRevision
	for _, r := range s.repoRevs {
		if r.RepositoryRevSpec == repoRevSpec {
			repoRevs = append(repoRevs, r)
		}
	}
	return repoRevs, nil
}

func (s searcherFake) Search(_ context.Context, r RepositoryRevision, w CSVWriter) error {
	if err := w.WriteHeader("repo", "revspec", "revision"); err != nil {
		return err
	}
	return w.WriteRow(strconv.Itoa(int(r.Repository)), r.RevisionSpecifier, string(r.Revision))
}
