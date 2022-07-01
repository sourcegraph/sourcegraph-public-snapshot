package commitgraph

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type DBStore interface {
	MaxStaleAge(ctx context.Context) (_ time.Duration, err error)
}

type UploadService interface {
	GetDirtyRepositories(ctx context.Context) (map[int]int, error)
	GetOldestCommitDate(ctx context.Context, repositoryID int) (time.Time, bool, error)
	UpdateUploadsVisibleToCommits(
		ctx context.Context,
		repositoryID int,
		graph *gitdomain.CommitGraph,
		refDescriptions map[string][]gitdomain.RefDescription,
		maxAgeForNonStaleBranches, maxAgeForNonStaleTags time.Duration,
		dirtyToken int,
		now time.Time,
	) error
}

type Locker interface {
	Lock(ctx context.Context, key int32, blocking bool) (bool, locker.UnlockFunc, error)
}

type GitserverClient interface {
	RefDescriptions(ctx context.Context, repositoryID int, gitOjbs ...string) (map[string][]gitdomain.RefDescription, error)
	CommitGraph(ctx context.Context, repositoryID int, options gitserver.CommitGraphOptions) (*gitdomain.CommitGraph, error)
}
