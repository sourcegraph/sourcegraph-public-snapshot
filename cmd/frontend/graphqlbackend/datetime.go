package graphqlbackend

import (
	"encoding/json"
	"fmt"
	"time"
)

// DateTime implements the DateTime GraphQL scalar type.
type DateTime struct{ time.Time }

func (DateTime) ImplementsGraphQLType(name string) bool {
	return name == "DateTime"
}

func (v DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Time.Format(time.RFC3339))
}

func (v *DateTime) UnmarshalGraphQL(input interface{}) error {
	s, ok := input.(string)
	if !ok {
		return fmt.Errorf("invalid GraphQL DateTime scalar value input (got %T, expected string)", input)
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*v = DateTime{t}
	return nil
}
