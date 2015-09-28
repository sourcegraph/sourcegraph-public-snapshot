package authzchecked

import (
	"time"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/store"
)

// RepoCounters wraps base's methods with authorization checks.
func RepoCounters(base store.RepoCounters) store.RepoCounters { return &repoCounters{base} }

// repoCounters adds authorization checks to an underlying
// RepoCounters.
type repoCounters struct {
	noauthz store.RepoCounters
}

func (s *repoCounters) RecordHit(ctx context.Context, repo string) error {
	if err := auth.CheckRepo(ctx, repo, auth.Read); err != nil {
		return err
	}
	return s.noauthz.RecordHit(ctx, repo)
}

func (s *repoCounters) CountHits(ctx context.Context, repo string, since time.Time) (int, error) {
	if err := auth.CheckRepo(ctx, repo, auth.Read); err != nil {
		return 0, err
	}
	return s.noauthz.CountHits(ctx, repo, since)
}
