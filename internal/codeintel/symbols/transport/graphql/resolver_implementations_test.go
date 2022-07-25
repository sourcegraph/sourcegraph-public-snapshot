package graphql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	codeintelgitserver "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestImplementations(t *testing.T) {
	// mockDBStore := NewMockDBStore()
	// mockLSIFStore := NewMockLSIFStore()
	// mockGitserverClient := NewMockGitserverClient()
	// mockPositionAdjuster := noopPositionAdjuster()
	// mockSymbolsResolver := NewMockSymbolsResolver()
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

	// Empty result set (prevents nil pointer as scanner is always non-nil)
	mockSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{}, 0, 0, nil)

	locations := []shared.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockSvc.GetImplementationsFunc.PushReturn(locations[:1], 1, nil)
	mockSvc.GetImplementationsFunc.PushReturn(locations[1:4], 3, nil)
	mockSvc.GetImplementationsFunc.PushReturn(locations[4:], 1, nil)

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
		RepositoryID: 51,
		Commit:       "deadbeef",
		Path:         "s1/main.go",
		Line:         10,
		Character:    20,
		Limit:        50,
	}
	adjustedLocations, _, err := resolver.Implementations(context.Background(), mockRequest)
	if err != nil {
		t.Fatalf("unexpected error querying implementations: %s", err)
	}
	u := storeDumpToSymbolDump(uploads)
	expectedLocations := []shared.UploadLocation{
		{Dump: u[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: u[1], Path: "sub2/b.go", TargetCommit: "deadbeef", TargetRange: testRange2},
		{Dump: u[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
		{Dump: u[1], Path: "sub2/b.go", TargetCommit: "deadbeef", TargetRange: testRange4},
		{Dump: u[1], Path: "sub2/c.go", TargetCommit: "deadbeef", TargetRange: testRange5},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}

func TestImplementationsWithSubRepoPermissions(t *testing.T) {
	// mockDBStore := NewMockDBStore()
	// mockLSIFStore := NewMockLSIFStore()
	// mockGitserverClient := NewMockGitserverClient()
	// mockPositionAdjuster := noopPositionAdjuster()
	// mockSymbolsResolver := NewMockSymbolsResolver()
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

	// Empty result set (prevents nil pointer as scanner is always non-nil)
	mockSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{}, 0, 0, nil)

	locations := []shared.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockSvc.GetImplementationsFunc.PushReturn(locations[:1], 1, nil)
	mockSvc.GetImplementationsFunc.PushReturn(locations[1:4], 3, nil)
	mockSvc.GetImplementationsFunc.PushReturn(locations[4:], 1, nil)

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
		if content.Path == "sub2/a.go" {
			return authz.Read, nil
		}
		return authz.None, nil
	})
	resolver.SetAuthChecker(checker)

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
	// 	checker,
	// 	50,
	// 	mockSymbolsResolver,
	// )

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	mockRequest := shared.RequestArgs{
		RepositoryID: 42,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
		Limit:        50,
	}
	adjustedLocations, _, err := resolver.Implementations(ctx, mockRequest)
	if err != nil {
		t.Fatalf("unexpected error querying implementations: %s", err)
	}
	u := storeDumpToSymbolDump(uploads)
	expectedLocations := []shared.UploadLocation{
		{Dump: u[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: u[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}

func TestImplementationsRemote(t *testing.T) {
	// mockDBStore := NewMockDBStore()
	// mockLSIFStore := NewMockLSIFStore()
	// mockGitserverClient := NewMockGitserverClient()
	// mockPositionAdjuster := noopPositionAdjuster()
	// mockSymbolsResolver := NewMockSymbolsResolver()

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

	referenceUploads := []shared.Dump{
		{ID: 250, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 251, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 252, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 253, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockSvc.GetDumpsByIDsFunc.PushReturn(referenceUploads[:2], nil)
	mockSvc.GetDumpsByIDsFunc.PushReturn(referenceUploads[2:], nil)

	// scanner1 := dbstore.PackageReferenceScannerFromSlice(
	// 	shared.PackageReference{Package: shared.Package{DumpID: 250}},
	// 	shared.PackageReference{Package: shared.Package{DumpID: 251}},
	// )
	// scanner2 := dbstore.PackageReferenceScannerFromSlice(
	// 	shared.PackageReference{Package: shared.Package{DumpID: 252}},
	// 	shared.PackageReference{Package: shared.Package{DumpID: 253}},
	// )
	// mockDBStore.ReferenceIDsFunc.PushReturn(scanner1, 4, nil)
	// mockDBStore.ReferenceIDsFunc.PushReturn(scanner2, 2, nil)
	mockSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{250, 251}, 0, 4, nil)
	mockSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{252, 253}, 0, 2, nil)

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
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[0]}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[1]}}, nil)

	packageInformation1 := precise.PackageInformationData{Name: "leftpad", Version: "0.1.0"}
	packageInformation2 := precise.PackageInformationData{Name: "leftpad", Version: "0.2.0"}
	mockSvc.GetPackageInformationFunc.PushReturn(packageInformation1, true, nil)
	mockSvc.GetPackageInformationFunc.PushReturn(packageInformation2, true, nil)

	locations := []shared.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockSvc.GetImplementationsFunc.PushReturn(locations[:1], 1, nil)
	mockSvc.GetImplementationsFunc.PushReturn(locations[1:4], 3, nil)
	mockSvc.GetImplementationsFunc.PushReturn(locations[4:5], 1, nil)

	monikerLocations := []shared.Location{
		{DumpID: 53, Path: "a.go", Range: testRange1},
		{DumpID: 53, Path: "b.go", Range: testRange2},
		{DumpID: 53, Path: "a.go", Range: testRange3},
		{DumpID: 53, Path: "b.go", Range: testRange4},
		{DumpID: 53, Path: "c.go", Range: testRange5},
	}
	mockSvc.GetBulkMonikerLocationsFunc.PushReturn(monikerLocations[0:1], 1, nil) // defs
	mockSvc.GetBulkMonikerLocationsFunc.PushReturn(monikerLocations[1:2], 1, nil) // impls batch 1
	mockSvc.GetBulkMonikerLocationsFunc.PushReturn(monikerLocations[2:], 3, nil)  // impls batch 2

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
	adjustedLocations, _, err := resolver.Implementations(context.Background(), mockRequest)
	if err != nil {
		t.Fatalf("unexpected error querying references: %s", err)
	}

	u := storeDumpToSymbolDump(uploads)
	expectedLocations := []shared.UploadLocation{
		{Dump: u[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: u[1], Path: "sub2/b.go", TargetCommit: "deadbeef", TargetRange: testRange2},
		{Dump: u[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
		{Dump: u[1], Path: "sub2/b.go", TargetCommit: "deadbeef", TargetRange: testRange4},
		{Dump: u[1], Path: "sub2/c.go", TargetCommit: "deadbeef", TargetRange: testRange5},
		{Dump: u[3], Path: "sub4/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: u[3], Path: "sub4/b.go", TargetCommit: "deadbeef", TargetRange: testRange2},
		{Dump: u[3], Path: "sub4/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
		{Dump: u[3], Path: "sub4/b.go", TargetCommit: "deadbeef", TargetRange: testRange4},
		{Dump: u[3], Path: "sub4/c.go", TargetCommit: "deadbeef", TargetRange: testRange5},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}

	if history := mockSvc.GetUploadsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
		t.Fatalf("unexpected call count for dbstore.DefinitionDump. want=%d have=%d", 1, len(history))
	} else {
		expectedMonikers := []precise.QualifiedMonikerData{
			{MonikerData: monikers[0], PackageInformationData: packageInformation1},
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
			t.Errorf("unexpected monikers (-want +got):\n%s", diff)
		}
	}

	if history := mockSvc.GetBulkMonikerLocationsFunc.History(); len(history) != 3 {
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
	// mockDBStore := NewMockDBStore()
	// mockLSIFStore := NewMockLSIFStore()
	// mockGitserverClient := NewMockGitserverClient()
	// mockPositionAdjuster := noopPositionAdjuster()
	// mockSymbolsResolver := NewMockSymbolsResolver()
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

	definitionUploads := []shared.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	// definitionUploads := storeDumpToSymbolDump(definitionDumps)
	// mockSymbolsResolver.GetUploadsWithDefinitionsForMonikersFunc.PushReturn(definitionUploads, nil)
	mockSvc.GetUploadsWithDefinitionsForMonikersFunc.PushReturn(definitionUploads, nil)

	referenceUploads := []shared.Dump{
		{ID: 250, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 251, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 252, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 253, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockSvc.GetDumpsByIDsFunc.PushReturn(referenceUploads[:2], nil)
	mockSvc.GetDumpsByIDsFunc.PushReturn(referenceUploads[2:], nil)

	// scanner1 := dbstore.PackageReferenceScannerFromSlice(
	// 	shared.PackageReference{Package: shared.Package{DumpID: 250}},
	// 	shared.PackageReference{Package: shared.Package{DumpID: 251}},
	// )
	// scanner2 := dbstore.PackageReferenceScannerFromSlice(
	// 	shared.PackageReference{Package: shared.Package{DumpID: 252}},
	// 	shared.PackageReference{Package: shared.Package{DumpID: 253}},
	// )
	// mockDBStore.ReferenceIDsFunc.PushReturn(scanner1, 4, nil)
	// mockDBStore.ReferenceIDsFunc.PushReturn(scanner2, 2, nil)

	mockSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{250, 251}, 0, 4, nil)
	mockSvc.GetUploadIDsWithReferencesFunc.PushReturn([]int{252, 253}, 0, 2, nil)

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
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[0]}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{}}, nil)
	mockSvc.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[1]}}, nil)

	packageInformation1 := precise.PackageInformationData{Name: "leftpad", Version: "0.1.0"}
	packageInformation2 := precise.PackageInformationData{Name: "leftpad", Version: "0.2.0"}
	mockSvc.GetPackageInformationFunc.PushReturn(packageInformation1, true, nil)
	mockSvc.GetPackageInformationFunc.PushReturn(packageInformation2, true, nil)

	locations := []shared.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockSvc.GetImplementationsFunc.PushReturn(locations[:1], 1, nil)
	mockSvc.GetImplementationsFunc.PushReturn(locations[1:4], 3, nil)
	mockSvc.GetImplementationsFunc.PushReturn(locations[4:5], 1, nil)

	monikerLocations := []shared.Location{
		{DumpID: 53, Path: "a.go", Range: testRange1},
		{DumpID: 53, Path: "b.go", Range: testRange2},
		{DumpID: 53, Path: "a.go", Range: testRange3},
		{DumpID: 53, Path: "b.go", Range: testRange4},
		{DumpID: 53, Path: "c.go", Range: testRange5},
	}
	mockSvc.GetBulkMonikerLocationsFunc.PushReturn(monikerLocations[0:1], 1, nil) // defs
	mockSvc.GetBulkMonikerLocationsFunc.PushReturn(monikerLocations[1:2], 1, nil) // impls batch 1
	mockSvc.GetBulkMonikerLocationsFunc.PushReturn(monikerLocations[2:], 3, nil)  // impls batch 2

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
		if content.Path == "sub2/a.go" || content.Path == "sub4/a.go" {
			return authz.Read, nil
		}
		return authz.None, nil
	})
	resolver.SetAuthChecker(checker)

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
	// 	checker,
	// 	50,
	// 	mockSymbolsResolver,
	// )

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: 1})
	mockRequest := shared.RequestArgs{
		RepositoryID: 42,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
		Limit:        50,
	}
	adjustedLocations, _, err := resolver.Implementations(ctx, mockRequest)
	if err != nil {
		t.Fatalf("unexpected error querying references: %s", err)
	}
	u := storeDumpToSymbolDump(uploads)
	expectedLocations := []shared.UploadLocation{
		{Dump: u[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: u[1], Path: "sub2/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
		{Dump: u[3], Path: "sub4/a.go", TargetCommit: "deadbeef", TargetRange: testRange1},
		{Dump: u[3], Path: "sub4/a.go", TargetCommit: "deadbeef", TargetRange: testRange3},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}

	if history := mockSvc.GetUploadsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
		t.Fatalf("unexpected call count for dbstore.DefinitionDump. want=%d have=%d", 1, len(history))
	} else {
		expectedMonikers := []precise.QualifiedMonikerData{
			{MonikerData: monikers[0], PackageInformationData: packageInformation1},
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
			t.Errorf("unexpected monikers (-want +got):\n%s", diff)
		}
	}

	if history := mockSvc.GetBulkMonikerLocationsFunc.History(); len(history) != 3 {
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
