package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/shared"
)

type RankingService interface {
	DerivativeGraphKey(ctx context.Context) (string, bool, error)
	BumpDerivativeGraphKey(ctx context.Context) error
	Summaries(ctx context.Context) ([]shared.Summary, error)
	NextJobStartsAt(ctx context.Context) (time.Time, bool, error)
	CoverageCounts(ctx context.Context, graphKey string) (shared.CoverageCounts, error)
	DeleteRankingProgress(ctx context.Context, graphKey string) error
}
