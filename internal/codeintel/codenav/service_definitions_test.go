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

func TestDefinitions(t *testing.T) {
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
		{ID: 50, Commit: mockCommit, Root: "sub1/"},
		{ID: 51, Commit: mockCommit, Root: "sub2/"},
		{ID: 52, Commit: mockCommit, Root: "sub3/"},
		{ID: 53, Commit: mockCommit, Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

	locations := []shared.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockLsifStore.GetDefinitionLocationsFunc.PushReturn(locations, len(locations), nil)

	mockRequest := RequestArgs{
		RepositoryID: 51,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
	}
	adjustedLocations, err := svc.GetDefinitions(context.Background(), mockRequest, mockRequestState)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	}
	expectedLocations := []shared.UploadLocation{
		{Dump: uploads[1], Path: "sub2/a.go", TargetCommit: mockCommit, TargetRange: testRange1},
		{Dump: uploads[1], Path: "sub2/b.go", TargetCommit: mockCommit, TargetRange: testRange2},
		{Dump: uploads[1], Path: "sub2/a.go", TargetCommit: mockCommit, TargetRange: testRange3},
		{Dump: uploads[1], Path: "sub2/b.go", TargetCommit: mockCommit, TargetRange: testRange4},
		{Dump: uploads[1], Path: "sub2/c.go", TargetCommit: mockCommit, TargetRange: testRange5},
	}

	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}

