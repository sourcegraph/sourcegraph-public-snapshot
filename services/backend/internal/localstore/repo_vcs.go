package localstore

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitproto"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

// repoVCS is a local filesystem-backed implementation of the RepoVCS
// store interface.
type repoVCS struct{}

var _ store.RepoVCS = (*repoVCS)(nil)

func (s *repoVCS) Open(ctx context.Context, repo string) (vcs.Repository, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "RepoVCS.Open", repo); err != nil {
		return nil, err
	}
	r := gitcmd.Open(repo)
	r.AppdashRec = traceutil.Recorder(ctx)
	return r, nil
}

func (s *repoVCS) Clone(ctx context.Context, repo string, info *store.CloneInfo) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "RepoVCS.Clone", repo); err != nil {
		return err
	}

	return gitserver.Clone(repo, info.CloneURL, &info.RemoteOpts)
}

func (s *repoVCS) OpenGitTransport(ctx context.Context, repo string) (gitproto.Transport, error) {
	return s.Open(ctx, repo)
}
