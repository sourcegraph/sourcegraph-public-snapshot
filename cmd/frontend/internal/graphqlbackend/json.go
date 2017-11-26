package graphqlbackend

import (
	"encoding/json"
	"errors"
	"fmt"
)

// jsonString implements the JSONString scalar type. In GraphQL queries, it is
// represented as a string containing the JSON representation of its Go
// value.
type jsonString string

func (jsonString) ImplementsGraphQLType(name string) bool {
	return name == "JSONString"
}

func (v *jsonString) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case string:
		// Validate.
		if !json.Valid([]byte(input)) {
			return errors.New("invalid JSON")
		}
		*v = jsonString(input)
		return nil
	default:
		return fmt.Errorf("wrong type")
	}
}
