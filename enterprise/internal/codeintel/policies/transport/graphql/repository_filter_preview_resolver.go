package graphql

import (
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type repositoryFilterPreviewResolver struct {
	repositoryResolvers []resolverstubs.RepositoryResolver
	totalCount          int
	totalMatches        int
	matchesAllRepos     bool
	limit               *int
}

func NewRepositoryFilterPreviewResolver(repositoryResolvers []resolverstubs.RepositoryResolver, totalCount, totalMatches int, matchesAllRepos bool, limit *int) resolverstubs.RepositoryFilterPreviewResolver {
	return &repositoryFilterPreviewResolver{
		repositoryResolvers: repositoryResolvers,
		totalCount:          totalCount,
		totalMatches:        totalMatches,
		matchesAllRepos:     matchesAllRepos,
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

func (r *repositoryFilterPreviewResolver) MatchesAllRepos() bool {
	return r.matchesAllRepos
}

func (r *repositoryFilterPreviewResolver) Limit() *int32 {
	if r.limit == nil {
		return nil
	}

	v := int32(*r.limit)
	return &v
}
