package resolvers

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func packageByID(db database.DB, id graphql.ID) *packageResolver {
	for _, pkg := range catalog.AllPackages() {
		pr := &packageResolver{db: db, pkg: pkg}
		if pr.ID() == id {
			return pr
		}
	}
	return nil
}
