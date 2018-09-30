package graphqlutil

// PageInfo implements the GraphQL type PageInfo.
type PageInfo struct {
	hasNextPage bool
}

// HasNextPage returns a new PageInfo with the given hasNextPage value.
func HasNextPage(hasNextPage bool) *PageInfo {
	return &PageInfo{hasNextPage: hasNextPage}
}

func (r *PageInfo) HasNextPage() bool { return r.hasNextPage }
