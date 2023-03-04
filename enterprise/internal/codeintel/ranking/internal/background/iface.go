package background

import (
	"context"
)

type RankingService interface {
	ExportRankingGraph(ctx context.Context, numRoutines int, numBatchSize int) error
	MapRankingGraph(ctx context.Context) (int, int, error)
	ReduceRankingGraph(ctx context.Context) (float64, float64, error)
	VacuumRankingGraph(ctx context.Context) error
}
