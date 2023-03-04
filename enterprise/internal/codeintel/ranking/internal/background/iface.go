package background

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type RankingService interface {
	ExportRankingGraph(ctx context.Context, numRoutines int, numBatchSize int) error
	MapRankingGraph(ctx context.Context) (int, int, error)
	ReduceRankingGraph(ctx context.Context) (float64, float64, error)
	VacuumRankingGraph(ctx context.Context) error
}

type GitserverClient interface {
	RefDescriptions(ctx context.Context, repositoryID int, pointedAt ...string) (_ map[string][]gitdomain.RefDescription, err error)
}
