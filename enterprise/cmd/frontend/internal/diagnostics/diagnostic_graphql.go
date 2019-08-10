package diagnostics

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// GQLDiagnostic implements the GraphQL Diagnostic type.
type GQLDiagnostic struct {
	Type_ string          `json:"type"` // diagnostic type
	Data_ json.RawMessage `json:"data"` // diagnostic data (conforms to sourcegraph.Diagnostic extension API type)
}

func (v GQLDiagnostic) Type() string                   { return v.Type_ }
func (v GQLDiagnostic) Data() graphqlbackend.JSONValue { return graphqlbackend.JSONValue{v.Data_} }
