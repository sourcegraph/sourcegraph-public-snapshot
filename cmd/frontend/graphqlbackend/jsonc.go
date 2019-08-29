package graphqlbackend

import (
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
)

// JSONC implements the JSONC GraphQL type. Its value is a JSONC document, with comments and
// trailing commas allowed.
type JSONC string

func (JSONC) ImplementsGraphQLType(name string) bool {
	return name == "JSONC"
}

func (v JSONC) Raw() JSONCString { return JSONCString(v) }

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

// JSONCString implements the JSONCString GraphQL input type.
type JSONCString string

func (JSONCString) ImplementsGraphQLType(name string) bool {
	return name == "JSONCString"
}

func (v JSONCString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(v))
}

func (v *JSONCString) UnmarshalGraphQL(input interface{}) error {
	s, ok := input.(string)
	if !ok {
		return fmt.Errorf("invalid GraphQL JSONCString scalar value input (got %T, expected string)", input)
	}
	*v = JSONCString(s)
	return nil
}
