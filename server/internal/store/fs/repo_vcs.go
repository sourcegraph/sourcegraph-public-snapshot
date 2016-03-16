package fs

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitproto"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
	"sourcegraph.com/sourcegraph/sourcegraph/server/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil"
)

// RepoVCS is a local filesystem-backed implementation of the RepoVCS
// store interface.
type RepoVCS struct{}

var _ store.RepoVCS = (*RepoVCS)(nil)

func (s *RepoVCS) Open(ctx context.Context, repo string) (vcs.Repository, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "RepoVCS.Open", repo); err != nil {
		return nil, err
	}
	r := gitcmd.Open(repo)
	r.AppdashRec = traceutil.Recorder(ctx)
	return r, nil
}

func (s *RepoVCS) Clone(ctx context.Context, repo string, info *store.CloneInfo) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "RepoVCS.Clone", repo); err != nil {
		return err
	}

	return gitserver.Clone(repo, info.CloneURL, &info.RemoteOpts)
}

func (s *RepoVCS) OpenGitTransport(ctx context.Context, repo string) (gitproto.Transport, error) {
	return s.Open(ctx, repo)
}
