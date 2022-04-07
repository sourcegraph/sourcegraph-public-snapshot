package graphql

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

const DefaultExecutorsLimit = 50

type params struct {
	query  string
	active bool
	offset int
	limit  int
}

func validateArgs(query *string, active *bool, first *int32, after *string) (p params, err error) {
	if query != nil {
		p.query = *query
	}

	if active != nil {
		p.active = *active
	}

	offset, err := graphqlutil.DecodeIntCursor(after)
	if err != nil {
		return
	}
	p.offset = offset

	limit := DefaultExecutorsLimit
	if first != nil {
		limit = int(*first)
	}
	p.limit = limit

	return
}
