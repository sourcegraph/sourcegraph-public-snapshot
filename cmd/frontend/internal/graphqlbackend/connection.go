package graphqlbackend

import (
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// connectionArgs is the common set of args to fields that return connections (lists).
type connectionArgs struct {
	First *int32 // return the first n items
}

func (a connectionArgs) set(o *sourcegraph.ListOptions) {
	if a.First != nil {
		o.PerPage = *a.First
	}
}
