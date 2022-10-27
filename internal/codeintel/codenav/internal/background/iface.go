package background

import (
	"context"
)

type CodeNavService interface {
	SerializeRankingGraph(ctx context.Context) error
}
