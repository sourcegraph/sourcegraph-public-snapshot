package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type SearchJobsResolver interface {
	// Mutations
	CreateSearchJob(ctx context.Context, args *CreateSearchJobArgs) (SearchJobResolver, error)
	CancelSearchJob(ctx context.Context, args *CancelSearchJobArgs) (*EmptyResponse, error)
	DeleteSearchJob(ctx context.Context, args *DeleteSearchJobArgs) (*EmptyResponse, error)

	// Queries
	SearchJobs(ctx context.Context, args *SearchJobsArgs) (*graphqlutil.ConnectionResolver[SearchJobResolver], error)

	NodeResolvers() map[string]NodeByIDFunc
}

type ValidateSearchJobQueryArgs struct {
	Query string
}

type ValidateSearchJobQueryResolver interface {
	Query() string
	Valid() bool
	Errors() *[]string
}

type CreateSearchJobArgs struct {
	Query string
}

type SearchJobResolver interface {
	ID() graphql.ID
	Query() string
	State(ctx context.Context) string
	Creator(ctx context.Context) (*UserResolver, error)
	CreatedAt() gqlutil.DateTime
	StartedAt(ctx context.Context) *gqlutil.DateTime
	FinishedAt(ctx context.Context) *gqlutil.DateTime
	URL(ctx context.Context) (*string, error)
	LogURL(ctx context.Context) (*string, error)
	RepoStats(ctx context.Context) (SearchJobStatsResolver, error)
}

type SearchJobStatsResolver interface {
	Total() int32
	Completed() int32
	Failed() int32
	InProgress() int32
}

type SearchJobRepositoriesArgs struct {
	First int32
	After *string
}

type SearchJobRepoRevisionsArgs struct {
	First int32
	After *string
}

type CancelSearchJobArgs struct {
	ID graphql.ID
}

type DeleteSearchJobArgs struct {
	ID graphql.ID
}

type RetrySearchJobArgs struct {
	ID graphql.ID
}

type SearchJobArgs struct {
	ID graphql.ID
}

type SearchJobsArgs struct {
	graphqlutil.ConnectionResolverArgs
	Query      *string
	States     *[]string
	OrderBy    string
	Descending bool
	UserIDs    *[]graphql.ID
}

type SearchJobsConnectionResolver interface {
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
	Nodes(ctx context.Context) ([]SearchJobResolver, error)
}
