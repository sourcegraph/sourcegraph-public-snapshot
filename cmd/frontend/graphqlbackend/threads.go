package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Threads is the implementation of the GraphQL type ThreadsMutation. If it is not set at runtime, a
// "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Threads ThreadsResolver

// ThreadByID is called to look up a Thread given its GraphQL ID.
func ThreadByID(ctx context.Context, id graphql.ID) (Thread, error) {
	if Threads == nil {
		return nil, errors.New("threads is not implemented")
	}
	return Threads.ThreadByID(ctx, id)
}

// ThreadsForRepository returns an instance of the GraphQL ThreadConnection type with the list of
// threads defined in a repository.
func ThreadsForRepository(ctx context.Context, repository graphql.ID, arg *graphqlutil.ConnectionArgs) (ThreadConnection, error) {
	if Threads == nil {
		return nil, errors.New("threads is not implemented")
	}
	return Threads.ThreadsForRepository(ctx, repository, arg)
}

func (schemaResolver) Threads() (ThreadsResolver, error) {
	if Threads == nil {
		return nil, errors.New("threads is not implemented")
	}
	return Threads, nil
}

// ThreadsResolver is the interface for the GraphQL type ThreadsMutation.
type ThreadsResolver interface {
	CreateThread(context.Context, *CreateThreadArgs) (Thread, error)
	UpdateThread(context.Context, *UpdateThreadArgs) (Thread, error)
	DeleteThread(context.Context, *DeleteThreadArgs) (*EmptyResponse, error)

	// ThreadByID is called by the ThreadByID func but is not in the GraphQL API.
	ThreadByID(context.Context, graphql.ID) (Thread, error)

	// ThreadsForRepository is called by the ThreadsForRepository func but is not in the GraphQL
	// API.
	ThreadsForRepository(ctx context.Context, repository graphql.ID, arg *graphqlutil.ConnectionArgs) (ThreadConnection, error)
}

type CreateThreadArgs struct {
	Input struct {
		Repository  graphql.ID
		Title       string
		ExternalURL *string
	}
}

type UpdateThreadArgs struct {
	Input struct {
		ID          graphql.ID
		Title       *string
		ExternalURL *string
	}
}

type DeleteThreadArgs struct {
	Thread graphql.ID
}

// Thread is the interface for the GraphQL type Thread.
type Thread interface {
	ID() graphql.ID
	Repository(context.Context) (*RepositoryResolver, error)
	Title() string
	ExternalURL() *string
	URL(context.Context) (string, error)
}

// ThreadConnection is the interface for the GraphQL type ThreadConnection.
type ThreadConnection interface {
	Nodes(context.Context) ([]Thread, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
