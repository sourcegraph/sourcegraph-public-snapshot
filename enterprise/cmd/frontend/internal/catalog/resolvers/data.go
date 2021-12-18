package resolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// TODO(sqs): dummy data
func dummyComponents(db database.DB) []*componentResolver {
	components, _, _ := catalog.Data()
	resolvers := make([]*componentResolver, len(components))
	for i, c := range components {
		resolvers[i] = &componentResolver{
			component: c,
			db:        db,
		}
	}
	return resolvers
}
