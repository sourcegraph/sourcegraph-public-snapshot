package live

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

type syncer struct {
	cli *repoupdater.Client
}

func NewSyncer() dependencies.Syncer {
	return &syncer{
		cli: repoupdater.DefaultClient,
	}
}

func (s *syncer) Sync(ctx context.Context, repo api.RepoName) error {
	_, err := s.cli.RepoLookup(ctx, protocol.RepoLookupArgs{Repo: repo, Update: true})
	return err
}
