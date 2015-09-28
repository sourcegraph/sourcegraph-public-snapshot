package fs

import (
	"os"
	"path/filepath"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/pkg/gitproto"
	"src.sourcegraph.com/sourcegraph/store"
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

	return vcs.Open("git", dir)
}

func (s *RepoVCS) Clone(ctx context.Context, repo string, info *vcsclient.CloneInfo) error {
	dir := absolutePathForRepo(ctx, repo)
	if err := os.MkdirAll(filepath.Dir(dir), 0700); err != nil {
		return err
	}

	_, err := vcs.Clone(info.VCS, info.CloneURL, dir, vcs.CloneOpt{
		Bare:       true,
		Mirror:     true,
		RemoteOpts: info.RemoteOpts,
	})
	return err
}

func (s *RepoVCS) OpenGitTransport(ctx context.Context, repo string) (gitproto.Transport, error) {
	dir := absolutePathForRepo(ctx, repo)
	return &localGitTransport{dir: dir}, nil
}
