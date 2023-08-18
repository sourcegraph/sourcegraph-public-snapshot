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

func TestSearchJobsResolver_CreateSearchJob(t *testing.T) {
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
		"query": "test",
	}

	query := `mutation($query: String!) {
	createSearchJob(query: $query) {
		id
	}
}`

	var actual string
	errors := exec(ctx, t, s, query, variables, &actual)
	require.Equal(t, 1, len(errors))
	assert.Equal(t, errors[0].Message, "panic occurred: implement me")
}

func TestSearchJobsResolver_CancelSearchJob(t *testing.T) {
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

	query := `mutation($searchJob: ID!) {
	cancelSearchJob(id: $searchJob) { alwaysNil }
}`

	var actual string
	errors := exec(ctx, t, s, query, variables, &actual)
	require.Equal(t, 1, len(errors))
	assert.Equal(t, errors[0].Message, "panic occurred: implement me")
}

func TestSearchJobsResolver_DeleteSearchJob(t *testing.T) {
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

	query := `mutation($searchJob: ID!) {
	deleteSearchJob(id: $searchJob) { alwaysNil }
}`

	var actual string
	errors := exec(ctx, t, s, query, variables, &actual)
	require.Equal(t, 1, len(errors))
	assert.Equal(t, errors[0].Message, "panic occurred: implement me")
}

func TestSearchJobsResolver_RetrySearchJob(t *testing.T) {
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

	query := `mutation($searchJob: ID!) {
	retrySearchJob(id: $searchJob) {
		id
	}
}`

	var actual string
	errors := exec(ctx, t, s, query, variables, &actual)
	require.Equal(t, 1, len(errors))
	assert.Equal(t, errors[0].Message, "panic occurred: implement me")
}

func TestSearchJobsResolver_ValidateSearchJobQuery(t *testing.T) {
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
		"query": "test",
	}

	query := `query($query: String!) {
	validateSearchJobQuery(query: $query) {
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

func TestSearchJobsResolver_SearchJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	resolver := resolvers.New(logger, db)
	s, err := graphqlbackend.NewSchemaWithSearchJobsResolver(db, resolver)
	require.NoError(t, err)

	variables := map[string]any{}

	query := `query {
	searchJobs(first: 1) {
		totalCount
		pageInfo {
			hasNextPage
			endCursor
		}
		nodes {
			id
		}
	}
}`

	var actual string
	errors := exec(ctx, t, s, query, variables, &actual)
	require.Equal(t, 1, len(errors))
	assert.Equal(t, errors[0].Message, "panic occurred: implement me")
}
