package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Threads is the implementation of the GraphQL threads queries and mutations. If it is not set at
// runtime, a "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Threads ThreadsResolver

var errThreadsNotImplemented = errors.New("threads is not implemented")

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
		return nil, errThreadsNotImplemented
	}
	return Threads.ThreadsForRepository(ctx, repository, arg)
}

func (schemaResolver) Threads(ctx context.Context, arg *graphqlutil.ConnectionArgs) (ThreadConnection, error) {
	if Threads == nil {
		return nil, errThreadsNotImplemented
	}
	return Threads.Threads(ctx, arg)
}

func (r schemaResolver) CreateThread(ctx context.Context, arg *CreateThreadArgs) (Thread, error) {
	if Threads == nil {
		return nil, errThreadsNotImplemented
	}
	return Threads.CreateThread(ctx, arg)
}

func (r schemaResolver) UpdateThread(ctx context.Context, arg *UpdateThreadArgs) (Thread, error) {
	if Threads == nil {
		return nil, errThreadsNotImplemented
	}
	return Threads.UpdateThread(ctx, arg)
}

func (r schemaResolver) DeleteThread(ctx context.Context, arg *DeleteThreadArgs) (*EmptyResponse, error) {
	if Threads == nil {
		return nil, errThreadsNotImplemented
	}
	return Threads.DeleteThread(ctx, arg)
}

// ThreadsResolver is the interface for the GraphQL threads queries and mutations.
type ThreadsResolver interface {
	// Queries
	Threads(context.Context, *graphqlutil.ConnectionArgs) (ThreadConnection, error)

	// Mutations
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
		Settings    *string
		Status      *ThreadStatus
		Type        ThreadType
	}
}

type UpdateThreadArgs struct {
	Input struct {
		ID          graphql.ID
		Title       *string
		ExternalURL *string
		Settings    *string
		Status      *ThreadStatus
	}
}

type DeleteThreadArgs struct {
	Thread graphql.ID
}

type ThreadType string

const (
	ThreadTypeThread    ThreadType = "THREAD"
	ThreadTypeIssue                = "ISSUE"
	ThreadTypeChangeset            = "CHANGESET"
)

// IsValidThreadType reports whether t is a valid thread type.
func IsValidThreadType(t string) bool {
	return ThreadType(t) == ThreadTypeThread || ThreadType(t) == ThreadTypeIssue || ThreadType(t) == ThreadTypeChangeset
}

type ThreadStatus string

const (
	ThreadStatusPreview ThreadStatus = "PREVIEW"
	ThreadStatusOpen                 = "OPEN"
	ThreadStatusMerged               = "MERGED"
	ThreadStatusClosed               = "CLOSED"
)

// IsValidThreadStatus reports whether t is a valid thread status.
func IsValidThreadStatus(t string) bool {
	return ThreadStatus(t) == ThreadStatusPreview || ThreadStatus(t) == ThreadStatusOpen || ThreadStatus(t) == ThreadStatusMerged || ThreadStatus(t) == ThreadStatusClosed
}

// Thread is the interface for the GraphQL type Thread.
type Thread interface {
	ID() graphql.ID
	IDWithoutKind() string
	DBID() int64
	Repository(context.Context) (*RepositoryResolver, error)
	Title() string
	ExternalURL() *string
	URL(context.Context) (string, error)
	Settings() string
	Status() ThreadStatus
	Type() ThreadType
}

// ThreadConnection is the interface for the GraphQL type ThreadConnection.
type ThreadConnection interface {
	Nodes(context.Context) ([]Thread, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
