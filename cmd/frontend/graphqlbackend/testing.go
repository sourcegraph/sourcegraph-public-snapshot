package graphqlbackend

import (
	"database/sql"
	"sync"
	"testing"

	graphql "github.com/graph-gophers/graphql-go"
)

var (
	parseSchemaOnce sync.Once
	parseSchemaErr  error
	parsedSchema    *graphql.Schema
)

func mustParseGraphQLSchema(t *testing.T, db *sql.DB) *graphql.Schema {
	t.Helper()

	parseSchemaOnce.Do(func() {
		parsedSchema, parseSchemaErr = NewSchema(nil, nil, nil)
	})
	if parseSchemaErr != nil {
		t.Fatal(parseSchemaErr)
	}

	return parsedSchema
}
