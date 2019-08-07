package graphqlbackend

import "encoding/json"

// JSONValue implements the JSONValue scalar type. In GraphQL queries, it is represented the JSON
// representation of its Go value.
type JSONValue struct{ value interface{} }

func (JSONValue) ImplementsGraphQLType(name string) bool {
	return name == "JSONValue"
}

func (v *JSONValue) UnmarshalGraphQL(input interface{}) error {
	*v = JSONValue{value: input}
	return nil
}

func (v JSONValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}
