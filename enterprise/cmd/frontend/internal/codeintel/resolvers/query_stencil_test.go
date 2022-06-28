package resolvers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestStencil(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()
	mockPositionAdjuster := noopPositionAdjuster()

	expectedRanges := []lsifstore.Range{
		{Start: lsifstore.Position{Line: 10, Character: 20}, End: lsifstore.Position{Line: 10, Character: 30}},
		{Start: lsifstore.Position{Line: 11, Character: 20}, End: lsifstore.Position{Line: 11, Character: 30}},
		{Start: lsifstore.Position{Line: 12, Character: 20}, End: lsifstore.Position{Line: 12, Character: 30}},
		{Start: lsifstore.Position{Line: 13, Character: 20}, End: lsifstore.Position{Line: 13, Character: 30}},
		{Start: lsifstore.Position{Line: 14, Character: 20}, End: lsifstore.Position{Line: 14, Character: 30}},
		{Start: lsifstore.Position{Line: 15, Character: 20}, End: lsifstore.Position{Line: 15, Character: 30}},
		{Start: lsifstore.Position{Line: 16, Character: 20}, End: lsifstore.Position{Line: 16, Character: 30}},
		{Start: lsifstore.Position{Line: 17, Character: 20}, End: lsifstore.Position{Line: 17, Character: 30}},
		{Start: lsifstore.Position{Line: 18, Character: 20}, End: lsifstore.Position{Line: 18, Character: 30}},
		{Start: lsifstore.Position{Line: 19, Character: 20}, End: lsifstore.Position{Line: 19, Character: 30}},
	}
	mockLSIFStore.StencilFunc.PushReturn(nil, nil)
	mockLSIFStore.StencilFunc.PushReturn(expectedRanges, nil)

	uploads := []dbstore.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	resolver := newQueryResolver(
		database.NewMockDB(),
		mockDBStore,
		mockLSIFStore,
		newCachedCommitChecker(mockGitserverClient),
		mockPositionAdjuster,
		42,
		"deadbeef",
		"s1/main.go",
		uploads,
		newOperations(&observation.TestContext),
		authz.NewMockSubRepoPermissionChecker(),
		50,
	)
	ranges, err := resolver.Stencil(context.Background())
	if err != nil {
		t.Fatalf("unexpected error querying hover: %s", err)
	}

	if diff := cmp.Diff(expectedRanges, ranges); diff != "" {
		t.Errorf("unexpected range (-want +got):\n%s", diff)
	}
}
