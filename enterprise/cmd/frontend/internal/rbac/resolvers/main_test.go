package resolvers

import (
	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func newSchema(db database.DB, r gql.RBACResolver) (*graphql.Schema, error) {
	return gql.MewSchemaWithRBACResolver(db, r)
}
