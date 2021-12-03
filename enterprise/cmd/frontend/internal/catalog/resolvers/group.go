package resolvers

import (
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
	if r.group.Title == "" {
		return nil
	}
	return &r.group.Title
}

func (r *groupResolver) URL() string { return "/catalog/groups/" + r.group.Name }

func (r *groupResolver) ParentGroup() gql.GroupResolver {
	return groupByName(r.db, r.group.ParentGroup)
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
	var members []*gql.PersonResolver
	for _, member := range r.group.Members {
		members = append(members, gql.NewPersonResolver(r.db, "", member+"@sourcegraph.com", false))
	}
	return members
}
