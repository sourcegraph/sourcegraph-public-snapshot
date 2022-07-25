package graphql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestStencil(t *testing.T) {
	// Set up mocks
	mockLogger := logtest.Scoped(t)
	mockDB := database.NewDB(mockLogger, dbtest.NewDB(mockLogger, t))
	mockGitServer := gitserver.NewClient(mockDB)
	mockGitserverClient := NewMockGitserverClient()
	mockSvc := NewMockService()

	// Init resolver and set local request context
	resolver := New(mockSvc, 50, &observation.TestContext)
	resolver.SetLocalCommitCache(mockGitserverClient)
	resolver.SetLocalGitTreeTranslator(mockGitServer, &types.Repo{}, mockCommit, mockPath)

	expectedRanges := []shared.Range{
		{Start: shared.Position{Line: 10, Character: 20}, End: shared.Position{Line: 10, Character: 30}},
		{Start: shared.Position{Line: 11, Character: 20}, End: shared.Position{Line: 11, Character: 30}},
		{Start: shared.Position{Line: 12, Character: 20}, End: shared.Position{Line: 12, Character: 30}},
		{Start: shared.Position{Line: 13, Character: 20}, End: shared.Position{Line: 13, Character: 30}},
		{Start: shared.Position{Line: 14, Character: 20}, End: shared.Position{Line: 14, Character: 30}},
		{Start: shared.Position{Line: 15, Character: 20}, End: shared.Position{Line: 15, Character: 30}},
		{Start: shared.Position{Line: 16, Character: 20}, End: shared.Position{Line: 16, Character: 30}},
		{Start: shared.Position{Line: 17, Character: 20}, End: shared.Position{Line: 17, Character: 30}},
		{Start: shared.Position{Line: 18, Character: 20}, End: shared.Position{Line: 18, Character: 30}},
		{Start: shared.Position{Line: 19, Character: 20}, End: shared.Position{Line: 19, Character: 30}},
	}
	mockSvc.GetStencilFunc.PushReturn(nil, nil)
	mockSvc.GetStencilFunc.PushReturn(expectedRanges, nil)

	uploads := []dbstore.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	resolver.SetUploadsDataLoader(uploads)
	// resolver := newQueryResolver(
	// 	database.NewMockDB(),
	// 	mockDBStore,
	// 	mockLSIFStore,
	// 	newCachedCommitChecker(mockGitserverClient),
	// 	mockPositionAdjuster,
	// 	42,
	// 	"deadbeef",
	// 	"s1/main.go",
	// 	uploads,
	// 	newOperations(&observation.TestContext),
	// 	authz.NewMockSubRepoPermissionChecker(),
	// 	50,
	// 	mockSymbolsResolver,
	// )
	mockRequest := shared.RequestArgs{
		RepositoryID: 42,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
		Limit:        50,
	}
	ranges, err := resolver.Stencil(context.Background(), mockRequest)
	if err != nil {
		t.Fatalf("unexpected error querying hover: %s", err)
	}

	if diff := cmp.Diff(expectedRanges, ranges); diff != "" {
		t.Errorf("unexpected range (-want +got):\n%s", diff)
	}
}
