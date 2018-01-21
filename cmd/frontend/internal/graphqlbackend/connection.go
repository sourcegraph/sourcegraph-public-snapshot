package graphqlbackend

import "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"

const maxFirst = 250 // maximum number of items to return from a query using connectionArgs

// connectionArgs is the common set of args to fields that return connections (lists).
type connectionArgs struct {
	First *int32 // return the first n items
}

func (a connectionArgs) set(o **db.LimitOffset) {
	if a.First == nil || *a.First > maxFirst {
		n := int32(maxFirst)
		a.First = &n
	}

	*o = &db.LimitOffset{Limit: int(*a.First)}
}
