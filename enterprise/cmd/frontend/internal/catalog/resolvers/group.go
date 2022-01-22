package resolvers

import (
	"sort"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type groupResolver struct {
	group catalog.Group

	db database.DB
}

func (r *groupResolver) ID() graphql.ID { return relay.MarshalID("Group", r.group.Name) }
func (r *groupResolver) Name() string   { return r.group.Name }
func (r *groupResolver) Title() string  { return r.group.Title }

func (r *groupResolver) Description() *string {
	if r.group.Description == "" {
		return nil
	}
	return &r.group.Description
}

func (r *groupResolver) URL() string { return "/catalog/groups/" + r.group.Name }

func (r *groupResolver) ParentGroup() gql.GroupResolver {
	g := groupByName(r.db, r.group.ParentGroup)
	if g != nil {
		return g
	}
	return nil
}

func (r *groupResolver) AncestorGroups() []gql.GroupResolver {
	var ancestors []gql.GroupResolver

	cur := r
	for {
		cur = groupByName(r.db, cur.group.ParentGroup)
		if cur == nil {
			break
		}
		ancestors = append(ancestors, cur)
	}
	for i, j := 0, len(ancestors)-1; i < j; i, j = i+1, j-1 {
		ancestors[i], ancestors[j] = ancestors[j], ancestors[i]
	}
	return ancestors
}

func (r *groupResolver) ChildGroups() []gql.GroupResolver {
	var childGroups []gql.GroupResolver
	for _, group := range allGroups(r.db) {
		if group.group.ParentGroup == r.group.Name {
			childGroups = append(childGroups, group)
		}
	}
	return childGroups
}

func (r *groupResolver) DescendentGroups() []gql.GroupResolver {
	var descendentGroups []gql.GroupResolver
	for _, group := range r.ChildGroups() {
		descendentGroups = append(descendentGroups, group)
		descendentGroups = append(descendentGroups, group.DescendentGroups()...)
	}
	return descendentGroups
}

func (r *groupResolver) Members() []*gql.PersonResolver {
	var (
		members []*gql.PersonResolver
		seen    = map[string]struct{}{}
	)
	for _, member := range r.group.Members {
		email := member + "@sourcegraph.com"
		if _, seen := seen[email]; seen {
			continue
		}
		members = append(members, gql.NewPersonResolver(r.db, "", email, false))
		seen[email] = struct{}{}
	}
	for _, childGroup := range r.ChildGroups() {
		for _, member := range childGroup.Members() {
			if _, seen := seen[member.Email()]; seen {
				continue
			}
			members = append(members, member)
			seen[member.Email()] = struct{}{}
		}
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].Email() < members[j].Email()
	})

	return members
}

func (r *groupResolver) Components() []gql.ComponentResolver {
	var entities []gql.ComponentResolver

	for _, c := range dummyComponents(r.db) {
		if c.component.OwnedBy == r.group.Name {
			entities = append(entities, c)
		}
	}
	for _, childGroup := range r.ChildGroups() {
		entities = append(entities, childGroup.Components()...)
	}

	return entities
}
