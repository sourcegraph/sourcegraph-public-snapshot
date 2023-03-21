package gqlutil

import (
	"encoding/json"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DateTime implements the DateTime GraphQL scalar type.
type DateTime struct{ time.Time }

// DateTimeOrNil is a helper function that returns nil for time == nil and otherwise wraps time in
// DateTime.
func DateTimeOrNil(timePtr *time.Time) *DateTime {
	if timePtr == nil {
		return nil
	}
	return &DateTime{Time: *timePtr}
}

// FromTime is a helper function that returns nil for a zero-valued time and
// otherwise wraps time in DateTime.
func FromTime(inputTime time.Time) *DateTime {
	if inputTime.IsZero() {
		return nil
	}
	return &DateTime{Time: inputTime}
}

func (DateTime) ImplementsGraphQLType(name string) bool {
	return name == "DateTime"
}

func (v DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Time.UTC().Format(time.RFC3339))
}

func (v *DateTime) UnmarshalGraphQL(input any) error {
	s, ok := input.(string)
	if !ok {
		return errors.Errorf("invalid GraphQL DateTime scalar value input (got %T, expected string)", input)
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*v = DateTime{Time: t.UTC()}
	return nil
}
