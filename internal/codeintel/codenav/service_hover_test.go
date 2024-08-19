package codenav

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	lsifstoremocks "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestHover(t *testing.T) {
	// Set up mocks
	fakeRepoStore := AllPresentFakeRepoStore{}
	mockLsifStore := lsifstoremocks.NewMockLsifStore()
	mockUploadSvc := NewMockUploadService()
	mockGitserverClient := gitserver.NewMockClient()
	mockSearchClient := client.NewMockSearchClient()

	// Init service
	svc := newService(observation.TestContextTB(t), fakeRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient, mockSearchClient, log.NoOp())

	// Set up request state
	mockRequestState := RequestState{}
	mockRequestState.SetLocalCommitCache(fakeRepoStore, mockGitserverClient)
	mockRequestState.GitTreeTranslator = noopTranslator()
	uploads := []uploadsshared.CompletedUpload{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

	expectedRange := shared.Range{
		Start: shared.Position{Line: 10, Character: 10},
		End:   shared.Position{Line: 15, Character: 25},
	}
	mockLsifStore.GetHoverFunc.PushReturn("", shared.Range{}, false, nil)
	mockLsifStore.GetHoverFunc.PushReturn("doctext", expectedRange, true, nil)

	mockRequest := PositionalRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
			Limit:        50,
		},
		Path:      mockPath,
		Line:      10,
		Character: 20,
	}
	text, rn, exists, err := svc.GetHover(context.Background(), mockRequest, mockRequestState)
	if err != nil {
		t.Fatalf("unexpected error querying hover: %s", err)
	}
	if !exists {
		t.Fatalf("expected hover to exist")
	}

	if text != "doctext" {
		t.Errorf("unexpected text. want=%q have=%q", "doctext", text)
	}
	if diff := cmp.Diff(expectedRange, rn); diff != "" {
		t.Errorf("unexpected range (-want +got):\n%s", diff)
	}
}

func TestHoverRemote(t *testing.T) {
	// Set up mocks
	fakeRepoStore := AllPresentFakeRepoStore{}
	mockLsifStore := lsifstoremocks.NewMockLsifStore()
	mockUploadSvc := NewMockUploadService()
	mockGitserverClient := gitserver.NewMockClient()
	mockSearchClient := client.NewMockSearchClient()

	// Init service
	svc := newService(observation.TestContextTB(t), fakeRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient, mockSearchClient, log.NoOp())

	// Set up request state
	mockRequestState := RequestState{}
	mockRequestState.SetLocalCommitCache(fakeRepoStore, mockGitserverClient)
	mockRequestState.GitTreeTranslator = noopTranslator()
	uploads := []uploadsshared.CompletedUpload{
		{ID: 50, Commit: "deadbeef"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

	expectedRange := shared.Range{
		Start: shared.Position{Line: 10, Character: 10},
		End:   shared.Position{Line: 15, Character: 25},
	}
	mockLsifStore.GetHoverFunc.PushReturn("", expectedRange, true, nil)

	remoteRange := shared.Range{
		Start: shared.Position{Line: 30, Character: 30},
		End:   shared.Position{Line: 35, Character: 45},
	}
	mockLsifStore.GetHoverFunc.PushReturn("doctext", remoteRange, true, nil)

	uploadsWithDefinitions := []uploadsshared.CompletedUpload{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockUploadSvc.GetCompletedUploadsWithDefinitionsForMonikersFunc.PushReturn(uploadsWithDefinitions, nil)

	monikers := []precise.MonikerData{
		{Kind: "import", Scheme: "tsc", Identifier: "padLeft", PackageInformationID: "51"},
		{Kind: "export", Scheme: "tsc", Identifier: "pad_left", PackageInformationID: "52"},
		{Kind: "import", Scheme: "tsc", Identifier: "pad-left", PackageInformationID: "53"},
		{Kind: "import", Scheme: "tsc", Identifier: "left_pad"},
	}
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[0]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[1]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[2]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[3]}}, nil)

	packageInformation1 := precise.PackageInformationData{Name: "leftpad", Version: "0.1.0"}
	packageInformation2 := precise.PackageInformationData{Name: "leftpad", Version: "0.2.0"}
	mockLsifStore.GetPackageInformationFunc.PushReturn(packageInformation1, true, nil)
	mockLsifStore.GetPackageInformationFunc.PushReturn(packageInformation2, true, nil)

	uploadRelPath := core.NewUploadRelPathUnchecked
	usages := []shared.Usage{
		{UploadID: 151, Path: uploadRelPath("a.go"), Range: testRange1},
		{UploadID: 151, Path: uploadRelPath("b.go"), Range: testRange2},
		{UploadID: 151, Path: uploadRelPath("a.go"), Range: testRange3},
		{UploadID: 151, Path: uploadRelPath("b.go"), Range: testRange4},
		{UploadID: 151, Path: uploadRelPath("c.go"), Range: testRange5},
	}
	mockLsifStore.GetSymbolUsagesFunc.PushReturn(usages, 0, nil)
	mockLsifStore.GetSymbolUsagesFunc.PushReturn(usages, len(usages), nil)

	mockGitserverClient.GetCommitFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID) (*gitdomain.Commit, error) {
		return &gitdomain.Commit{ID: "sha"}, nil
	})

	mockRequest := PositionalRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
			Limit:        50,
		},
		Path:      mockPath,
		Line:      10,
		Character: 20,
	}
	text, rn, exists, err := svc.GetHover(context.Background(), mockRequest, mockRequestState)
	if err != nil {
		t.Fatalf("unexpected error querying hover: %s", err)
	}
	if !exists {
		t.Fatalf("expected hover to exist")
	}

	if text != "doctext" {
		t.Errorf("unexpected text. want=%q have=%q", "doctext", text)
	}
	if diff := cmp.Diff(expectedRange, rn); diff != "" {
		t.Errorf("unexpected range (-want +got):\n%s", diff)
	}
}
