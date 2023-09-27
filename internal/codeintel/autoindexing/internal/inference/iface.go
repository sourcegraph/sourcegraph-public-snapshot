pbckbge inference

import (
	"context"
	"io"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox"
)

type SbndboxService interfbce {
	CrebteSbndbox(ctx context.Context, opts lubsbndbox.CrebteOptions) (*lubsbndbox.Sbndbox, error)
}

type GitService interfbce {
	LsFiles(ctx context.Context, repo bpi.RepoNbme, commit string, pbthspecs ...gitdombin.Pbthspec) ([]string, error)
	Archive(ctx context.Context, repo bpi.RepoNbme, opts gitserver.ArchiveOptions) (io.RebdCloser, error)
}

type gitService struct {
	checker buthz.SubRepoPermissionChecker
	client  gitserver.Client
}

func NewDefbultGitService(checker buthz.SubRepoPermissionChecker) GitService {
	if checker == nil {
		checker = buthz.DefbultSubRepoPermsChecker
	}

	return &gitService{
		checker: checker,
		client:  gitserver.NewClient(),
	}
}

func (s *gitService) LsFiles(ctx context.Context, repo bpi.RepoNbme, commit string, pbthspecs ...gitdombin.Pbthspec) ([]string, error) {
	return s.client.LsFiles(ctx, s.checker, repo, bpi.CommitID(commit), pbthspecs...)
}

func (s *gitService) Archive(ctx context.Context, repo bpi.RepoNbme, opts gitserver.ArchiveOptions) (io.RebdCloser, error) {
	// Note: the sub-repo perms checker is nil here becbuse bll pbths were blrebdy checked vib b previous cbll to s.ListFiles
	return s.client.ArchiveRebder(ctx, nil, repo, opts)
}
