package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const cursorKind = "SavedSearchCursor"

// marshalCursor marshals a saved searches pagination cursor.
func marshalCursor(cursor types.MultiCursor) string {
	return string(relay.MarshalID(cursorKind, cursor))
}

// unmarshalCursor unmarshals a saved searches pagination cursor.
func unmarshalCursor(cursor string) (types.MultiCursor, error) {
	if kind := relay.UnmarshalKind(graphql.ID(cursor)); kind != cursorKind {
		return nil, errors.Errorf("invalid cursor type %q (expected %q)", kind, cursorKind)
	}
	var spec types.MultiCursor
	if err := relay.UnmarshalSpec(graphql.ID(cursor), &spec); err != nil {
		return nil, err
	}
	return spec, nil
}

type savedSearchesConnectionStore struct {
	db       database.DB
	listArgs database.SavedSearchListArgs
}

func (s *savedSearchesConnectionStore) MarshalCursor(node graphqlbackend.SavedSearchResolver, orderBy database.OrderBy) (*string, error) {
	cursor, err := s.db.SavedSearches().MarshalToCursor(&node.(*savedSearchResolver).s, orderBy)
	if err != nil {
		return nil, err
	}
	cursorStr := marshalCursor(cursor)
	return &cursorStr, nil
}

func (s *savedSearchesConnectionStore) UnmarshalCursor(cursorStr string, orderBy database.OrderBy) ([]any, error) {
	cursor, err := unmarshalCursor(cursorStr)
	if err != nil {
		return nil, err
	}
	return s.db.SavedSearches().UnmarshalValuesFromCursor(cursor)
}

func (s *savedSearchesConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	count, err := s.db.SavedSearches().Count(ctx, s.listArgs)
	return int32(count), err
}

func (s *savedSearchesConnectionStore) ComputeNodes(ctx context.Context, pgArgs *database.PaginationArgs) ([]graphqlbackend.SavedSearchResolver, error) {
	dbResults, err := s.db.SavedSearches().List(ctx, s.listArgs, pgArgs)
	if err != nil {
		return nil, err
	}

	var results []graphqlbackend.SavedSearchResolver
	for _, savedSearch := range dbResults {
		results = append(results, &savedSearchResolver{db: s.db, s: *savedSearch})
	}

	return results, nil
}
