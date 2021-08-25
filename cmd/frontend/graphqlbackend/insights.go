package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// This file just contains stub GraphQL resolvers and data types for Code Insights which merely
// return an error if not running in enterprise mode. The actual resolvers can be found in
// enterprise/internal/insights/resolvers

// InsightsResolver is the root resolver.
type InsightsResolver interface {
	Insights(ctx context.Context, args *InsightsArgs) (InsightConnectionResolver, error)
}

type InsightsArgs struct {
	Ids *[]graphql.ID
}

type InsightsDataPointResolver interface {
	DateTime() DateTime
	Value() float64
}

type InsightStatusResolver interface {
	TotalPoints() int32
	PendingJobs() int32
	CompletedJobs() int32
	FailedJobs() int32
	BackfillQueuedAt() *DateTime
}

type InsightsPointsArgs struct {
	From             *DateTime
	To               *DateTime
	IncludeRepoRegex *string
	ExcludeRepoRegex *string
}

type InsightSeriesResolver interface {
	Label() string
	Points(ctx context.Context, args *InsightsPointsArgs) ([]InsightsDataPointResolver, error)
	Status(ctx context.Context) (InsightStatusResolver, error)
	DirtyMetadata(ctx context.Context) ([]InsightDirtyQueryResolver, error)
}

type InsightResolver interface {
	Title() string
	Description() string
	Series() []InsightSeriesResolver
	ID() string
}

type InsightConnectionResolver interface {
	Nodes(ctx context.Context) ([]InsightResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type InsightDirtyQueryResolver interface {
	Reason(ctx context.Context) string
	Time(ctx context.Context) DateTime
	Count(ctx context.Context) int32
}
