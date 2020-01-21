package usagestats

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/usagestats"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
)

type UsageStats struct{}

func NewUsageStats() usagestats.CodeIntelUsageStats {
	return &UsageStats{}
}

func (us *UsageStats) NumUpToDateRepositoriesWithLSIF(ctx context.Context) (int, error) {
	stats, err := client.DefaultClient.Stats(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, mostRecentUpload := range stats.MostRecentUploads {
		repo, err := backend.Repos.Get(ctx, mostRecentUpload.RepositoryID)
		if err != nil {
			return 0, err
		}

		branch, err := graphqlbackend.NewRepositoryResolver(repo).DefaultBranch(ctx)
		if err != nil {
			return 0, err
		}

		commit, err := branch.Target().Commit(ctx)
		if err != nil {
			return 0, err
		}

		if mostRecentUpload.UploadedAt.Sub(commit.Committer().RawDate()) >= -time.Hour*24 {
			count++
		}
	}

	return count, nil
}
