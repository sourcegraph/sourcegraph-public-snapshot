package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type componentConnectionResolver struct {
	components []gql.ComponentResolver
	first      *int32

	db database.DB
}

func (r *componentConnectionResolver) Nodes(ctx context.Context) ([]gql.ComponentResolver, error) {
	var nodes []gql.ComponentResolver
	if r.first != nil && len(r.components) > int(*r.first) {
		nodes = r.components[:int(*r.first)]
	} else {
		nodes = r.components
	}
	return nodes, nil
}

func (r *componentConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.components)), nil
}

func (r *componentConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(r.first != nil && int(*r.first) < len(r.components)), nil
}

func (r *componentConnectionResolver) Tags(ctx context.Context) ([]gql.ComponentTagResolver, error) {
	// TODO(sqs): this should return the set of tags that are present on all components in the
	// connection plus those that would be if not for any existing tag filters.
	tagSet := map[string]struct{}{}
	for _, c := range r.components {
		for _, tag := range c.(*componentResolver).component.Tags {
			tagSet[tag] = struct{}{}
		}
	}
	return toTagResolvers(r.db, tagSet), nil
}
