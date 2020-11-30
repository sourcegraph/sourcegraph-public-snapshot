package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

type repositoryStatsResolver struct {
	gitDirBytes       uint64
	indexedLinesCount uint64
}

func (r *repositoryStatsResolver) GitDirBytes() int32 {
	return int32(r.gitDirBytes)
}

func (r *repositoryStatsResolver) IndexedLinesCount() int32 {
	return int32(r.indexedLinesCount)
}

func (r *schemaResolver) RepositoryStats(ctx context.Context) (*repositoryStatsResolver, error) {
	// 🚨 SECURITY: Only site admins may query repository statistics for the site.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	stats, err := usagestats.GetRepositories(ctx)
	if err != nil {
		return nil, err
	}

	return &repositoryStatsResolver{
		gitDirBytes:       stats.GitDirBytes,
		indexedLinesCount: stats.DefaultBranchNewLinesCount + stats.OtherBranchesNewLinesCount,
	}, nil
}
