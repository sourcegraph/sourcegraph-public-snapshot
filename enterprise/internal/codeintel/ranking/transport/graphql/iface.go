package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/shared"
)

type RankingService interface {
	BumpDerivativeGraphKey(ctx context.Context) error
	Summaries(ctx context.Context) ([]shared.Summary, error)
}
