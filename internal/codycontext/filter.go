package codycontext

import (
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type FileChunkFilterFunc func([]FileChunkContext) []FileChunkContext
type RepoContentFilter interface {
	GetFilter(repos []types.RepoIDName, logger log.Logger) (_ []types.RepoIDName, _ FileChunkFilterFunc, ok bool)
}

func newRepoContentFilter(logger log.Logger, client gitserver.Client) RepoContentFilter {
	if dotcom.SourcegraphDotComMode() {
		return newDotcomFilter(client)
	}
	return newEnterpriseFilter(logger)
}
