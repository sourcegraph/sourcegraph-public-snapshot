package graphqlbackend

// PageInfo implements the GraphQL type PageInfo.
type PageInfo struct {
	hasNextPage bool
}

// NewPageInfo returns a new PageInfo.
func NewPageInfo(hasNextPage bool) *PageInfo { return &PageInfo{hasNextPage: hasNextPage} }

func (r *PageInfo) HasNextPage() bool { return r.hasNextPage }
