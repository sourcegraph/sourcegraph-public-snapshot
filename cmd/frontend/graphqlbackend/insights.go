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
	// Queries
	Insights(ctx context.Context, args *InsightsArgs) (InsightConnectionResolver, error)
	InsightsDashboards(ctx context.Context, args *InsightsDashboardsArgs) (InsightsDashboardConnectionResolver, error)

	// Mutations
	CreateInsightsDashboard(ctx context.Context, args *CreateInsightsDashboardArgs) (InsightsDashboardPayloadResolver, error)
	UpdateInsightsDashboard(ctx context.Context, args *UpdateInsightsDashboardArgs) (InsightsDashboardPayloadResolver, error)
	DeleteInsightsDashboard(ctx context.Context, args *DeleteInsightsDashboardArgs) (*EmptyResponse, error)
	RemoveInsightViewFromDashboard(ctx context.Context, args *RemoveInsightViewFromDashboardArgs) (InsightsDashboardPayloadResolver, error)
	AddInsightViewToDashboard(ctx context.Context, args *AddInsightViewToDashboardArgs) (InsightsDashboardPayloadResolver, error)
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

type InsightsDashboardsArgs struct {
	First *int32
	After *string
}

type InsightsDashboardConnectionResolver interface {
	Nodes(ctx context.Context) ([]InsightsDashboardResolver, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type InsightsDashboardResolver interface {
	Title() string
	ID() graphql.ID
	Views() InsightViewConnectionResolver
}

type CreateInsightsDashboardArgs struct {
	Input CreateInsightsDashboardInput
}

type CreateInsightsDashboardInput struct {
	Title  string
	Grants InsightsPermissionGrants
}

type UpdateInsightsDashboardArgs struct {
	Id    graphql.ID
	Input UpdateInsightsDashboardInput
}

type UpdateInsightsDashboardInput struct {
	Title  *string
	Grants *InsightsPermissionGrants
}

type InsightsPermissionGrants struct {
	Users         *[]graphql.ID
	Organizations *[]graphql.ID
	Global        *bool
}

type DeleteInsightsDashboardArgs struct {
	Id graphql.ID
}

type InsightViewConnectionResolver interface {
	Nodes(ctx context.Context) ([]InsightViewResolver, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type InsightViewResolver interface {
	ID() graphql.ID
	// Until this interface becomes uniquely identifyable in the node resolvers
	// ToXX type guard methods, we need _something_ that makes this interface
	// not match any other Node implementing type.
	VeryUniqueResolver() bool
}

type InsightsDashboardPayloadResolver interface {
	Dashboard(ctx context.Context) (InsightsDashboardResolver, error)
}

type AddInsightViewToDashboardArgs struct {
	Input AddInsightViewToDashboardInput
}

type AddInsightViewToDashboardInput struct {
	InsightViewID graphql.ID
	DashboardID   graphql.ID
}

type RemoveInsightViewFromDashboardArgs struct {
	Input RemoveInsightViewFromDashboardInput
}

type RemoveInsightViewFromDashboardInput struct {
	InsightViewID graphql.ID
	DashboardID   graphql.ID
}
