package resolvers_test

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/search/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestValidateExhaustiveSearchQueryResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	resolver := resolvers.New(logger, db)
	s, err := graphqlbackend.NewSchemaWithExhaustiveSearchesResolver(db, resolver)
	require.NoError(t, err)

	variables := map[string]any{
		"query": "test",
	}

	query := `query($query: String!) {
	validateExhaustiveSearchQuery(query: $query) {
		query
		valid
		errors
	}
}`

	var actual string
	errors := exec(ctx, t, s, query, variables, &actual)
	require.Equal(t, 1, len(errors))
	assert.Equal(t, errors[0].Message, "panic occurred: implement me")
}
