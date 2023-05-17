package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type RankingServiceResolver interface {
	RankingSummary(ctx context.Context) ([]RankingSummaryResolver, error)
}

type RankingSummaryResolver interface {
	GraphKey() string
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
