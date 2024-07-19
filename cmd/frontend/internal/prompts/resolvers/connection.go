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

const cursorKind = "PromptCursor"

// marshalCursor marshals a prompts pagination cursor.
func marshalCursor(cursor types.MultiCursor) string {
	return string(relay.MarshalID(cursorKind, cursor))
}

// unmarshalCursor unmarshals a prompts pagination cursor.
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

type promptsConnectionStore struct {
	db       database.DB
	listArgs database.PromptListArgs
}

func (s *promptsConnectionStore) MarshalCursor(node graphqlbackend.PromptResolver, orderBy database.OrderBy) (*string, error) {
	cursor, err := s.db.Prompts().MarshalToCursor(&node.(*promptResolver).s, orderBy)
	if err != nil {
		return nil, err
	}
	cursorStr := marshalCursor(cursor)
	return &cursorStr, nil
}

func (s *promptsConnectionStore) UnmarshalCursor(cursorStr string, orderBy database.OrderBy) ([]any, error) {
	cursor, err := unmarshalCursor(cursorStr)
	if err != nil {
		return nil, err
	}
	return s.db.Prompts().UnmarshalValuesFromCursor(cursor)
}

func (s *promptsConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	count, err := s.db.Prompts().Count(ctx, s.listArgs)
	return int32(count), err
}

func (s *promptsConnectionStore) ComputeNodes(ctx context.Context, pgArgs *database.PaginationArgs) ([]graphqlbackend.PromptResolver, error) {
	dbResults, err := s.db.Prompts().List(ctx, s.listArgs, pgArgs)
	if err != nil {
		return nil, err
	}

	var results []graphqlbackend.PromptResolver
	for _, prompt := range dbResults {
		results = append(results, &promptResolver{db: s.db, s: *prompt})
	}

	return results, nil
}
