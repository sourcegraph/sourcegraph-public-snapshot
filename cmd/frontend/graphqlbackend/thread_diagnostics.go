package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

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

func (schemaResolver) ThreadDiagnostics(ctx context.Context, arg *ThreadDiagnosticConnectionArgs) (ThreadDiagnosticConnection, error) {
	if ThreadDiagnostics == nil {
		return nil, errThreadDiagnosticsNotImplemented
	}
	return ThreadDiagnostics.ThreadDiagnostics(ctx, arg)
}

// ThreadDiagnosticEdgeByID is called to look up a ThreadDiagnosticEdge given its GraphQL ID.
func ThreadDiagnosticEdgeByID(ctx context.Context, id graphql.ID) (ThreadDiagnosticEdge, error) {
	if Threads == nil {
		return nil, errThreadDiagnosticsNotImplemented
	}
	return ThreadDiagnostics.ThreadDiagnosticEdgeByID(ctx, id)
}

func (schemaResolver) AddDiagnosticsToThread(ctx context.Context, arg *AddDiagnosticsToThreadArgs) (ThreadDiagnosticConnection, error) {
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
	AddDiagnosticsToThread(context.Context, *AddDiagnosticsToThreadArgs) (ThreadDiagnosticConnection, error)
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
	ID() *graphql.ID
	Thread(context.Context) (*ToThreadOrThreadPreview, error)
	Diagnostic() (Diagnostic, error)
	Updatable
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

type AddRemoveDiagnosticToFromThreadEvent struct {
	EventCommon
	Edge_       ThreadDiagnosticEdge // only set for AddDiagnosticToThreadEvent
	Thread_     Thread
	Diagnostic_ Diagnostic
}

func (v AddRemoveDiagnosticToFromThreadEvent) Edge() ThreadDiagnosticEdge { return v.Edge_ }
func (v AddRemoveDiagnosticToFromThreadEvent) Thread() Thread             { return v.Thread_ }
func (v AddRemoveDiagnosticToFromThreadEvent) Diagnostic() Diagnostic     { return v.Diagnostic_ }
