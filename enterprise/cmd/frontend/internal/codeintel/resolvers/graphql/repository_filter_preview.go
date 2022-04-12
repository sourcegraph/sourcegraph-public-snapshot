package graphql

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type repositoryFilterPreviewResolver struct {
	repositoryResolvers []*gql.RepositoryResolver
	totalCount          int
	offset              int
	totalMatches        int
	limit               *int
}

var _ gql.RepositoryFilterPreviewResolver = &repositoryFilterPreviewResolver{}

func (r *repositoryFilterPreviewResolver) Nodes() []*gql.RepositoryResolver {
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

func (r *repositoryFilterPreviewResolver) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.EncodeIntCursor(toInt32(graphqlutil.NextOffset(r.offset, len(r.repositoryResolvers), r.totalCount)))
}
