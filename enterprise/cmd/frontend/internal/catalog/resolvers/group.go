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

func (r *groupResolver) ChildGroups() []gql.GroupResolver {
	var childGroups []gql.GroupResolver
	for _, group := range allGroups(r.db) {
		if group.group.ParentGroup == r.group.Name {
			childGroups = append(childGroups, group)
		}
	}
	return childGroups
}

func (r *groupResolver) Members() []*gql.PersonResolver {
	var (
		members []*gql.PersonResolver
		seen    = map[string]struct{}{}
	)
	recordMember := func(member string) {
		if _, seen := seen[member]; seen {
			return
		}
		members = append(members, gql.NewPersonResolver(r.db, "", member+"@sourcegraph.com", false))
		seen[member] = struct{}{}
	}

	for _, member := range r.group.Members {
		recordMember(member)
	}
	for _, childGroup := range r.ChildGroups() {
		for _, member := range childGroup.(*groupResolver).group.Members {
			recordMember(member)
		}
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].Email() < members[j].Email()
	})

	return members
}
