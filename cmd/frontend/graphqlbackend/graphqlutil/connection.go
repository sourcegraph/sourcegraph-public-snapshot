package graphqlutil

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/db"
)

// ConnectionArgs is the common set of arguments to GraphQL fields that return connections (lists).
type ConnectionArgs struct {
	First *int32 // return the first n items
}

// Set is a convenience method for setting the DB limit and offset in a DB XyzListOptions struct.
func (a ConnectionArgs) Set(o **db.LimitOffset) {
	if a.First != nil {
		*o = &db.LimitOffset{Limit: int(*a.First)}
	}
}

// GetFirst is a convenience method returning the value of First, defaulting to
// the type's zero value if nil.
func (a ConnectionArgs) GetFirst() int32 {
	if a.First == nil {
		return 0
	}
	return *a.First
}

func (a ConnectionArgs) GetFirstMin(min int32) (int32, error) {
	first := a.GetFirst()
	if first < min {
		return first, fmt.Errorf("invalid first argument given, min=%d", min)
	}
	return first, nil
}

func (a ConnectionArgs) GetFirstMinMax(min, max int32) (int32, error) {
	first := a.GetFirst()
	if first < min || first > max {
		return first, fmt.Errorf("invalid first argument given, min=%d, max=%d", min, max)
	}
	return first, nil
}
