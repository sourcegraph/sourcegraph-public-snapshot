package janitor

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	lru "github.com/hashicorp/golang-lru"
)

// repositoryExpirationState is the state of the expiration process running over a single repository.
// This includes several fields to help make the expiration scan of a repository's uploads more efficient
// (what commits we can avoid re-scanning due to the result of previous passes) as well as a small cache
// used for gitserver responses, which are often asked for multiple times when processing a repository.
type repositoryExpirationState struct {
	// id is the repository identifier.
	id int

	// gitserverClient is a client to gitserver.
	gitserverClient GitserverClient

	// branchesContaining is an LRU cache from commits to the set of branches that contains that commit.
	// Unfortunately we can't easily order our scan over commits, so it is possible to revisit the same
	// commit at arbitrary intervals, but is unlikely as the order of commits and the order of uploads
	// (which we follow) should usually be correlated. An LRU cache therefore is likely to benefit from
	// some degree of natural locality.
	branchesContaining *lru.Cache

	// protectedCommits is the set of commits that have been shown to be protected. Because we process
	// uploads in descending age, once we write a commit to this map, all future uploads we see visible
	// from this commit will necessarily be younger, and therefore also protected by the same policy.
	protectedCommits map[string]struct{}

	// unprotectedCommits is a map from commit to the minimum retention duration that could protected
	// it. Because we process uploads in descending age, once we write a commit to this map, no other
	// policy with a retention duration _smaller_ than this value will be able to protect this commit.
	unprotectedCommits map[string]time.Duration
}

func newRepositoryExpirationState(id int, gitserverClient GitserverClient, maxKeys int) (*repositoryExpirationState, error) {
	branchesContaining, err := lru.New(maxKeys)
	if err != nil {
		return nil, err
	}

	return &repositoryExpirationState{
		id:                 id,
		gitserverClient:    gitserverClient,
		branchesContaining: branchesContaining,
		protectedCommits:   map[string]struct{}{},
		unprotectedCommits: map[string]time.Duration{},
	}, nil
}

// IsProtected returns true if the given commit has already been shown to be protected.
func (s *repositoryExpirationState) IsProtected(commit string) bool {
	_, ok := s.protectedCommits[commit]
	return ok
}

// MarkProtected marks the given commit as protected.
func (s *repositoryExpirationState) MarkProtected(commit string) {
	s.protectedCommits[commit] = struct{}{}
}

// UnprotectedUntil returns the minimum retention duration a policy must have in order to protect the
// given commit. If this value is unknown, a false-valued flag is returned.
func (s *repositoryExpirationState) UnprotectedUntil(commit string) (time.Duration, bool) {
	duration, ok := s.unprotectedCommits[commit]
	return duration, ok
}

// MarkUnprotectedUntil marks the given commit as unprotected by policies that have a smaller-than-provided
// retention duration.
func (s *repositoryExpirationState) MarkUnprotectedUntil(commit string, duration time.Duration) {
	s.unprotectedCommits[commit] = duration
}

// GetBranchesFor will return (fetching if necessary) the set of branch names that include the given commit.
func (s *repositoryExpirationState) GetBranchesFor(ctx context.Context, commit string) ([]string, error) {
	if v, ok := s.branchesContaining.Get(commit); ok {
		return v.([]string), nil
	}

	newBranches, err := s.gitserverClient.BranchesContaining(ctx, s.id, commit)
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.BranchesContaining")
	}

	s.branchesContaining.Add(commit, newBranches)
	return newBranches, nil
}
