package graphql

import (
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
)

type RepositoryFilterPreviewResolver interface {
	Nodes() []*sharedresolvers.RepositoryResolver
	TotalCount() int32
	Limit() *int32
	TotalMatches() int32
	PageInfo() *PageInfo
}

type repositoryFilterPreviewResolver struct {
	repositoryResolvers []*sharedresolvers.RepositoryResolver
	totalCount          int
	offset              int
	totalMatches        int
	limit               *int
}

func NewRepositoryFilterPreviewResolver(repositoryResolvers []*sharedresolvers.RepositoryResolver, totalCount, offset, totalMatches int, limit *int) RepositoryFilterPreviewResolver {
	return &repositoryFilterPreviewResolver{
		repositoryResolvers: repositoryResolvers,
		totalCount:          totalCount,
		offset:              offset,
		totalMatches:        totalMatches,
		limit:               limit,
	}
}

func (r *repositoryFilterPreviewResolver) Nodes() []*sharedresolvers.RepositoryResolver {
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

func (r *repositoryFilterPreviewResolver) PageInfo() *PageInfo {
	return EncodeIntCursor(toInt32(NextOffset(r.offset, len(r.repositoryResolvers), r.totalCount)))
}
