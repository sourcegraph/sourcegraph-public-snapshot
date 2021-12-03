package resolvers

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
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
