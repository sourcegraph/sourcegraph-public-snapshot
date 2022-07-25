package graphql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	codeintelgitserver "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

var (
	testRange1 = shared.Range{Start: shared.Position{Line: 11, Character: 21}, End: shared.Position{Line: 31, Character: 41}}
	testRange2 = shared.Range{Start: shared.Position{Line: 12, Character: 22}, End: shared.Position{Line: 32, Character: 42}}
	testRange3 = shared.Range{Start: shared.Position{Line: 13, Character: 23}, End: shared.Position{Line: 33, Character: 43}}
	testRange4 = shared.Range{Start: shared.Position{Line: 14, Character: 24}, End: shared.Position{Line: 34, Character: 44}}
	testRange5 = shared.Range{Start: shared.Position{Line: 15, Character: 25}, End: shared.Position{Line: 35, Character: 45}}

	mockPath   = "s1/main.go"
	mockCommit = "deadbeef"
)

func TestDefinitions(t *testing.T) {
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

	locations := []shared.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockSvc.GetDefinitionsFunc.PushReturn(locations, len(locations), nil)

	uploads := []store.Dump{
		{ID: 50, Commit: mockCommit, Root: "sub1/"},
		{ID: 51, Commit: mockCommit, Root: "sub2/"},
		{ID: 52, Commit: mockCommit, Root: "sub3/"},
		{ID: 53, Commit: mockCommit, Root: "sub4/"},
	}
	resolver.SetUploadsDataLoader(uploads)

	mockRequest := shared.RequestArgs{
		RepositoryID: 51,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
	}
	adjustedLocations, err := resolver.Definitions(context.Background(), mockRequest)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	}
	sharedUploads := storeDumpToSymbolDump(uploads)
	expectedLocations := []shared.UploadLocation{
		{Dump: sharedUploads[1], Path: "sub2/a.go", TargetCommit: mockCommit, TargetRange: testRange1},
		{Dump: sharedUploads[1], Path: "sub2/b.go", TargetCommit: mockCommit, TargetRange: testRange2},
		{Dump: sharedUploads[1], Path: "sub2/a.go", TargetCommit: mockCommit, TargetRange: testRange3},
		{Dump: sharedUploads[1], Path: "sub2/b.go", TargetCommit: mockCommit, TargetRange: testRange4},
		{Dump: sharedUploads[1], Path: "sub2/c.go", TargetCommit: mockCommit, TargetRange: testRange5},
	}

	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}

