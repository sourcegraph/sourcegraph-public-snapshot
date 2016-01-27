package fs

import (
	"os"
	"path/filepath"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/util/tracer"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/pkg/gitproto"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/traceutil"
)

// RepoVCS is a local filesystem-backed implementation of the RepoVCS
// store interface.
type RepoVCS struct{}

var _ store.RepoVCS = (*RepoVCS)(nil)

func (s *RepoVCS) Open(ctx context.Context, repo string) (vcs.Repository, error) {
	dir := absolutePathForRepo(ctx, repo)
	if err := os.MkdirAll(filepath.Dir(dir), 0700); err != nil {
		return nil, err
	}

	r, err := vcs.Open("git", dir)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		r.Close()
	}()

	return tracer.Wrap(r, traceutil.Recorder(ctx)), nil
}

func (s *RepoVCS) Clone(ctx context.Context, repo string, bare, mirror bool, info *vcsclient.CloneInfo) error {
	dir := absolutePathForRepo(ctx, repo)
	if err := os.MkdirAll(filepath.Dir(dir), 0700); err != nil {
		return err
	}

	_, err := vcs.Clone(info.VCS, info.CloneURL, dir, vcs.CloneOpt{
		Bare:       bare,
		Mirror:     mirror,
		RemoteOpts: info.RemoteOpts,
	})
	return err
}

func (s *RepoVCS) OpenGitTransport(ctx context.Context, repo string) (gitproto.Transport, error) {
	dir := absolutePathForRepo(ctx, repo)
	return &localGitTransport{dir: dir}, nil
}
