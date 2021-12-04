package resolvers

import (
	"strings"

	"github.com/graph-gophers/graphql-go"
)

func getQueryMatcher(q string) func(*catalogComponentResolver) bool {
	parseQuery := func(q string) (literal string, entityID, groupID graphql.ID) {
		parts := strings.Fields(q)
		for _, part := range parts {
			const (
				entityPrefix = "entity:"
				groupPrefix  = "group:"
			)
			if strings.HasPrefix(part, entityPrefix) {
				entityID = graphql.ID(strings.TrimPrefix(part, entityPrefix))
			} else if strings.HasPrefix(part, groupPrefix) {
				groupID = graphql.ID(strings.TrimPrefix(part, groupPrefix))
			} else {
				literal += part
			}
		}
		return
	}

	literal, entityID, groupID := parseQuery(q)

	isEntityMatch := func(c *catalogComponentResolver) bool {
		return c.ID() == entityID
	}

	group := groupByID(groupID)
	isComponentInGroup := func(c *catalogComponentResolver) bool {
		if c.component.OwnedBy == group.group.Name {
			return true
		}
		for _, dg := range group.DescendentGroups() {
			if c.component.OwnedBy == dg.Name() {
				return true
			}
		}
		return false
	}

	return func(c *catalogComponentResolver) bool {
		return strings.Contains(c.component.Name, literal) && (entityID == "" || isEntityMatch(c)) && (group == nil || isComponentInGroup(c))
	}
}
