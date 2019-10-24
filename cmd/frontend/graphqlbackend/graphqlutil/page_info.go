package graphqlutil

import graphql "github.com/graph-gophers/graphql-go"

// PageInfo implements the GraphQL type PageInfo.
type PageInfo struct {
	endCursor   *graphql.ID
	hasNextPage bool
}

// HasNextPage returns a new PageInfo with the given hasNextPage value.
func HasNextPage(hasNextPage bool) *PageInfo {
	return &PageInfo{hasNextPage: hasNextPage}
}

// NextPageCursor returns a new PageInfo indicating there is a next page with
// the given end cursor.
func NextPageCursor(endCursor graphql.ID) *PageInfo {
	return &PageInfo{endCursor: &endCursor, hasNextPage: true}
}

func (r *PageInfo) EndCursor() *graphql.ID { return r.endCursor }
func (r *PageInfo) HasNextPage() bool      { return r.hasNextPage }
