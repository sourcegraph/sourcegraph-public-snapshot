package resolvers

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func componentByID(db database.DB, id graphql.ID) *componentResolver {
	components := dummyComponents(db)
	for _, c := range components {
		if c.ID() == id {
			return c
		}
	}
	return nil
}
