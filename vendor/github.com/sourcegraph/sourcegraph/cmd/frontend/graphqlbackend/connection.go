package graphqlbackend

import "github.com/sourcegraph/sourcegraph/cmd/frontend/db"

// connectionArgs is the common set of args to fields that return connections (lists).
type connectionArgs struct {
	First *int32 // return the first n items
}

func (a connectionArgs) set(o **db.LimitOffset) {
	if a.First != nil {
		*o = &db.LimitOffset{Limit: int(*a.First)}
	}
}
