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

func TestExhaustiveSearchRepoResolver(t *testing.T) {
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
		"repoSearchID": string(resolvers.MarshalExhaustiveSearchRepoID(int64(123))),
	}

	query := `query($repoSearchID: ID!) {
	node(id: $repoSearchID) {
		... on ExhaustiveSearchRepo {
			id
			state
			repository {
				id
			}
			createdAt
			startedAt
			finishedAt
			failureMessage
			revisions(first: 1) {
				nodes {
					id
				}
			}
		}
	}
}`

	var actual string
	errors := exec(ctx, t, s, query, variables, &actual)
	require.Equal(t, 8, len(errors))
	assert.Equal(t, errors[0].Message, "panic occurred: implement me")
}
