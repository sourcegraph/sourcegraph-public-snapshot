package resolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// TODO(sqs): dummy data
func dummyData(db database.DB) []*catalogComponentResolver {
	components := catalog.Components()
	resolvers := make([]*catalogComponentResolver, len(components))
	for i, c := range components {
		resolvers[i] = &catalogComponentResolver{
			component: c,
			db:        db,
		}
	}
	return resolvers
}
