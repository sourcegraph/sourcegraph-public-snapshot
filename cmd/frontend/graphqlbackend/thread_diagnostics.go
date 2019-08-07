package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

/*
   addDiagnosticsToThread(thread: ID!, rawDiagnostics: [String!]!): ThreadDiagnosticColleciton

   # Remove diagnostics from a thread.
   removeDiagnosticsFromThread(thread: ID!, threadDiagnosticEdges: [ID!]!): EmptyResponse

*/

// ThreadDiagnostics is the implementation of the GraphQL API for thread diagnostics queries and
// mutations. If it is not set at runtime, a "not implemented" error is returned to API clients who
// invoke it.
//
// This is contributed by enterprise.
var ThreadDiagnostics ThreadDiagnosticsResolver

const GQLTypeThreadDiagnosticEdge = "ThreadDiagnosticEdge"

func MarshalThreadDiagnosticEdgeID(id int64) graphql.ID {
	return relay.MarshalID(GQLTypeThreadDiagnosticEdge, id)
}

func UnmarshalThreadDiagnosticEdgeID(id graphql.ID) (dbID int64, err error) {
	if typ := relay.UnmarshalKind(id); typ != GQLTypeThreadDiagnosticEdge {
		return 0, fmt.Errorf("thread diagnostic edge ID has unexpected type type %q", typ)
	}
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

var errThreadDiagnosticsNotImplemented = errors.New("thread diagnostics is not implemented")

func (schemaResolver) ThreadsDiagnostics(ctx context.Context, arg *ThreadDiagnosticConnectionArgs) (ThreadDiagnosticConnection, error) {
	if ThreadDiagnostics == nil {
		return nil, errThreadDiagnosticsNotImplemented
	}
	return ThreadDiagnostics.ThreadDiagnostics(ctx, arg)
}

func (schemaResolver) ThreadsDiagnosticEdgeByID(ctx context.Context, id graphql.ID) (ThreadDiagnosticEdge, error) {
	if ThreadDiagnostics == nil {
		return nil, errThreadDiagnosticsNotImplemented
	}
	return ThreadDiagnostics.ThreadDiagnosticEdgeByID(ctx, id)
}

func (schemaResolver) AddDiagnosticsToThread(ctx context.Context, arg *AddDiagnosticsToThreadArgs) (*EmptyResponse, error) {
	if ThreadDiagnostics == nil {
		return nil, errThreadDiagnosticsNotImplemented
	}
	return ThreadDiagnostics.AddDiagnosticsToThread(ctx, arg)
}

func (schemaResolver) RemoveDiagnosticsFromThread(ctx context.Context, arg *RemoveDiagnosticsFromThreadArgs) (*EmptyResponse, error) {
	if ThreadDiagnostics == nil {
		return nil, errThreadDiagnosticsNotImplemented
	}
	return ThreadDiagnostics.RemoveDiagnosticsFromThread(ctx, arg)
}

// ThreadDiagnosticsResolver is the interface for the GraphQL threads diagnostics queries and
// mutations.
type ThreadDiagnosticsResolver interface {
	// Queries
	ThreadDiagnostics(context.Context, *ThreadDiagnosticConnectionArgs) (ThreadDiagnosticConnection, error)

	// Mutations
	AddDiagnosticsToThread(context.Context, *AddDiagnosticsToThreadArgs) (*EmptyResponse, error)
	RemoveDiagnosticsFromThread(context.Context, *RemoveDiagnosticsFromThreadArgs) (*EmptyResponse, error)

	// ThreadDiagnosticEdgeByID is called by the ThreadDiagnosticEdgeByID func but is not in the
	// GraphQL API.
	ThreadDiagnosticEdgeByID(context.Context, graphql.ID) (ThreadDiagnosticEdge, error)
}

type AddDiagnosticsToThreadArgs struct {
	Thread         graphql.ID
	RawDiagnostics []string
}

type RemoveDiagnosticsFromThreadArgs struct {
	Thread                graphql.ID
	ThreadDiagnosticEdges []graphql.ID
}

type ThreadDiagnosticEdge interface {
	ID() graphql.ID
	Thread(context.Context) (Thread, error)
	Diagnostic() (Diagnostic, error)
	updatable
}

type hasThreadDiagnostics interface {
	Diagnostics(context.Context, *ThreadDiagnosticConnectionArgs) (ThreadDiagnosticConnection, error)
}

type ThreadDiagnosticConnectionArgs struct {
	graphqlutil.ConnectionArgs
	Thread   *graphql.ID
	Campaign *graphql.ID
}

// ThreadDiagnosticConnection is the interface for the GraphQL type ThreadDiagnosticConnection.
type ThreadDiagnosticConnection interface {
	Edges(context.Context) ([]ThreadDiagnosticEdge, error)
	Nodes(context.Context) ([]Diagnostic, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
