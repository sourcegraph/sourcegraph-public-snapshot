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

func TestSearchJobResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	resolver := resolvers.New(logger, db)
	s, err := graphqlbackend.NewSchemaWithSearchJobsResolver(db, resolver)
	require.NoError(t, err)

	variables := map[string]any{
		"searchJob": string(resolvers.MarshalSearchJobID(int64(123))),
	}

	query := `query($searchJob: ID!) {
	node(id: $searchJob) {
		... on SearchJob {
			id
			query
			state
			creator {
				username
			}
			createdAt
			startedAt
			finishedAt
			csvURL
			repoStats {
				total
				completed
				failed
				inProgress
			}
			repositories(first: 1) {
				totalCount
				pageInfo {
					hasNextPage
					endCursor
				}
				nodes {
					id
				}
			}
		}
	}
}`

	var actual string
	errors := exec(ctx, t, s, query, variables, &actual)
	require.Equal(t, 10, len(errors))
	assert.Equal(t, errors[0].Message, "panic occurred: implement me")
}
