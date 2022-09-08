package codenav

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	codeintelgitserver "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestStencil(t *testing.T) {
	// Set up mocks
	mockStore := NewMockStore()
	mockLsifStore := NewMockLsifStore()
	mockUploadSvc := NewMockUploadService()
	mockDBStore := NewMockDBStore()
	mockGitserverClient := NewMockGitserverClient()
	mockGitServer := codeintelgitserver.New(database.NewMockDB(), mockDBStore, &observation.TestContext)

	// Init service
	svc := newService(mockStore, mockLsifStore, mockUploadSvc, mockGitserverClient, nil, &observation.TestContext)

	// Set up request state
	mockRequestState := RequestState{}
	mockRequestState.SetLocalCommitCache(mockGitserverClient)
	mockRequestState.SetLocalGitTreeTranslator(mockGitServer, &types.Repo{}, mockCommit, mockPath, 50)
	uploads := []shared.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

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
	mockLsifStore.GetStencilFunc.PushReturn(nil, nil)
	mockLsifStore.GetStencilFunc.PushReturn(expectedRanges, nil)

	mockRequest := shared.RequestArgs{
		RepositoryID: 42,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
		Limit:        50,
	}
	ranges, err := svc.GetStencil(context.Background(), mockRequest, mockRequestState)
	if err != nil {
		t.Fatalf("unexpected error querying hover: %s", err)
	}

	if diff := cmp.Diff(expectedRanges, ranges); diff != "" {
		t.Errorf("unexpected range (-want +got):\n%s", diff)
	}
}
