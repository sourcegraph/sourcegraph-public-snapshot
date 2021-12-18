package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func (r *componentResolver) Tags(ctx context.Context) ([]gql.ComponentTagResolver, error) {
	tagSet := map[string]struct{}{}
	components, _, _ := catalog.Data()
	for _, c := range components {
		for _, tag := range c.Tags {
			tagSet[tag] = struct{}{}
		}
	}

	allTags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		allTags = append(allTags, tag)
	}

	rs := make([]gql.ComponentTagResolver, len(allTags))
	for i, tag := range allTags {
		rs[i] = &componentTagResolver{tagName: tag, db: r.db}
	}
	return rs, nil
}

type componentTagResolver struct {
	tagName string

	db database.DB
}

func (r *componentTagResolver) Name() string { return r.tagName }

func (r *componentTagResolver) Components(ctx context.Context, args *gql.ComponentTagComponentsArgs) (gql.ComponentConnectionResolver, error) {
	hasTag := func(c catalog.Component, tag string) bool {
		for _, t := range c.Tags {
			if t == tag {
				return true
			}
		}
		return false
	}

	components, _, _ := catalog.Data()
	keep := components[:0]
	for _, c := range components {
		if hasTag(c, r.tagName) {
			keep = append(keep, c)
		}
	}
	components = keep

	return &componentConnectionResolver{components: componentResolversGQLIface(r.db, components)}, nil
}
