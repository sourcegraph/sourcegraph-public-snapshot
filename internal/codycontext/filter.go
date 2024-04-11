package codycontext

import (
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type FileChunkFilterFunc func([]FileChunkContext) []FileChunkContext
type RepoContentFilter interface {
	GetFilter(repos []types.RepoIDName, logger log.Logger) ([]types.RepoIDName, FileChunkFilterFunc)
}

func NewRepoContentFilter(logger log.Logger, client gitserver.Client) (RepoContentFilter, error) {
	if dotcom.SourcegraphDotComMode() {
		return newDotcomFilter(client), nil
	}
	return newEnterpriseFilter(logger)
}
