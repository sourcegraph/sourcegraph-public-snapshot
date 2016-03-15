package fs

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/pkg/gitproto"
	"src.sourcegraph.com/sourcegraph/pkg/gitserver"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
	"src.sourcegraph.com/sourcegraph/pkg/vcs/gitcmd"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/traceutil"
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