func TestDefinitionsWithSubRepoPermissions(t *testing.T) {
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

	uploads := []dbstore.Dump{
		{ID: 50, Commit: mockCommit, Root: "sub1/"},
		{ID: 51, Commit: mockCommit, Root: "sub2/"},
		{ID: 52, Commit: mockCommit, Root: "sub3/"},
		{ID: 53, Commit: mockCommit, Root: "sub4/"},
	}
	resolver.SetUploadsDataLoader(uploads)

	locations := []shared.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockSvc.GetDefinitionsFunc.PushReturn(locations, len(locations), nil)

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
	resolver.SetAuthChecker(checker)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	mockRequest := shared.RequestArgs{
		RepositoryID: 51,
		Commit:       "deadbeef",
		Path:         "s1/main.go",
		Line:         10,
		Character:    20,
	}
	adjustedLocations, err := resolver.Definitions(ctx, mockRequest)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	}
	sharedUploads := storeDumpToSymbolDump(uploads)
	expectedLocations := []shared.UploadLocation{
		{Dump: sharedUploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: sharedUploads[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}

func TestDefinitionsRemote(t *testing.T) {
	// Set up mocks
	mockLogger := logtest.Scoped(t)
	mockDB := database.NewDB(mockLogger, dbtest.NewDB(mockLogger, t))
	mockGitServer := gitserver.NewClient(mockDB)
	mockGitserverClient := NewMockGitserverClient()
	mockSvc := NewMockService()

	// Init resolver and set local request context
	resolver := New(mockSvc, 50, &observation.TestContext)
	resolver.SetLocalCommitCache(mockGitserverClient)
	err := resolver.SetLocalGitTreeTranslator(mockGitServer, &types.Repo{ID: 42}, mockCommit, mockPath)
	if err != nil {
		t.Fatalf("unexpected error setting local git tree translator: %s", err)
	}
	resolver.GitTreeTranslator = mockedGitTreeTranslator()

	dumps := []dbstore.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	remoteUploads := storeDumpToSymbolDump(dumps)
	mockSvc.GetUploadsWithDefinitionsForMonikersFunc.PushReturn(remoteUploads, nil)

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
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[0]}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[1]}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[2]}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[3]}}, nil)

	packageInformation1 := precise.PackageInformationData{Name: "leftpad", Version: "0.1.0"}
	packageInformation2 := precise.PackageInformationData{Name: "leftpad", Version: "0.2.0"}
	mockSvc.GetPackageInformationFunc.PushReturn(packageInformation1, true, nil)
	mockSvc.GetPackageInformationFunc.PushReturn(packageInformation2, true, nil)

	locations := []shared.Location{
		{DumpID: 151, Path: "a.go", Range: testRange1},
		{DumpID: 151, Path: "b.go", Range: testRange2},
		{DumpID: 151, Path: "a.go", Range: testRange3},
		{DumpID: 151, Path: "b.go", Range: testRange4},
		{DumpID: 151, Path: "c.go", Range: testRange5},
	}
	mockSvc.GetBulkMonikerLocationsFunc.PushReturn(locations, len(locations), nil)

	uploads := []dbstore.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	resolver.SetUploadsDataLoader(uploads)
	mockRequest := shared.RequestArgs{
		RepositoryID: 42,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
	}
	adjustedLocations, err := resolver.Definitions(context.Background(), mockRequest)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	}

	xLocations := []shared.UploadLocation{
		{Dump: remoteUploads[0], Path: "sub2/a.go", TargetCommit: "deadbeef2", TargetRange: testRange1},
		{Dump: remoteUploads[0], Path: "sub2/b.go", TargetCommit: "deadbeef2", TargetRange: testRange2},
		{Dump: remoteUploads[0], Path: "sub2/a.go", TargetCommit: "deadbeef2", TargetRange: testRange3},
		{Dump: remoteUploads[0], Path: "sub2/b.go", TargetCommit: "deadbeef2", TargetRange: testRange4},
		{Dump: remoteUploads[0], Path: "sub2/c.go", TargetCommit: "deadbeef2", TargetRange: testRange5},
	}
	expectedLocations := uploadLocationsToAdjustedLocations(xLocations)
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}

	if history := mockSvc.GetUploadsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
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

	if history := mockSvc.GetBulkMonikerLocationsFunc.History(); len(history) != 1 {
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

func TestDefinitionsRemoteWithSubRepoPermissions(t *testing.T) {
	// Set up mocks
	mockLogger := logtest.Scoped(t)
	mockDB := database.NewDB(mockLogger, dbtest.NewDB(mockLogger, t))
	mockGitServer := gitserver.NewClient(mockDB)
	mockGitserverClient := NewMockGitserverClient()
	mockSvc := NewMockService()

	// Init resolver and set local request context
	resolver := New(mockSvc, 50, &observation.TestContext)
	resolver.SetLocalCommitCache(mockGitserverClient)
	err := resolver.SetLocalGitTreeTranslator(mockGitServer, &types.Repo{ID: 42}, mockCommit, mockPath)
	if err != nil {
		t.Fatalf("unexpected error setting local git tree translator: %s", err)
	}
	resolver.GitTreeTranslator = mockedGitTreeTranslator()

	dumps := []dbstore.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	remoteUploads := storeDumpToSymbolDump(dumps)
	mockSvc.GetUploadsWithDefinitionsForMonikersFunc.PushReturn(remoteUploads, nil)

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
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[0]}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[1]}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[2]}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[3]}}, nil)

	packageInformation1 := precise.PackageInformationData{Name: "leftpad", Version: "0.1.0"}
	packageInformation2 := precise.PackageInformationData{Name: "leftpad", Version: "0.2.0"}
	mockSvc.GetPackageInformationFunc.PushReturn(packageInformation1, true, nil)
	mockSvc.GetPackageInformationFunc.PushReturn(packageInformation2, true, nil)

	locations := []shared.Location{
		{DumpID: 151, Path: "a.go", Range: testRange1},
		{DumpID: 151, Path: "b.go", Range: testRange2},
		{DumpID: 151, Path: "a.go", Range: testRange3},
		{DumpID: 151, Path: "b.go", Range: testRange4},
		{DumpID: 151, Path: "c.go", Range: testRange5},
	}
	mockSvc.GetBulkMonikerLocationsFunc.PushReturn(locations, 0, nil)
	mockSvc.GetBulkMonikerLocationsFunc.PushReturn(locations, len(locations), nil)

	uploads := []dbstore.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	resolver.SetUploadsDataLoader(uploads)

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
	resolver.SetAuthChecker(checker)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	mockRequest := shared.RequestArgs{
		RepositoryID: 42,
		Commit:       "deadbeef",
		Path:         "s1/main.go",
		Line:         10,
		Character:    20,
	}
	adjustedLocations, err := resolver.Definitions(ctx, mockRequest)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	}
	expectedLocations := []shared.UploadLocation{
		{Dump: remoteUploads[0], Path: "sub2/b.go", TargetCommit: "deadbeef2", TargetRange: testRange2},
		{Dump: remoteUploads[0], Path: "sub2/b.go", TargetCommit: "deadbeef2", TargetRange: testRange4},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}

	if history := mockSvc.GetUploadsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
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

	if history := mockSvc.GetBulkMonikerLocationsFunc.History(); len(history) != 1 {
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

func storeDumpToSymbolDump(storeDumps []store.Dump) []shared.Dump {
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
