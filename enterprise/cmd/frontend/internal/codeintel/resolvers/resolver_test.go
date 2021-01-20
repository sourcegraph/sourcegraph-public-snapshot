package resolvers

import (
	"context"
	"testing"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestQueryResolver(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockCodeIntelAPI := NewMockCodeIntelAPI() // returns no dumps

	resolver := NewResolver(mockDBStore, mockLSIFStore, mockCodeIntelAPI, nil, &observation.TestContext)
	queryResolver, err := resolver.QueryResolver(context.Background(), &gql.GitBlobLSIFDataArgs{
		Repo:      &types.Repo{ID: 50},
		Commit:    api.CommitID("deadbeef"),
		Path:      "/foo/bar.go",
		ExactPath: true,
		ToolName:  "lsif-go",
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if queryResolver != nil {
		t.Errorf("expected nil-valued resolver")
	}
}
