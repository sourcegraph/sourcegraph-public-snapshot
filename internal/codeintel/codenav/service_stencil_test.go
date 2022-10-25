package codenav

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	codeintelgitserver "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
)

func TestStencil(t *testing.T) {
	// Set up mocks
	mockStore := NewMockStore()
	mockLsifStore := NewMockLsifStore()
	mockUploadSvc := NewMockUploadService()
	mockGitserverClient := NewMockGitserverClient()
	mockGitServer := codeintelgitserver.New(database.NewMockDB(), &observation.TestContext)

	// Init service
	svc := newService(mockStore, mockLsifStore, mockUploadSvc, mockGitserverClient, &observation.TestContext)

	// Set up request state
	mockRequestState := RequestState{}
	mockRequestState.SetLocalCommitCache(mockGitserverClient)
	mockRequestState.SetLocalGitTreeTranslator(mockGitServer, &sgtypes.Repo{}, mockCommit, mockPath, 50)
	uploads := []types.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

	expectedRanges := []types.Range{
		{Start: types.Position{Line: 10, Character: 20}, End: types.Position{Line: 10, Character: 30}},
		{Start: types.Position{Line: 11, Character: 20}, End: types.Position{Line: 11, Character: 30}},
		{Start: types.Position{Line: 12, Character: 20}, End: types.Position{Line: 12, Character: 30}},
		{Start: types.Position{Line: 13, Character: 20}, End: types.Position{Line: 13, Character: 30}},
		{Start: types.Position{Line: 14, Character: 20}, End: types.Position{Line: 14, Character: 30}},
		{Start: types.Position{Line: 15, Character: 20}, End: types.Position{Line: 15, Character: 30}},
		{Start: types.Position{Line: 16, Character: 20}, End: types.Position{Line: 16, Character: 30}},
		{Start: types.Position{Line: 17, Character: 20}, End: types.Position{Line: 17, Character: 30}},
		{Start: types.Position{Line: 18, Character: 20}, End: types.Position{Line: 18, Character: 30}},
		{Start: types.Position{Line: 19, Character: 20}, End: types.Position{Line: 19, Character: 30}},
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

func TestStencilWithDuplicateRanges(t *testing.T) {
	// Set up mocks
	mockStore := NewMockStore()
	mockLsifStore := NewMockLsifStore()
	mockUploadSvc := NewMockUploadService()
	mockGitserverClient := NewMockGitserverClient()
	mockGitServer := codeintelgitserver.New(database.NewMockDB(), &observation.TestContext)

	// Init service
	svc := newService(mockStore, mockLsifStore, mockUploadSvc, mockGitserverClient, &observation.TestContext)

	// Set up request state
	mockRequestState := RequestState{}
	mockRequestState.SetLocalCommitCache(mockGitserverClient)
	mockRequestState.SetLocalGitTreeTranslator(mockGitServer, &sgtypes.Repo{}, mockCommit, mockPath, 50)
	uploads := []types.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

	expectedRanges := []types.Range{
		{Start: types.Position{Line: 10, Character: 20}, End: types.Position{Line: 10, Character: 30}},
		{Start: types.Position{Line: 11, Character: 20}, End: types.Position{Line: 11, Character: 30}},
		{Start: types.Position{Line: 12, Character: 20}, End: types.Position{Line: 12, Character: 30}},
		{Start: types.Position{Line: 13, Character: 20}, End: types.Position{Line: 13, Character: 30}},
		{Start: types.Position{Line: 14, Character: 20}, End: types.Position{Line: 14, Character: 30}},
		{Start: types.Position{Line: 15, Character: 20}, End: types.Position{Line: 15, Character: 30}},
		{Start: types.Position{Line: 16, Character: 20}, End: types.Position{Line: 16, Character: 30}},
		{Start: types.Position{Line: 17, Character: 20}, End: types.Position{Line: 17, Character: 30}},
		{Start: types.Position{Line: 18, Character: 20}, End: types.Position{Line: 18, Character: 30}},
		{Start: types.Position{Line: 19, Character: 20}, End: types.Position{Line: 19, Character: 30}},
	}
	mockLsifStore.GetStencilFunc.PushReturn(nil, nil)

	// Duplicate the ranges to test that we dedupe them
	mockLsifStore.GetStencilFunc.PushReturn(append(expectedRanges, expectedRanges...), nil)

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
