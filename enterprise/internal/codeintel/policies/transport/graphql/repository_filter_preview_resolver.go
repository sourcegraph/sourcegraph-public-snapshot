package graphql

import (
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type repositoryFilterPreviewResolver struct {
	repositoryResolvers []resolverstubs.RepositoryResolver
	totalCount          int
	offset              int
	totalMatches        int
	limit               *int
}

func NewRepositoryFilterPreviewResolver(repositoryResolvers []resolverstubs.RepositoryResolver, totalCount, offset, totalMatches int, limit *int) resolverstubs.RepositoryFilterPreviewResolver {
	return &repositoryFilterPreviewResolver{
		repositoryResolvers: repositoryResolvers,
		totalCount:          totalCount,
		offset:              offset,
		totalMatches:        totalMatches,
		limit:               limit,
	}
}

func (r *repositoryFilterPreviewResolver) Nodes() []resolverstubs.RepositoryResolver {
	return r.repositoryResolvers
}

func (r *repositoryFilterPreviewResolver) TotalCount() int32 {
	return int32(r.totalCount)
}

func (r *repositoryFilterPreviewResolver) TotalMatches() int32 {
	return int32(r.totalMatches)
}

func (r *repositoryFilterPreviewResolver) Limit() *int32 {
	if r.limit == nil {
		return nil
	}

	v := int32(*r.limit)
	return &v
}

func (r *repositoryFilterPreviewResolver) PageInfo() resolverstubs.PageInfo {
	return EncodeIntCursor(toInt32(NextOffset(r.offset, len(r.repositoryResolvers), r.totalCount)))
}
