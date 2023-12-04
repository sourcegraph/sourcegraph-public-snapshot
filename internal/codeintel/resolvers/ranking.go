package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type RankingServiceResolver interface {
	RankingSummary(ctx context.Context) (GlobalRankingSummaryResolver, error)
	BumpDerivativeGraphKey(ctx context.Context) (*EmptyResponse, error)
	DeleteRankingProgress(ctx context.Context, args *DeleteRankingProgressArgs) (*EmptyResponse, error)
}

type DeleteRankingProgressArgs struct {
	GraphKey string
}

type GlobalRankingSummaryResolver interface {
	DerivativeGraphKey() *string
	RankingSummary() []RankingSummaryResolver
	NextJobStartsAt() *gqlutil.DateTime
	NumExportedIndexes() int32
	NumTargetIndexes() int32
	NumRepositoriesWithoutCurrentRanks() int32
}

type RankingSummaryResolver interface {
	GraphKey() string
	VisibleToZoekt() bool
	PathMapperProgress() RankingSummaryProgressResolver
	ReferenceMapperProgress() RankingSummaryProgressResolver
	ReducerProgress() RankingSummaryProgressResolver
}

type RankingSummaryProgressResolver interface {
	StartedAt() gqlutil.DateTime
	CompletedAt() *gqlutil.DateTime
	Processed() int32
	Total() int32
}
