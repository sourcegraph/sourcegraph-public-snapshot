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

func (schemaResolver) Threads(ctx context.Context, arg *ThreadConnectionArgs) (ThreadConnection, error) {
	if Threads == nil {
		return nil, errThreadsNotImplemented
	}
	return Threads.Threads(ctx, arg)
}

// ThreadsResolver is the interface for the GraphQL threads queries and mutations.
type ThreadsResolver interface {
	// Queries
	Threads(context.Context, *ThreadConnectionArgs) (ThreadConnection, error)

	// ThreadByID is called by the ThreadByID func but is not in the GraphQL API.
	ThreadByID(context.Context, graphql.ID) (Thread, error)
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

type ThreadState string

const (
	ThreadStateOpen   ThreadState = "OPEN"
	ThreadStateMerged ThreadState = "MERGED"
	ThreadStateClosed ThreadState = "CLOSED"
)

type ThreadKind string

const (
	ThreadKindIssue     ThreadKind = "ISSUE"
	ThreadKindChangeset ThreadKind = "CHANGESET"
)

// Thread is the interface for the GraphQL type Thread.
type Thread interface {
	ID() graphql.ID
	DBID() int64
	Repository(context.Context) (*RepositoryResolver, error)
	Internal_RepositoryID() api.RepoID
	Title() string
	State() ThreadState
	BaseRef() *string
	HeadRef() *string
	CreatedAt() DateTime
	UpdatedAt() DateTime
	Updatable
	Kind(context.Context) (ThreadKind, error)
	ExternalURLs(context.Context) ([]*externallink.Resolver, error)
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
