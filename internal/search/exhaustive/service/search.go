package service

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	types2 "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

type NewSearcher interface {
	// NewSearch parses and minimally resolves the search query q. The
	// expectation is that this method is always fast and is deterministic, such
	// that calling this again in the future should return the same Searcher. IE
	// it can speak to the DB, but maybe not gitserver.
	//
	// userID is explicitly passed in and must match the actor for ctx. This
	// is done to prevent accidental bugs where we do a search on behalf of a
	// user as an internal user/etc.
	//
	// I expect this to be roughly equivalent to creation of a search plan in
	// our search codes job creator.
	//
	// Note: I expect things like feature flags for the user behind ctx could
	// affect what is returned. Alternatively as we release new versions of
	// Sourcegraph what is returned could change. This means we are not exactly
	// safe across repeated calls.
	NewSearch(ctx context.Context, userID int32, q string) (SearchQuery, error)
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
	RepositoryRevSpecs(context.Context) *iterator.Iterator[types.RepositoryRevSpecs]

	ResolveRepositoryRevSpec(context.Context, types.RepositoryRevSpecs) ([]types.RepositoryRevision, error)

	Search(context.Context, types.RepositoryRevision, MatchWriter) error
}

type MatchWriter interface {
	Write(match result.Match) error
}

// CSVWriter makes it, so we can avoid caring about search types and leave it
// up to the search job to decide the shape of data.
type CSVWriter interface {
	// WriteHeader should be called first and only once.
	WriteHeader(...string) error

	// WriteRow should have the same number of values as WriteHeader and can be
	// called zero or more times.
	WriteRow(...string) error

	io.Closer
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
	return newSearcherFunc(fakeNewSearch)
}

type newSearcherFunc func(context.Context, int32, string) (SearchQuery, error)

func (f newSearcherFunc) NewSearch(ctx context.Context, userID int32, q string) (SearchQuery, error) {
	return f(ctx, userID, q)
}

func fakeNewSearch(ctx context.Context, userID int32, q string) (SearchQuery, error) {
	if err := isSameUser(ctx, userID); err != nil {
		return nil, err
	}

	var repoRevs []types.RepositoryRevision
	for _, part := range strings.Fields(q) {
		var r types.RepositoryRevision
		if n, err := fmt.Sscanf(part, "%d@%s", &r.Repository, &r.Revision); n != 2 || err != nil {
			continue
		}
		r.RepositoryRevSpecs.Repository = r.Repository
		r.RepositoryRevSpecs.RevisionSpecifiers = types.RevisionSpecifiers("spec")
		repoRevs = append(repoRevs, r)
	}
	if len(repoRevs) == 0 {
		return nil, errors.Errorf("no repository revisions found in %q", q)
	}
	return searcherFake{
		userID:   userID,
		repoRevs: repoRevs,
	}, nil
}

type searcherFake struct {
	userID   int32
	repoRevs []types.RepositoryRevision
}

func (s searcherFake) RepositoryRevSpecs(ctx context.Context) *iterator.Iterator[types.RepositoryRevSpecs] {
	if err := isSameUser(ctx, s.userID); err != nil {
		iterator.New(func() ([]types.RepositoryRevSpecs, error) {
			return nil, err
		})
	}

	seen := map[types.RepositoryRevSpecs]bool{}
	var repoRevSpecs []types.RepositoryRevSpecs
	for _, r := range s.repoRevs {
		if seen[r.RepositoryRevSpecs] {
			continue
		}
		seen[r.RepositoryRevSpecs] = true
		repoRevSpecs = append(repoRevSpecs, r.RepositoryRevSpecs)
	}
	return iterator.From(repoRevSpecs)
}

func (s searcherFake) ResolveRepositoryRevSpec(ctx context.Context, repoRevSpec types.RepositoryRevSpecs) ([]types.RepositoryRevision, error) {
	if err := isSameUser(ctx, s.userID); err != nil {
		return nil, err
	}

	var repoRevs []types.RepositoryRevision
	for _, r := range s.repoRevs {
		if r.RepositoryRevSpecs == repoRevSpec {
			repoRevs = append(repoRevs, r)
		}
	}
	return repoRevs, nil
}

func (s searcherFake) Search(ctx context.Context, r types.RepositoryRevision, w MatchWriter) error {
	if err := isSameUser(ctx, s.userID); err != nil {
		return err
	}

	return w.Write(&result.FileMatch{
		File: result.File{
			Repo:     types2.MinimalRepo{ID: r.Repository, Name: "repo" + api.RepoName(strconv.Itoa(int(r.Repository)))},
			CommitID: api.CommitID(r.Revision),
			Path:     "path/to/file.go",
		},
	})
}

func isSameUser(ctx context.Context, userID int32) error {
	if userID == 0 {
		return errors.New("exhaustive search must be done on behalf of an authenticated user")
	}
	a := actor.FromContext(ctx)
	if a == nil || a.UID != userID {
		return errors.Errorf("exhaustive search must be run as user %d", userID)
	}
	return nil
}
