package codenav

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestImplementations(t *testing.T) {
	// Set up mocks
	mockRepoStore := defaultMockRepoStore()
	mockLsifStore := NewMockLsifStore()
	mockUploadSvc := NewMockUploadService()
	mockGitserverClient := gitserver.NewMockClient()
	hunkCache, _ := NewHunkCache(50)

	// Init service
	svc := newService(&observation.TestContext, mockRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient)

	// Set up request state
	mockRequestState := RequestState{}
	mockRequestState.SetLocalCommitCache(mockRepoStore, mockGitserverClient)
	mockRequestState.SetLocalGitTreeTranslator(mockGitserverClient, &sgtypes.Repo{}, mockCommit, mockPath, hunkCache)

	// Empty result set (prevents nil pointer as scanner is always non-nil)
	mockUploadSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{}, 0, 0, nil)

	locations := []shared.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockLsifStore.GetImplementationLocationsFunc.PushReturn(locations[:1], 1, nil)
	mockLsifStore.GetImplementationLocationsFunc.PushReturn(locations[1:4], 3, nil)
	mockLsifStore.GetImplementationLocationsFunc.PushReturn(locations[4:], 1, nil)

	uploads := []uploadsshared.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)
	mockCursor := ImplementationsCursor{Phase: "local"}
	mockRequest := RequestArgs{
		RepositoryID: 51,
		Commit:       "deadbeef",
		Path:         "s1/main.go",
		Line:         10,
		Character:    20,
		Limit:        50,
	}
	adjustedLocations, _, err := svc.GetImplementations(context.Background(), mockRequest, mockRequestState, mockCursor)
	if err != nil {
		t.Fatalf("unexpected error querying implementations: %s", err)
	}

	expectedLocations := []shared.UploadLocation{
		{Dump: uploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: uploads[1], Path: "sub2/b.go", TargetCommit: "deadbeef", TargetRange: testRange2},
		{Dump: uploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
		{Dump: uploads[1], Path: "sub2/b.go", TargetCommit: "deadbeef", TargetRange: testRange4},
		{Dump: uploads[1], Path: "sub2/c.go", TargetCommit: "deadbeef", TargetRange: testRange5},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}

func TestImplementationsWithSubRepoPermissions(t *testing.T) {
	// Set up mocks
	mockRepoStore := defaultMockRepoStore()
	mockLsifStore := NewMockLsifStore()
	mockUploadSvc := NewMockUploadService()
	mockGitserverClient := gitserver.NewMockClient()
	hunkCache, _ := NewHunkCache(50)

	// Init service
	svc := newService(&observation.TestContext, mockRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient)

	// Set up request state
	mockRequestState := RequestState{}
	mockRequestState.SetLocalCommitCache(mockRepoStore, mockGitserverClient)
	mockRequestState.SetLocalGitTreeTranslator(mockGitserverClient, &sgtypes.Repo{}, mockCommit, mockPath, hunkCache)

	// Empty result set (prevents nil pointer as scanner is always non-nil)
	mockUploadSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{}, 0, 0, nil)

	locations := []shared.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockLsifStore.GetImplementationLocationsFunc.PushReturn(locations[:1], 1, nil)
	mockLsifStore.GetImplementationLocationsFunc.PushReturn(locations[1:4], 3, nil)
	mockLsifStore.GetImplementationLocationsFunc.PushReturn(locations[4:], 1, nil)

	uploads := []uploadsshared.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

	// Applying sub-repo permissions
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		if content.Path == "sub2/a.go" {
			return authz.Read, nil
		}
		return authz.None, nil
	})
	mockRequestState.SetAuthChecker(checker)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	mockCursor := ImplementationsCursor{Phase: "local"}
	mockRequest := RequestArgs{
		RepositoryID: 42,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
		Limit:        50,
	}
	adjustedLocations, _, err := svc.GetImplementations(ctx, mockRequest, mockRequestState, mockCursor)
	if err != nil {
		t.Fatalf("unexpected error querying implementations: %s", err)
	}

	expectedLocations := []shared.UploadLocation{
		{Dump: uploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: uploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}

func TestImplementationsRemote(t *testing.T) {
	// Set up mocks
	mockRepoStore := defaultMockRepoStore()
	mockLsifStore := NewMockLsifStore()
	mockUploadSvc := NewMockUploadService()
	mockGitserverClient := gitserver.NewMockClient()
	hunkCache, _ := NewHunkCache(50)

	// Init service
	svc := newService(&observation.TestContext, mockRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient)

	// Set up request state
	mockRequestState := RequestState{}
	mockRequestState.SetLocalCommitCache(mockRepoStore, mockGitserverClient)
	mockRequestState.SetLocalGitTreeTranslator(mockGitserverClient, &sgtypes.Repo{ID: 42}, mockCommit, mockPath, hunkCache)
	uploads := []uploadsshared.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

	remoteUploads := []uploadsshared.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockUploadSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(remoteUploads, nil)

	referenceUploads := []uploadsshared.Dump{
		{ID: 250, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 251, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 252, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 253, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockUploadSvc.GetDumpsByIDsFunc.PushReturn(referenceUploads[:2], nil)
	mockUploadSvc.GetDumpsByIDsFunc.PushReturn(referenceUploads[2:], nil)

	mockUploadSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{250, 251}, 0, 4, nil)
	mockUploadSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{252, 253}, 0, 2, nil)

	// upload #150/#250's commits no longer exists; all others do
	mockGitserverClient.CommitsExistFunc.SetDefaultHook(func(ctx context.Context, _ authz.SubRepoPermissionChecker, rcs []api.RepoCommit) (exists []bool, _ error) {
		for _, rc := range rcs {
			exists = append(exists, rc.CommitID != "deadbeef1")
		}
		return
	})

	monikers := []precise.MonikerData{
		{Kind: "implementation", Scheme: "tsc", Identifier: "padLeft", PackageInformationID: "51"},
		{Kind: "export", Scheme: "tsc", Identifier: "pad_left", PackageInformationID: "52"},
	}
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[0]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[1]}}, nil)

	packageInformation1 := precise.PackageInformationData{Name: "leftpad", Version: "0.1.0"}
	packageInformation2 := precise.PackageInformationData{Name: "leftpad", Version: "0.2.0"}
	mockLsifStore.GetPackageInformationFunc.PushReturn(packageInformation1, true, nil)
	mockLsifStore.GetPackageInformationFunc.PushReturn(packageInformation2, true, nil)

	locations := []shared.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockLsifStore.GetImplementationLocationsFunc.PushReturn(locations[:1], 1, nil)
	mockLsifStore.GetImplementationLocationsFunc.PushReturn(locations[1:4], 3, nil)
	mockLsifStore.GetImplementationLocationsFunc.PushReturn(locations[4:5], 1, nil)

	monikerLocations := []shared.Location{
		{DumpID: 53, Path: "a.go", Range: testRange1},
		{DumpID: 53, Path: "b.go", Range: testRange2},
		{DumpID: 53, Path: "a.go", Range: testRange3},
		{DumpID: 53, Path: "b.go", Range: testRange4},
		{DumpID: 53, Path: "c.go", Range: testRange5},
	}
	mockLsifStore.GetBulkMonikerLocationsFunc.PushReturn(monikerLocations[0:1], 1, nil) // defs
	mockLsifStore.GetBulkMonikerLocationsFunc.PushReturn(monikerLocations[1:2], 1, nil) // impls batch 1
	mockLsifStore.GetBulkMonikerLocationsFunc.PushReturn(monikerLocations[2:], 3, nil)  // impls batch 2

	mockCursor := ImplementationsCursor{Phase: "local"}
	mockRequest := RequestArgs{
		RepositoryID: 42,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
		Limit:        50,
	}
	adjustedLocations, _, err := svc.GetImplementations(context.Background(), mockRequest, mockRequestState, mockCursor)
	if err != nil {
		t.Fatalf("unexpected error querying references: %s", err)
	}

	expectedLocations := []shared.UploadLocation{
		{Dump: uploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: uploads[1], Path: "sub2/b.go", TargetCommit: "deadbeef", TargetRange: testRange2},
		{Dump: uploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
		{Dump: uploads[1], Path: "sub2/b.go", TargetCommit: "deadbeef", TargetRange: testRange4},
		{Dump: uploads[1], Path: "sub2/c.go", TargetCommit: "deadbeef", TargetRange: testRange5},
		{Dump: uploads[3], Path: "sub4/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: uploads[3], Path: "sub4/b.go", TargetCommit: "deadbeef", TargetRange: testRange2},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}

func TestImplementationsRemoteWithSubRepoPermissions(t *testing.T) {
	// Set up mocks
	mockRepoStore := defaultMockRepoStore()
	mockLsifStore := NewMockLsifStore()
	mockUploadSvc := NewMockUploadService()
	mockGitserverClient := gitserver.NewMockClient()
	hunkCache, _ := NewHunkCache(50)

	// Init service
	svc := newService(&observation.TestContext, mockRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient)

	// Set up request state
	mockRequestState := RequestState{}
	mockRequestState.SetLocalCommitCache(mockRepoStore, mockGitserverClient)
	mockRequestState.SetLocalGitTreeTranslator(mockGitserverClient, &sgtypes.Repo{}, mockCommit, mockPath, hunkCache)
	uploads := []uploadsshared.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

	// Applying sub-repo permissions
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		if content.Path == "sub2/a.go" || content.Path == "sub4/a.go" {
			return authz.Read, nil
		}
		return authz.None, nil
	})
	mockRequestState.SetAuthChecker(checker)

	definitionUploads := []uploadsshared.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockUploadSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(definitionUploads, nil)

	referenceUploads := []uploadsshared.Dump{
		{ID: 250, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 251, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 252, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 253, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockUploadSvc.GetDumpsByIDsFunc.PushReturn(referenceUploads[:2], nil)
	mockUploadSvc.GetDumpsByIDsFunc.PushReturn(referenceUploads[2:], nil)

	mockUploadSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{250, 251}, 0, 4, nil)
	mockUploadSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{252, 253}, 0, 2, nil)

	// upload #150/#250's commits no longer exists; all others do
	mockGitserverClient.CommitsExistFunc.SetDefaultHook(func(ctx context.Context, _ authz.SubRepoPermissionChecker, rcs []api.RepoCommit) (exists []bool, _ error) {
		for _, rc := range rcs {
			exists = append(exists, rc.CommitID != "deadbeef1")
		}
		return
	})

	monikers := []precise.MonikerData{
		{Kind: "implementation", Scheme: "tsc", Identifier: "padLeft", PackageInformationID: "51"},
		{Kind: "export", Scheme: "tsc", Identifier: "pad_left", PackageInformationID: "52"},
	}
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[0]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[1]}}, nil)

	packageInformation1 := precise.PackageInformationData{Name: "leftpad", Version: "0.1.0"}
	packageInformation2 := precise.PackageInformationData{Name: "leftpad", Version: "0.2.0"}
	mockLsifStore.GetPackageInformationFunc.PushReturn(packageInformation1, true, nil)
	mockLsifStore.GetPackageInformationFunc.PushReturn(packageInformation2, true, nil)

	locations := []shared.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockLsifStore.GetImplementationLocationsFunc.PushReturn(locations[:1], 1, nil)
	mockLsifStore.GetImplementationLocationsFunc.PushReturn(locations[1:4], 3, nil)
	mockLsifStore.GetImplementationLocationsFunc.PushReturn(locations[4:5], 1, nil)

	monikerLocations := []shared.Location{
		{DumpID: 53, Path: "a.go", Range: testRange1},
		{DumpID: 53, Path: "b.go", Range: testRange2},
		{DumpID: 53, Path: "a.go", Range: testRange3},
		{DumpID: 53, Path: "b.go", Range: testRange4},
		{DumpID: 53, Path: "c.go", Range: testRange5},
	}
	mockLsifStore.GetBulkMonikerLocationsFunc.PushReturn(monikerLocations[0:1], 1, nil) // defs
	mockLsifStore.GetBulkMonikerLocationsFunc.PushReturn(monikerLocations[1:2], 1, nil) // impls batch 1
	mockLsifStore.GetBulkMonikerLocationsFunc.PushReturn(monikerLocations[2:], 3, nil)  // impls batch 2

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	mockCursor := ImplementationsCursor{Phase: "local"}
	mockRequest := RequestArgs{
		RepositoryID: 42,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
		Limit:        50,
	}
	adjustedLocations, _, err := svc.GetImplementations(ctx, mockRequest, mockRequestState, mockCursor)
	if err != nil {
		t.Fatalf("unexpected error querying references: %s", err)
	}

	expectedLocations := []shared.UploadLocation{
		{Dump: uploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: uploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
		{Dump: uploads[3], Path: "sub4/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}
