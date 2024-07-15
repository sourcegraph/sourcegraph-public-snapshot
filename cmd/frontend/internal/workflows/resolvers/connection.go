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

const cursorKind = "WorkflowCursor"

// marshalCursor marshals a workflows pagination cursor.
func marshalCursor(cursor types.MultiCursor) string {
	return string(relay.MarshalID(cursorKind, cursor))
}

// unmarshalCursor unmarshals a workflows pagination cursor.
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

type workflowsConnectionStore struct {
	db       database.DB
	listArgs database.WorkflowListArgs
}

func (s *workflowsConnectionStore) MarshalCursor(node graphqlbackend.WorkflowResolver, orderBy database.OrderBy) (*string, error) {
	cursor, err := s.db.Workflows().MarshalToCursor(&node.(*workflowResolver).s, orderBy)
	if err != nil {
		return nil, err
	}
	cursorStr := marshalCursor(cursor)
	return &cursorStr, nil
}

func (s *workflowsConnectionStore) UnmarshalCursor(cursorStr string, orderBy database.OrderBy) ([]any, error) {
	cursor, err := unmarshalCursor(cursorStr)
	if err != nil {
		return nil, err
	}
	return s.db.Workflows().UnmarshalValuesFromCursor(cursor)
}

func (s *workflowsConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	count, err := s.db.Workflows().Count(ctx, s.listArgs)
	return int32(count), err
}

func (s *workflowsConnectionStore) ComputeNodes(ctx context.Context, pgArgs *database.PaginationArgs) ([]graphqlbackend.WorkflowResolver, error) {
	dbResults, err := s.db.Workflows().List(ctx, s.listArgs, pgArgs)
	if err != nil {
		return nil, err
	}

	var results []graphqlbackend.WorkflowResolver
	for _, workflow := range dbResults {
		results = append(results, &workflowResolver{db: s.db, s: *workflow})
	}

	return results, nil
}
