package inference

import (
	"context"
	"regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

type GitserverClientWrapper interface {
	FileExists(ctx context.Context, file string) (bool, error)
	RawContents(ctx context.Context, file string) ([]byte, error)
	ListFiles(ctx context.Context, pattern *regexp.Regexp) ([]string, error)
}

type GitserverClient interface {
	FileExists(ctx context.Context, repositoryID int, commit, file string) (bool, error)
	RawContents(ctx context.Context, repositoryID int, commit, file string) ([]byte, error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
}

type GitserverClientShim struct {
	gitserverClient GitserverClient
	commit          string
	repositoryID    int
}

var _ GitserverClientWrapper = &GitserverClientShim{}

func NewGitserverClientShim(repositoryID int, commit string, gitserverClient GitserverClient) *GitserverClientShim {
	return &GitserverClientShim{
		gitserverClient, commit, repositoryID,
	}
}

func (s *GitserverClientShim) FileExists(ctx context.Context, file string) (bool, error) {
	return s.gitserverClient.FileExists(ctx, s.repositoryID, s.commit, file)
}

func (s *GitserverClientShim) RawContents(ctx context.Context, file string) ([]byte, error) {
	return s.gitserverClient.RawContents(ctx, s.repositoryID, s.commit, file)
}

func (s *GitserverClientShim) ListFiles(ctx context.Context, pattern *regexp.Regexp) ([]string, error) {
	return s.gitserverClient.ListFiles(ctx, s.repositoryID, s.commit, pattern)
}

type RepoUpdaterClient interface {
	EnqueueRepoUpdate(ctx context.Context, repo api.RepoName) (*protocol.RepoUpdateResponse, error)
}
