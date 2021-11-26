package resolvers

import (
	"context"
	"sort"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func toTagResolvers(db database.DB, tagSet map[string]struct{}) []gql.ComponentTagResolver {
	allTags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		allTags = append(allTags, tag)
	}
	sort.Strings(allTags)

	rs := make([]gql.ComponentTagResolver, len(allTags))
	for i, tag := range allTags {
		rs[i] = &componentTagResolver{tagName: tag, db: db}
	}
	return rs
}

func (r *rootResolver) ComponentTags(ctx context.Context) ([]gql.ComponentTagResolver, error) {
	tagSet := map[string]struct{}{}
	components := catalog.Components()
	for _, c := range components {
		for _, tag := range c.Tags {
			tagSet[tag] = struct{}{}
		}
	}
	return toTagResolvers(r.db, tagSet), nil
}

func (r *componentResolver) Tags(ctx context.Context) ([]gql.ComponentTagResolver, error) {
	tagSet := map[string]struct{}{}
	for _, tag := range r.component.Tags {
		tagSet[tag] = struct{}{}
	}
	return toTagResolvers(r.db, tagSet), nil
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

	components := catalog.Components()
	keep := components[:0]
	for _, c := range components {
		if hasTag(c, r.tagName) {
			keep = append(keep, c)
		}
	}
	components = keep

	return &componentConnectionResolver{
		components: componentResolversGQLIface(r.db, components),
		first:      args.First,
		db:         r.db,
	}, nil
}
