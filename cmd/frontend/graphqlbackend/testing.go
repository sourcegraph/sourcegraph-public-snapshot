package graphqlbackend

import (
	"database/sql"
	"testing"

	graphql "github.com/graph-gophers/graphql-go"
)

func mustParseGraphQLSchema(t *testing.T, db *sql.DB) *graphql.Schema {
	t.Helper()

	schema, err := NewSchema(nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	return schema
}
