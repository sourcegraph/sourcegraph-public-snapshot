package graphqlbackend

import (
	"encoding/json"
	"fmt"
	"time"
)

// DateTime implements the DateTime GraphQL scalar type.
type DateTime struct{ time.Time }

// TimeOrNil returns nil if d == nil and otherwise a pointer to d's time.
func (d *DateTime) TimeOrNil() *time.Time {
	if d == nil {
		return nil
	}
	return &d.Time
}

// Equal reports whether d is equal to other.
func (d *DateTime) Equal(other *DateTime) bool {
	return (d == nil && other == nil) || (d != nil && other != nil && d.Time.Equal(other.Time))
}

// DateTimeOrNil is a helper function that returns nil for time == nil and otherwise wraps time in
// DateTime.
func DateTimeOrNil(time *time.Time) *DateTime {
	if time == nil {
		return nil
	}
	return &DateTime{Time: *time}
}

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
	*v = DateTime{Time: t}
	return nil
}
