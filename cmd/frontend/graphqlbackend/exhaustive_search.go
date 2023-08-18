package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type ExhaustiveSearchesResolver interface {
	// Mutations
	CreateExhaustiveSearch(ctx context.Context, args *CreateExhaustiveSearchArgs) (ExhaustiveSearchResolver, error)
	CancelExhaustiveSearch(ctx context.Context, args *CancelExhaustiveSearchArgs) (*EmptyResponse, error)
	DeleteExhaustiveSearch(ctx context.Context, args *DeleteExhaustiveSearchArgs) (*EmptyResponse, error)
	RetryExhaustiveSearch(ctx context.Context, args *RetryExhaustiveSearchArgs) (ExhaustiveSearchResolver, error)

	// Queries
	ValidateExhaustiveSearchQuery(ctx context.Context, args *ValidateExhaustiveSearchQueryArgs) (ValidateExhaustiveSearchQueryResolver, error)
	ExhaustiveSearches(ctx context.Context, args *ExhaustiveSearchesArgs) (ExhaustiveSearchesConnectionResolver, error)

	NodeResolvers() map[string]NodeByIDFunc
}

type ValidateExhaustiveSearchQueryArgs struct {
	Query string
}

type ValidateExhaustiveSearchQueryResolver interface {
	Query() string
	Valid() bool
	Errors() *[]string
}

type CreateExhaustiveSearchArgs struct {
	Query string
}

type ExhaustiveSearchResolver interface {
	ID() graphql.ID
	Query() string
	State(ctx context.Context) string
	Creator(ctx context.Context) (*UserResolver, error)
	CreatedAt() gqlutil.DateTime
	StartedAt(ctx context.Context) (*gqlutil.DateTime, error)
	FinishedAt(ctx context.Context) (*gqlutil.DateTime, error)
	CsvURL(ctx context.Context) (*string, error)
	RepoStats(ctx context.Context) (ExhaustiveSearchStatsResolver, error)
	Repositories(ctx context.Context, args *ExhaustiveSearchRepositoriesArgs) (ExhaustiveSearchRepoConnectionResolver, error)
}

type ExhaustiveSearchStatsResolver interface {
	Total() int32
	Completed() int32
	Errored() int32
	InProgress() int32
}

type ExhaustiveSearchRepositoriesArgs struct {
	First int32
	After *string
}

type ExhaustiveSearchRepoConnectionResolver interface {
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
	Nodes(ctx context.Context) ([]ExhaustiveSearchRepoResolver, error)
}

type ExhaustiveSearchRepoResolver interface {
	ID() graphql.ID
	State() string
	Repository(ctx context.Context) *RepositoryResolver
	CreatedAt() gqlutil.DateTime
	StartedAt(ctx context.Context) (*gqlutil.DateTime, error)
	FinishedAt(ctx context.Context) (*gqlutil.DateTime, error)
	FailureMessage() *string
	Revisions(ctx context.Context, args *ExhaustiveSearchRepoRevisionsArgs) (ExhaustiveSearchRepoRevisionConnectionResolver, error)
}

type ExhaustiveSearchRepoRevisionsArgs struct {
	First int32
	After *string
}

type ExhaustiveSearchRepoRevisionConnectionResolver interface {
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
	Nodes(ctx context.Context) ([]ExhaustiveSearchRepoRevisionResolver, error)
}

type ExhaustiveSearchRepoRevisionResolver interface {
	ID() graphql.ID
	State() string
	Revision() string
	CreatedAt() gqlutil.DateTime
	StartedAt(ctx context.Context) (*gqlutil.DateTime, error)
	FinishedAt(ctx context.Context) (*gqlutil.DateTime, error)
	FailureMessage() *string
}

type CancelExhaustiveSearchArgs struct {
	ID graphql.ID
}

type DeleteExhaustiveSearchArgs struct {
	ID graphql.ID
}

type RetryExhaustiveSearchArgs struct {
	ID graphql.ID
}

type ExhaustiveSearchArgs struct {
	ID graphql.ID
}

type ExhaustiveSearchesArgs struct {
	First      int32
	After      *string
	Query      *string
	States     *[]string
	OrderBy    string
	Descending bool
}

type ExhaustiveSearchesConnectionResolver interface {
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
	Nodes(ctx context.Context) ([]ExhaustiveSearchResolver, error)
}
