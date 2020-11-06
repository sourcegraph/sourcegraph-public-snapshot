package resolvers

import (
	"context"
	"testing"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	apimocks "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/api/mocks"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore/mocks"
	bundlemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore/mocks"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestQueryResolver(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleStore := bundlemocks.NewMockStore()
	mockCodeIntelAPI := apimocks.NewMockCodeIntelAPI() // returns no dumps

	resolver := NewResolver(mockStore, mockBundleStore, mockCodeIntelAPI, nil)
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
