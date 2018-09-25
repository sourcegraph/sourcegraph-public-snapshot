package graphqlbackend

import (
	"encoding/json"
)

// jsonValue implements the JSONValue scalar type. In GraphQL queries, it is represented the JSON
// representation of its Go value.
type jsonValue struct{ value interface{} }

func (jsonValue) ImplementsGraphQLType(name string) bool {
	return name == "JSONValue"
}

func (v *jsonValue) UnmarshalGraphQL(input interface{}) error {
	*v = jsonValue{value: input}
	return nil
}

func (v jsonValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}
