package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// Threads is the implementation of the GraphQL API for threads queries and mutations. If it is not
// set at runtime, a "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Threads ThreadsResolver

const GQLTypeThread = "Thread"

func MarshalThreadID(id int64) graphql.ID {
	return relay.MarshalID(GQLTypeThread, id)
}

func UnmarshalThreadID(id graphql.ID) (dbID int64, err error) {
	if typ := relay.UnmarshalKind(id); typ != GQLTypeThread {
		return 0, fmt.Errorf("thread ID has unexpected type type %q", typ)
	}
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

var errThreadsNotImplemented = errors.New("threads is not implemented")

// ThreadByID is called to look up a Thread given its GraphQL ID.
func ThreadByID(ctx context.Context, id graphql.ID) (Thread, error) {
	if Threads == nil {
		return nil, errThreadsNotImplemented
	}
	return Threads.ThreadByID(ctx, id)
}

// ThreadInRepository returns a specific thread in the specified repository.
func ThreadInRepository(ctx context.Context, repository graphql.ID, number string) (Thread, error) {
	if Threads == nil {
		return nil, errThreadsNotImplemented
	}
	return Threads.ThreadInRepository(ctx, repository, number)
}

// ThreadsForRepository returns an instance of the GraphQL ThreadConnection type with the list of
// threads defined in a repository.
func ThreadsForRepository(ctx context.Context, repository graphql.ID, arg *ThreadConnectionArgs) (ThreadConnection, error) {
	if Threads == nil {
		return nil, errThreadsNotImplemented
	}
	return Threads.ThreadsForRepository(ctx, repository, arg)
}

func (schemaResolver) Threads(ctx context.Context, arg *ThreadConnectionArgs) (ThreadConnection, error) {
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

func (r schemaResolver) ImportThreadsFromExternalService(ctx context.Context, arg *ImportThreadsFromExternalServiceArgs) ([]Thread, error) {
	if Threads == nil {
		return nil, errThreadsNotImplemented
	}
	return Threads.ImportThreadsFromExternalService(ctx, arg)
}

func (r schemaResolver) UpdateThread(ctx context.Context, arg *UpdateThreadArgs) (Thread, error) {
	if Threads == nil {
		return nil, errThreadsNotImplemented
	}
	return Threads.UpdateThread(ctx, arg)
}

func (r schemaResolver) MarkThreadAsReady(ctx context.Context, arg *MarkThreadAsReadyArgs) (Thread, error) {
	if Threads == nil {
		return nil, errThreadsNotImplemented
	}
	return Threads.MarkThreadAsReady(ctx, arg)
}

func (r schemaResolver) PublishThreadToExternalService(ctx context.Context, arg *PublishThreadToExternalServiceArgs) (Thread, error) {
	if Threads == nil {
		return nil, errThreadsNotImplemented
	}
	return Threads.PublishThreadToExternalService(ctx, arg)
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
	Threads(context.Context, *ThreadConnectionArgs) (ThreadConnection, error)

	// Mutations
	CreateThread(context.Context, *CreateThreadArgs) (Thread, error)
	ImportThreadsFromExternalService(context.Context, *ImportThreadsFromExternalServiceArgs) ([]Thread, error)
	UpdateThread(context.Context, *UpdateThreadArgs) (Thread, error)
	MarkThreadAsReady(context.Context, *MarkThreadAsReadyArgs) (Thread, error)
	PublishThreadToExternalService(context.Context, *PublishThreadToExternalServiceArgs) (Thread, error)
	DeleteThread(context.Context, *DeleteThreadArgs) (*EmptyResponse, error)

	// ThreadByID is called by the ThreadByID func but is not in the GraphQL API.
	ThreadByID(context.Context, graphql.ID) (Thread, error)

	// ThreadInRepository is called by the ThreadInRepository func but is not in the GraphQL API.
	ThreadInRepository(ctx context.Context, repository graphql.ID, number string) (Thread, error)

	// ThreadsForRepository is called by the ThreadsForRepository func but is not in the GraphQL
	// API.
	ThreadsForRepository(ctx context.Context, repository graphql.ID, arg *ThreadConnectionArgs) (ThreadConnection, error)
}

type ThreadConnectionArgs struct {
	graphqlutil.ConnectionArgs
	Filters *ThreadFiltersInput
}

type ThreadFiltersInput struct {
	Query        *string
	Repositories *[]graphql.ID
	States       *[]ThreadState
}

type CreateThreadInput struct {
	Repository     graphql.ID
	Title          string
	Body           *string
	Draft          *bool
	BaseRef        *string
	HeadRef        *string
	RawDiagnostics *[]string

	Internal_PendingExternalCreation bool
}

type CreateThreadArgs struct {
	Input CreateThreadInput
}

type ImportThreadsFromExternalServiceArgs struct {
	Input struct {
		ByRepositoryAndNumber *struct {
			RepositoryName string
			Number         int32
		}
		ByQuery *string
	}
}

type UpdateThreadArgs struct {
	Input struct {
		ID      graphql.ID
		Title   *string
		Body    *string
		BaseRef *string
		HeadRef *string
	}
}

type MarkThreadAsReadyArgs struct {
	Thread graphql.ID
}

type PublishThreadToExternalServiceArgs struct {
	Thread graphql.ID
}

type DeleteThreadArgs struct {
	Thread graphql.ID
}

type ThreadState string

const (
	ThreadStateOpen   ThreadState = "OPEN"
	ThreadStateMerged ThreadState = "MERGED"
	ThreadStateClosed ThreadState = "CLOSED"
)

type ThreadKind string

const (
	ThreadKindDiscussion ThreadKind = "DISCUSSION"
	ThreadKindIssue      ThreadKind = "ISSUE"
	ThreadKindChangeset  ThreadKind = "CHANGESET"
)

// Thread is the interface for the GraphQL type Thread.
type Thread interface {
	PartialComment
	ID() graphql.ID
	DBID() int64
	Repository(context.Context) (*RepositoryResolver, error)
	Internal_RepositoryID() api.RepoID
	Number() string
	Title() string
	IsDraft() bool
	IsPendingExternalCreation() bool
	State() ThreadState
	BaseRef() *string
	HeadRef() *string
	hasThreadDiagnostics
	Updatable
	commentable
	ruleContainer
	Kind(context.Context) (ThreadKind, error)
	URL(context.Context) (string, error)
	ExternalURLs(context.Context) ([]*externallink.Resolver, error)
	TimelineItems(context.Context, *EventConnectionCommonArgs) (EventConnection, error)
	RepositoryComparison(context.Context) (RepositoryComparison, error)
	CampaignNode
	Assignable
	Labelable
}

// ThreadConnection is the interface for the GraphQL type ThreadConnection.
type ThreadConnection interface {
	Nodes(context.Context) ([]Thread, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
	Filters(context.Context) (ThreadConnectionFilters, error)
}
