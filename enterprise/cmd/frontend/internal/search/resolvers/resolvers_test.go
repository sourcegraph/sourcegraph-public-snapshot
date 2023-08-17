package resolvers_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/graph-gophers/graphql-go"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/require"
)

func exec(
	ctx context.Context,
	t *testing.T,
	s *graphql.Schema,
	query string,
	variables map[string]any,
	out any,
) []*gqlerrors.QueryError {
	t.Helper()

	query = strings.ReplaceAll(query, "\t", "  ")

	res := s.Exec(ctx, query, "", variables)
	if len(res.Errors) != 0 {
		return res.Errors
	}

	err := json.Unmarshal(res.Data, out)
	require.NoError(t, err)
	return nil
}
