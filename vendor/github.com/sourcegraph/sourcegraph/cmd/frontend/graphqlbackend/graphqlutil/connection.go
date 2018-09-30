package graphqlutil

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

// graphqlutil.ConnectionArgs is the common set of arguments to GraphQL fields that return connections (lists).
type ConnectionArgs struct {
	First *int32 // return the first n items
}

// Set is a convenience method for setting the DB limit and offset in a DB XyzListOptions struct.
func (a ConnectionArgs) Set(o **db.LimitOffset) {
	if a.First != nil {
		*o = &db.LimitOffset{Limit: int(*a.First)}
	}
}