func TestDefinitionsWithSubRepoPermissions(t *testing.T) {
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
		{ID: 50, Commit: mockCommit, Root: "sub1/"},
		{ID: 51, Commit: mockCommit, Root: "sub2/"},
		{ID: 52, Commit: mockCommit, Root: "sub3/"},
		{ID: 53, Commit: mockCommit, Root: "sub4/"},
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

	locations := []shared.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockLsifStore.GetDefinitionLocationsFunc.PushReturn(locations, len(locations), nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	mockRequest := RequestArgs{
		RepositoryID: 51,
		Commit:       "deadbeef",
		Path:         "s1/main.go",
		Line:         10,
		Character:    20,
	}
	adjustedLocations, err := svc.GetDefinitions(ctx, mockRequest, mockRequestState)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	}

	expectedLocations := []shared.UploadLocation{
		{Dump: uploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: uploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}

func TestDefinitionsRemote(t *testing.T) {
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
	err := mockRequestState.SetLocalGitTreeTranslator(mockGitserverClient, &sgtypes.Repo{ID: 42}, mockCommit, mockPath, hunkCache)
	if err != nil {
		t.Fatalf("unexpected error setting local git tree translator: %s", err)
	}
	mockRequestState.GitTreeTranslator = mockedGitTreeTranslator()
	uploads := []uploadsshared.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

	dumps := []uploadsshared.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockUploadSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(dumps, nil)

	// upload #150's commit no longer exists; all others do
	mockGitserverClient.CommitsExistFunc.SetDefaultHook(func(ctx context.Context, _ authz.SubRepoPermissionChecker, rcs []api.RepoCommit) (exists []bool, _ error) {
		for _, rc := range rcs {
			exists = append(exists, rc.CommitID != "deadbeef1")
		}
		return
	})

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

	locations := []shared.Location{
		{DumpID: 151, Path: "a.go", Range: testRange1},
		{DumpID: 151, Path: "b.go", Range: testRange2},
		{DumpID: 151, Path: "a.go", Range: testRange3},
		{DumpID: 151, Path: "b.go", Range: testRange4},
		{DumpID: 151, Path: "c.go", Range: testRange5},
	}
	mockLsifStore.GetBulkMonikerLocationsFunc.PushReturn(locations, len(locations), nil)

	mockRequest := RequestArgs{
		RepositoryID: 42,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
	}
	remoteUploads := dumps
	adjustedLocations, err := svc.GetDefinitions(context.Background(), mockRequest, mockRequestState)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	}

	xLocations := []shared.UploadLocation{
		{Dump: remoteUploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef2", TargetRange: testRange1},
		{Dump: remoteUploads[1], Path: "sub2/b.go", TargetCommit: "deadbeef2", TargetRange: testRange2},
		{Dump: remoteUploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef2", TargetRange: testRange3},
		{Dump: remoteUploads[1], Path: "sub2/b.go", TargetCommit: "deadbeef2", TargetRange: testRange4},
		{Dump: remoteUploads[1], Path: "sub2/c.go", TargetCommit: "deadbeef2", TargetRange: testRange5},
	}

	if diff := cmp.Diff(xLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}

	if history := mockUploadSvc.GetDumpsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
		t.Fatalf("unexpected call count for dbstore.DefinitionDump. want=%d have=%d", 1, len(history))
	} else {
		expectedMonikers := []precise.QualifiedMonikerData{
			{MonikerData: monikers[0], PackageInformationData: packageInformation1},
			{MonikerData: monikers[2], PackageInformationData: packageInformation2},
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
			t.Errorf("unexpected monikers (-want +got):\n%s", diff)
		}
	}

	if history := mockLsifStore.GetBulkMonikerLocationsFunc.History(); len(history) != 1 {
		t.Fatalf("unexpected call count for lsifstore.BulkMonikerResults. want=%d have=%d", 1, len(history))
	} else {
		if diff := cmp.Diff([]int{151, 152, 153}, history[0].Arg2); diff != "" {
			t.Errorf("unexpected ids (-want +got):\n%s", diff)
		}

		expectedMonikers := []precise.MonikerData{
			monikers[0],
			monikers[2],
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg3); diff != "" {
			t.Errorf("unexpected ids (-want +got):\n%s", diff)
		}
	}
}

func TestDefinitionsRemoteWithSubRepoPermissions(t *testing.T) {
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
	mockRequestState.GitTreeTranslator = mockedGitTreeTranslator()

	// Applying sub-repo permissions
	checker := authz.NewMockSubRepoPermissionChecker()
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		if content.Path == "sub2/b.go" {
			return authz.Read, nil
		}
		return authz.None, nil
	})
	mockRequestState.SetAuthChecker(checker)

	dumps := []uploadsshared.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockUploadSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(dumps, nil)

	// upload #150's commit no longer exists; all others do
	mockGitserverClient.CommitsExistFunc.SetDefaultHook(func(ctx context.Context, _ authz.SubRepoPermissionChecker, rcs []api.RepoCommit) (exists []bool, _ error) {
		for _, rc := range rcs {
			exists = append(exists, rc.CommitID != "deadbeef1")
		}
		return
	})

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

	locations := []shared.Location{
		{DumpID: 151, Path: "a.go", Range: testRange1},
		{DumpID: 151, Path: "b.go", Range: testRange2},
		{DumpID: 151, Path: "a.go", Range: testRange3},
		{DumpID: 151, Path: "b.go", Range: testRange4},
		{DumpID: 151, Path: "c.go", Range: testRange5},
	}
	mockLsifStore.GetBulkMonikerLocationsFunc.PushReturn(locations, 0, nil)
	mockLsifStore.GetBulkMonikerLocationsFunc.PushReturn(locations, len(locations), nil)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	mockRequest := RequestArgs{
		RepositoryID: 42,
		Commit:       "deadbeef",
		Path:         "s1/main.go",
		Line:         10,
		Character:    20,
	}
	adjustedLocations, err := svc.GetDefinitions(ctx, mockRequest, mockRequestState)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	}

	expectedLocations := []shared.UploadLocation{
		{Dump: dumps[1], Path: "sub2/b.go", TargetCommit: "deadbeef2", TargetRange: testRange2},
		{Dump: dumps[1], Path: "sub2/b.go", TargetCommit: "deadbeef2", TargetRange: testRange4},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}

	if history := mockUploadSvc.GetDumpsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
		t.Fatalf("unexpected call count for dbstore.DefinitionDump. want=%d have=%d", 1, len(history))
	} else {
		expectedMonikers := []precise.QualifiedMonikerData{
			{MonikerData: monikers[0], PackageInformationData: packageInformation1},
			{MonikerData: monikers[2], PackageInformationData: packageInformation2},
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
			t.Errorf("unexpected monikers (-want +got):\n%s", diff)
		}
	}

	if history := mockLsifStore.GetBulkMonikerLocationsFunc.History(); len(history) != 1 {
		t.Fatalf("unexpected call count for lsifstore.BulkMonikerResults. want=%d have=%d", 1, len(history))
	} else {
		if diff := cmp.Diff([]int{151, 152, 153}, history[0].Arg2); diff != "" {
			t.Errorf("unexpected ids (-want +got):\n%s", diff)
		}

		expectedMonikers := []precise.MonikerData{
			monikers[0],
			monikers[2],
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg3); diff != "" {
			t.Errorf("unexpected ids (-want +got):\n%s", diff)
		}
	}
}

func mockedGitTreeTranslator() GitTreeTranslator {
	mockPositionAdjuster := NewMockGitTreeTranslator()
	mockPositionAdjuster.GetTargetCommitPathFromSourcePathFunc.SetDefaultHook(func(ctx context.Context, commit string, path string, _ bool) (string, bool, error) {
		return commit, true, nil
	})
	mockPositionAdjuster.GetTargetCommitPositionFromSourcePositionFunc.SetDefaultHook(func(ctx context.Context, commit string, pos shared.Position, _ bool) (string, shared.Position, bool, error) {
		return commit, pos, true, nil
	})
	mockPositionAdjuster.GetTargetCommitRangeFromSourceRangeFunc.SetDefaultHook(func(ctx context.Context, commit string, path string, rx shared.Range, _ bool) (string, shared.Range, bool, error) {
		return commit, rx, true, nil
	})

	return mockPositionAdjuster
}
