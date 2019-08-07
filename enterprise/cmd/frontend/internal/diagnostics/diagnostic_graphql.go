package diagnostics

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// gqlDiagnostic implements the GraphQL Diagnostic type.
type gqlDiagnostic struct {
	typ  string          // diagnostic type
	data json.RawMessage // diagnostic data (conforms to sourcegraph.Diagnostic extension API type)
}

func NewGQLDiagnostic(typ string, data json.RawMessage) graphqlbackend.Diagnostic {
	return &gqlDiagnostic{typ: typ, data: data}
}

func (v *gqlDiagnostic) Type() string                   { return v.typ }
func (v *gqlDiagnostic) Data() graphqlbackend.JSONValue { return graphqlbackend.JSONValue{v.data} }
