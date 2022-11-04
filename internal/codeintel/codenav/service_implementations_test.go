package codenav

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	codeintelgitserver "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	uploadsShared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestImplementations(t *testing.T) {
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

	uploads := []shared.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)
	mockCursor := shared.ImplementationsCursor{Phase: "local"}
	mockRequest := shared.RequestArgs{
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

	uploads := []shared.Dump{
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
	mockCursor := shared.ImplementationsCursor{Phase: "local"}
	mockRequest := shared.RequestArgs{
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
	mockRequestState.SetLocalGitTreeTranslator(mockGitServer, &types.Repo{ID: 42}, mockCommit, mockPath, 50)
	uploads := []shared.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

	remoteUploads := []uploadsShared.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockUploadSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(remoteUploads, nil)

	referenceUploads := []uploadsShared.Dump{
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
	mockGitserverClient.CommitsExistFunc.SetDefaultHook(func(ctx context.Context, rcs []codeintelgitserver.RepositoryCommit) (exists []bool, _ error) {
		for _, rc := range rcs {
			exists = append(exists, rc.Commit != "deadbeef1")
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

	mockCursor := shared.ImplementationsCursor{Phase: "local"}
	mockRequest := shared.RequestArgs{
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
		{Dump: uploads[3], Path: "sub4/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
		{Dump: uploads[3], Path: "sub4/b.go", TargetCommit: "deadbeef", TargetRange: testRange4},
		{Dump: uploads[3], Path: "sub4/c.go", TargetCommit: "deadbeef", TargetRange: testRange5},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}

	if history := mockUploadSvc.GetDumpsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
		t.Fatalf("unexpected call count for dbstore.DefinitionDump. want=%d have=%d", 1, len(history))
	} else {
		expectedMonikers := []precise.QualifiedMonikerData{
			{MonikerData: monikers[0], PackageInformationData: packageInformation1},
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
			t.Errorf("unexpected monikers (-want +got):\n%s", diff)
		}
	}

	if history := mockLsifStore.GetBulkMonikerLocationsFunc.History(); len(history) != 3 {
		t.Fatalf("unexpected call count for lsifstore.BulkMonikerResults. want=%d have=%d", 3, len(history))
	} else {
		if diff := cmp.Diff([]int{151, 152, 153}, history[0].Arg2); diff != "" {
			t.Errorf("unexpected ids (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]precise.MonikerData{monikers[0]}, history[0].Arg3); diff != "" {
			t.Errorf("unexpected monikers (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]int{251}, history[1].Arg2); diff != "" {
			t.Errorf("unexpected ids (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff([]precise.MonikerData{monikers[1]}, history[1].Arg3); diff != "" {
			t.Errorf("unexpected monikers (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]int{252, 253}, history[2].Arg2); diff != "" {
			t.Errorf("unexpected ids (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff([]precise.MonikerData{monikers[1]}, history[2].Arg3); diff != "" {
			t.Errorf("unexpected monikers (-want +got):\n%s", diff)
		}
	}
}

func TestImplementationsRemoteWithSubRepoPermissions(t *testing.T) {
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

	definitionUploads := []uploadsShared.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockUploadSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(definitionUploads, nil)

	referenceUploads := []uploadsShared.Dump{
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
	mockGitserverClient.CommitsExistFunc.SetDefaultHook(func(ctx context.Context, rcs []codeintelgitserver.RepositoryCommit) (exists []bool, _ error) {
		for _, rc := range rcs {
			exists = append(exists, rc.Commit != "deadbeef1")
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
	mockCursor := shared.ImplementationsCursor{Phase: "local"}
	mockRequest := shared.RequestArgs{
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
		{Dump: uploads[3], Path: "sub4/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}

	if history := mockUploadSvc.GetDumpsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
		t.Fatalf("unexpected call count for dbstore.DefinitionDump. want=%d have=%d", 1, len(history))
	} else {
		expectedMonikers := []precise.QualifiedMonikerData{
			{MonikerData: monikers[0], PackageInformationData: packageInformation1},
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
			t.Errorf("unexpected monikers (-want +got):\n%s", diff)
		}
	}

	if history := mockLsifStore.GetBulkMonikerLocationsFunc.History(); len(history) != 3 {
		t.Fatalf("unexpected call count for lsifstore.BulkMonikerResults. want=%d have=%d", 3, len(history))
	} else {
		if diff := cmp.Diff([]int{151, 152, 153}, history[0].Arg2); diff != "" {
			t.Errorf("unexpected ids (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]precise.MonikerData{monikers[0]}, history[0].Arg3); diff != "" {
			t.Errorf("unexpected monikers (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]int{251}, history[1].Arg2); diff != "" {
			t.Errorf("unexpected ids (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff([]precise.MonikerData{monikers[1]}, history[1].Arg3); diff != "" {
			t.Errorf("unexpected monikers (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff([]int{252, 253}, history[2].Arg2); diff != "" {
			t.Errorf("unexpected ids (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff([]precise.MonikerData{monikers[1]}, history[2].Arg3); diff != "" {
			t.Errorf("unexpected monikers (-want +got):\n%s", diff)
		}
	}
}
