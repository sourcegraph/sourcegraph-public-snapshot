package graphqlbackend

import (
	"context"
	"errors"
)

// This file just contains stub GraphQL resolvers and data types for Code Insights which merely
// return an error if not running in enterprise mode. The actual resolvers can be found in
// enterprise/internal/insights/resolvers

type InsightDataPointResolver interface {
	DateTime() DateTime
	Value() float64
}

type PointsArgs struct {
	From *DateTime
	To   *DateTime
}

type InsightsResolver interface {
	// Root resolver
	Insights(ctx context.Context) (InsightsResolver, error)

	// Insights type resolvers.
	Points(ctx context.Context, args *PointsArgs) ([]InsightDataPointResolver, error)
}

var insightsOnlyInEnterprise = errors.New("insights are only available in enterprise")

type defaultInsightsResolver struct{}

func (defaultInsightsResolver) Insights(ctx context.Context) (InsightsResolver, error) {
	return nil, insightsOnlyInEnterprise
}

func (defaultInsightsResolver) Points(ctx context.Context, args *PointsArgs) ([]InsightDataPointResolver, error) {
	return nil, insightsOnlyInEnterprise
}

var DefaultInsightsResolver InsightsResolver = defaultInsightsResolver{}
