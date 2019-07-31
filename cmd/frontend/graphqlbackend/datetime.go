package graphqlbackend

import (
	"time"
)

// DateTime implements the DateTime GraphQL scalar type.
type DateTime struct{ time.Time }

func (DateTime) ImplementsGraphQLType(name string) bool {
	return name == "DateTime"
}
