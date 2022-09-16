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

func TestDefinitions(t *testing.T) {
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

	mockRequest := shared.RequestArgs{
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
	mockRequest := shared.RequestArgs{
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
	err := mockRequestState.SetLocalGitTreeTranslator(mockGitServer, &types.Repo{ID: 42}, mockCommit, mockPath, 50)
	if err != nil {
		t.Fatalf("unexpected error setting local git tree translator: %s", err)
	}
	mockRequestState.GitTreeTranslator = mockedGitTreeTranslator()
	uploads := []shared.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	mockRequestState.SetUploadsDataLoader(uploads)

	dumps := []uploadsShared.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockUploadSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(dumps, nil)

	// upload #150's commit no longer exists; all others do
	mockGitserverClient.CommitsExistFunc.SetDefaultHook(func(ctx context.Context, rcs []codeintelgitserver.RepositoryCommit) (exists []bool, _ error) {
		for _, rc := range rcs {
			exists = append(exists, rc.Commit != "deadbeef1")
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

	mockRequest := shared.RequestArgs{
		RepositoryID: 42,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
	}
	remoteUploads := updateSvcDumpToSharedDump(dumps)
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
	expectedLocations := uploadLocationsToAdjustedLocations(xLocations)
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

func TestDefinitionsRemoteWithSubRepoPermissions(t *testing.T) {
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

	dumps := []uploadsShared.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockUploadSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(dumps, nil)

	// upload #150's commit no longer exists; all others do
	mockGitserverClient.CommitsExistFunc.SetDefaultHook(func(ctx context.Context, rcs []codeintelgitserver.RepositoryCommit) (exists []bool, _ error) {
		for _, rc := range rcs {
			exists = append(exists, rc.Commit != "deadbeef1")
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
	mockRequest := shared.RequestArgs{
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
	remoteUploads := uploadDumpToCodeNavDump(dumps)
	expectedLocations := []shared.UploadLocation{
		{Dump: remoteUploads[1], Path: "sub2/b.go", TargetCommit: "deadbeef2", TargetRange: testRange2},
		{Dump: remoteUploads[1], Path: "sub2/b.go", TargetCommit: "deadbeef2", TargetRange: testRange4},
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

func uploadLocationsToAdjustedLocations(location []shared.UploadLocation) []shared.UploadLocation {
	uploadLocation := make([]shared.UploadLocation, 0, len(location))
	for _, loc := range location {
		dump := shared.Dump{
			ID:                loc.Dump.ID,
			Commit:            loc.Dump.Commit,
			Root:              loc.Dump.Root,
			VisibleAtTip:      loc.Dump.VisibleAtTip,
			UploadedAt:        loc.Dump.UploadedAt,
			State:             loc.Dump.State,
			FailureMessage:    loc.Dump.FailureMessage,
			StartedAt:         loc.Dump.StartedAt,
			FinishedAt:        loc.Dump.FinishedAt,
			ProcessAfter:      loc.Dump.ProcessAfter,
			NumResets:         loc.Dump.NumResets,
			NumFailures:       loc.Dump.NumFailures,
			RepositoryID:      loc.Dump.RepositoryID,
			RepositoryName:    loc.Dump.RepositoryName,
			Indexer:           loc.Dump.Indexer,
			IndexerVersion:    loc.Dump.IndexerVersion,
			AssociatedIndexID: loc.Dump.AssociatedIndexID,
		}

		targetRange := shared.Range{
			Start: shared.Position{
				Line:      loc.TargetRange.Start.Line,
				Character: loc.TargetRange.Start.Character,
			},
			End: shared.Position{
				Line:      loc.TargetRange.End.Line,
				Character: loc.TargetRange.End.Character,
			},
		}

		uploadLocation = append(uploadLocation, shared.UploadLocation{
			Dump:         dump,
			Path:         loc.Path,
			TargetCommit: loc.TargetCommit,
			TargetRange:  targetRange,
		})
	}

	return uploadLocation
}

func uploadDumpToCodeNavDump(storeDumps []uploadsShared.Dump) []shared.Dump {
	dumps := make([]shared.Dump, 0, len(storeDumps))
	for _, d := range storeDumps {
		dumps = append(dumps, shared.Dump{
			ID:                d.ID,
			Commit:            d.Commit,
			Root:              d.Root,
			VisibleAtTip:      d.VisibleAtTip,
			UploadedAt:        d.UploadedAt,
			State:             d.State,
			FailureMessage:    d.FailureMessage,
			StartedAt:         d.StartedAt,
			FinishedAt:        d.FinishedAt,
			ProcessAfter:      d.ProcessAfter,
			NumResets:         d.NumResets,
			NumFailures:       d.NumFailures,
			RepositoryID:      d.RepositoryID,
			RepositoryName:    d.RepositoryName,
			Indexer:           d.Indexer,
			IndexerVersion:    d.IndexerVersion,
			AssociatedIndexID: d.AssociatedIndexID,
		})
	}

	return dumps
}
