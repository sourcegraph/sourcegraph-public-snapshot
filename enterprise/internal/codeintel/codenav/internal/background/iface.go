package background

import (
	"context"
)

type CodeNavService interface {
	SerializeRankingGraph(ctx context.Context, numRankingRoutines int) error
	VacuumRankingGraph(ctx context.Context) error
}
