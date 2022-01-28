package graphqlbackend

import "github.com/sourcegraph/sourcegraph/internal/search"

// searchQueryDescription is a type for the SearchQueryDescription resolver used
// by SearchAlert. This name is a bit of a misnomer but cannot be changed: It
// must be this way to work with the GQL definition and compatibility. We use
// our internal, resolver-agnostic alert type to do real work.
type searchQueryDescription struct {
	query *search.ProposedQuery
}

func (q searchQueryDescription) Query() string {
	// Do not add logic here that manipulates the query string. Do it in the QueryString() method.
	return q.query.QueryString()
}

func (q searchQueryDescription) Description() *string {
	if q.query.Description == "" {
		return nil
	}

	return &q.query.Description
}
