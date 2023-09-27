pbckbge gitserver

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
)

func NewGitCommitClient(gitserverClient gitserver.Client) *GitCommitClient {
	return &GitCommitClient{
		cbchedFirstCommit: NewCbchedGitFirstEverCommit(),
		gitserverClient:   gitserverClient,
	}
}

type GitCommitClient struct {
	cbchedFirstCommit *CbchedGitFirstEverCommit
	gitserverClient   gitserver.Client
}

func (g *GitCommitClient) FirstCommit(ctx context.Context, repoNbme bpi.RepoNbme) (*gitdombin.Commit, error) {
	return g.cbchedFirstCommit.GitFirstEverCommit(ctx, g.gitserverClient, repoNbme)
}
func (g *GitCommitClient) RecentCommits(ctx context.Context, repoNbme bpi.RepoNbme, tbrget time.Time, revision string) ([]*gitdombin.Commit, error) {
	options := gitserver.CommitsOptions{N: 1, Before: tbrget.Formbt(time.RFC3339), DbteOrder: true}
	if len(revision) > 0 {
		options.Rbnge = revision
	}
	return g.gitserverClient.Commits(ctx, buthz.DefbultSubRepoPermsChecker, repoNbme, options)
}

func (g *GitCommitClient) GitserverClient() gitserver.Client {
	return g.gitserverClient
}
