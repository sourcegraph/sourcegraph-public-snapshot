package idx

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
)

var MockResolveRevision func(ctx context.Context, repo *sourcegraph.Repo, spec string) (vcs.CommitID, error)

func ResolveRevision(ctx context.Context, repo *sourcegraph.Repo, spec string) (vcs.CommitID, error) {
	if MockResolveRevision != nil {
		return MockResolveRevision(ctx, repo, spec)
	}
	return gitcmd.Open(repo).ResolveRevision(ctx, spec)
}
