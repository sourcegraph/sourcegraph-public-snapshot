package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

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
}

// ThreadDiagnosticConnection is the interface for the GraphQL type ThreadDiagnosticConnection.
type ThreadDiagnosticConnection interface {
	Edges(context.Context) ([]ThreadDiagnosticEdge, error)
	Nodes(context.Context) ([]Diagnostic, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
