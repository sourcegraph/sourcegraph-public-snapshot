package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Diagnostic implements the Diagnostic GraphQL type.
type Diagnostic interface {
	Type() string
	Data() JSONValue
}

// DiagnosticConnection is the interface for the GraphQL type DiagnosticConnection.
type DiagnosticConnection interface {
	Nodes(context.Context) ([]Diagnostic, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
