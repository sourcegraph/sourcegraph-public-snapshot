package graphqlbackend

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// This file just contains stub GraphQL resolvers and data types for Code Insights which merely
// return an error if not running in enterprise mode. The actual resolvers can be found in
// enterprise/internal/insights/resolvers

type InsightsDataPointResolver interface {
	DateTime() DateTime
	Value() float64
}

type InsightsPointsArgs struct {
	From *DateTime
	To   *DateTime
}

type InsightSeriesResolver interface {
	Label() string
	Points(ctx context.Context, args *InsightsPointsArgs) ([]InsightsDataPointResolver, error)
}

type InsightResolver interface {
	Title() string
	Description() string
	Series() []InsightSeriesResolver
}

type InsightConnectionResolver interface {
	Nodes(ctx context.Context) ([]InsightResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

// InsightsResolver is the root resolver.
type InsightsResolver interface {
	Insights(ctx context.Context) (InsightConnectionResolver, error)
}

var insightsOnlyInEnterprise = errors.New("insights are only available in enterprise")

type defaultInsightsResolver struct{}

func (defaultInsightsResolver) Insights(ctx context.Context) (InsightConnectionResolver, error) {
	return nil, insightsOnlyInEnterprise
}

var DefaultInsightsResolver InsightsResolver = defaultInsightsResolver{}
