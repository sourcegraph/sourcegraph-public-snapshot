package codycontext

import (
	"context"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type FileChunkFilterFunc func([]FileChunkContext) []FileChunkContext
type RepoContentFilter interface {
	GetFilter(ctx context.Context, repos []types.RepoIDName) ([]types.RepoIDName, FileChunkFilterFunc, error)
}

func newRepoContentFilter(logger log.Logger, client gitserver.Client) RepoContentFilter {
	if dotcom.SourcegraphDotComMode() {
		return newDotcomFilter(logger, client)
	}
	return newEnterpriseFilter(logger)
}
