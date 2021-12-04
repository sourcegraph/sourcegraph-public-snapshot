package resolvers

import (
	"log"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func parseQuery(db database.DB, q string) *queryMatcher {
	var match queryMatcher

	parts := strings.Fields(q)
	for _, part := range parts {
		const (
			relatedToEntityPrefix = "relatedToEntity:"
			excludeEntityPrefix   = "-entity:"
			groupPrefix           = "group:"
		)
		switch {
		case strings.HasPrefix(part, relatedToEntityPrefix):
			relatedToEntityID := graphql.ID(strings.TrimPrefix(part, relatedToEntityPrefix))
			match.relatedToEntity = entityByID(db, relatedToEntityID)

		case strings.HasPrefix(part, excludeEntityPrefix):
			match.excludeEntity = graphql.ID(strings.TrimPrefix(part, excludeEntityPrefix))

		case strings.HasPrefix(part, groupPrefix):
			groupID := graphql.ID(strings.TrimPrefix(part, groupPrefix))
			group := groupByID(db, groupID)
			match.groups = append(match.groups, group)
			for _, dg := range group.DescendentGroups() {
				match.groups = append(match.groups, dg)
			}

		default:
			match.literal += part
		}
	}

	return &match
}

type queryMatcher struct {
	literal         string
	relatedToEntity *catalogComponentResolver
	excludeEntity   graphql.ID
	groups          []gql.GroupResolver

	once                     sync.Once
	relatedEntityNamesCached []string
}

func (q *queryMatcher) relatedEntityNames() []string {
	q.once.Do(func() {
		_, _, edges := catalog.Data()
		for _, edge := range edges {
			var relatedEntityName string
			switch {
			case edge.In == q.relatedToEntity.component.Name:
				relatedEntityName = edge.Out
			case edge.Out == q.relatedToEntity.component.Name:
				relatedEntityName = edge.In
			default:
				continue
			}

			q.relatedEntityNamesCached = append(q.relatedEntityNamesCached, relatedEntityName)
		}
	})
	return q.relatedEntityNamesCached
}

func (q *queryMatcher) matchNode(c *catalogComponentResolver) bool {
	isLiteralMatch := strings.Contains(c.Name(), q.literal)

	isInAnyGroup := func(c *catalogComponentResolver, groups []gql.GroupResolver) bool {
		for _, g := range groups {
			if c.component.OwnedBy == g.Name() {
				return true
			}
		}
		return false
	}
	isRelatedToEntity := func(c *catalogComponentResolver) bool {
		if c.ID() == q.relatedToEntity.ID() {
			return true
		}
		for _, name := range q.relatedEntityNames() {
			if c.Name() == name {
				return true
			}
		}
		return false
	}

	return isLiteralMatch &&
		(q.relatedToEntity == nil || isRelatedToEntity(c)) &&
		(c.ID() != q.excludeEntity) &&
		(len(q.groups) == 0 || isInAnyGroup(c, q.groups))
}

func (q *queryMatcher) matchEdge(e *catalogEntityRelationEdgeResolver) bool {
	var id graphql.ID
	if q.relatedToEntity != nil {
		id = q.relatedToEntity.ID()
	}
	log.Printf("q.relatedToEntity = %v e.outNode.ID()=%v e.inNode.ID()=%v", id, e.outNode.ID(), e.inNode.ID())
	if q.relatedToEntity != nil && e.outNode.ID() != q.relatedToEntity.ID() && e.inNode.ID() != q.relatedToEntity.ID() {
		return false
	}

	return true
}
