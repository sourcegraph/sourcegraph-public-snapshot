package authzchecked

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/pkg/gitproto"
	"src.sourcegraph.com/sourcegraph/store"
)

// RepoVCS wraps base's methods with authorization checks.
func RepoVCS(base store.RepoVCS) store.RepoVCS { return &repoVCS{base} }

// repoVCS adds authorization checks to an underlying RepoVCS.
type repoVCS struct {
	noauthz store.RepoVCS
}

func (s *repoVCS) Open(ctx context.Context, repo string) (vcs.Repository, error) {
	if err := auth.CheckRepo(ctx, repo, auth.Read); err != nil {
		return nil, err
	}
	return s.noauthz.Open(ctx, repo)
}

func (s *repoVCS) Clone(ctx context.Context, repo string, bare, mirror bool, info *vcsclient.CloneInfo) error {
	if err := auth.CheckRepo(ctx, repo, auth.Write); err != nil {
		return err
	}
	return s.noauthz.Clone(ctx, repo, bare, mirror, info)
}

func (s *repoVCS) OpenGitTransport(ctx context.Context, repo string) (gitproto.Transport, error) {
	if err := auth.CheckRepo(ctx, repo, auth.Write); err != nil {
		return nil, err
	}
	return s.noauthz.OpenGitTransport(ctx, repo)
}
