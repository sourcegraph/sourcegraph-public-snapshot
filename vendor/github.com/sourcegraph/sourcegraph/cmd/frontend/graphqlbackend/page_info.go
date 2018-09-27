package graphqlbackend

type pageInfo struct {
	hasNextPage bool
}

func (r *pageInfo) HasNextPage() bool { return r.hasNextPage }
