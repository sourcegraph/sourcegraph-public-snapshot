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
	RetrySearchJob(ctx context.Context, args *RetrySearchJobArgs) (SearchJobResolver, error)

	// Queries
	ValidateSearchJobQuery(ctx context.Context, args *ValidateSearchJobQueryArgs) (ValidateSearchJobQueryResolver, error)
	SearchJobs(ctx context.Context, args *SearchJobsArgs) (SearchJobsConnectionResolver, error)

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
	StartedAt(ctx context.Context) (*gqlutil.DateTime, error)
	FinishedAt(ctx context.Context) (*gqlutil.DateTime, error)
	CsvURL(ctx context.Context) (*string, error)
	RepoStats(ctx context.Context) (SearchJobStatsResolver, error)
	Repositories(ctx context.Context, args *SearchJobRepositoriesArgs) (SearchJobRepoConnectionResolver, error)
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

type SearchJobRepoConnectionResolver interface {
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
	Nodes(ctx context.Context) ([]SearchJobRepoResolver, error)
}

type SearchJobRepoResolver interface {
	ID() graphql.ID
	State() string
	Repository(ctx context.Context) *RepositoryResolver
	CreatedAt() gqlutil.DateTime
	StartedAt(ctx context.Context) (*gqlutil.DateTime, error)
	FinishedAt(ctx context.Context) (*gqlutil.DateTime, error)
	FailureMessage() *string
	Revisions(ctx context.Context, args *SearchJobRepoRevisionsArgs) (SearchJobRepoRevisionConnectionResolver, error)
}

type SearchJobRepoRevisionsArgs struct {
	First int32
	After *string
}

type SearchJobRepoRevisionConnectionResolver interface {
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
	Nodes(ctx context.Context) ([]SearchJobRepoRevisionResolver, error)
}

type SearchJobRepoRevisionResolver interface {
	ID() graphql.ID
	State() string
	Revision() string
	CreatedAt() gqlutil.DateTime
	StartedAt(ctx context.Context) (*gqlutil.DateTime, error)
	FinishedAt(ctx context.Context) (*gqlutil.DateTime, error)
	FailureMessage() *string
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
	First      int32
	After      *string
	Query      *string
	States     *[]string
	OrderBy    string
	Descending bool
}

type SearchJobsConnectionResolver interface {
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
	Nodes(ctx context.Context) ([]SearchJobResolver, error)
}
