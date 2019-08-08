package graphqlbackend

import "github.com/sourcegraph/sourcegraph/pkg/jsonc"

// JSONC implements the JSONC GraphQL type. Its value is a JSONC document, with comments and
// trailing commas allowed.
type JSONC string

func (JSONC) ImplementsGraphQLType(name string) bool {
	return name == "JSONC"
}

func (v JSONC) Raw() string { return string(v) }

func (v JSONC) Formatted() (string, error) {
	return jsonc.Format(string(v), nil)
}

func (v JSONC) Parsed() (*JSONValue, error) {
	var j JSONValue
	if err := jsonc.Unmarshal(string(v), &j.Value); err != nil {
		return nil, err
	}
	return &j, nil
}
