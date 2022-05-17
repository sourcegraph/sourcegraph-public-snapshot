package resolvers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestDefinitions(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()
	mockPositionAdjuster := noopPositionAdjuster()

	locations := []lsifstore.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockLSIFStore.DefinitionsFunc.PushReturn(nil, 0, nil)
	mockLSIFStore.DefinitionsFunc.PushReturn(locations, len(locations), nil)

	uploads := []dbstore.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	resolver := newQueryResolver(
		database.NewMockDB(),
		mockDBStore,
		mockLSIFStore,
		newCachedCommitChecker(mockGitserverClient),
		mockPositionAdjuster,
		42,
		"deadbeef",
		"s1/main.go",
		uploads,
		newOperations(&observation.TestContext),
		authz.NewMockSubRepoPermissionChecker(),
		50,
	)
	adjustedLocations, err := resolver.Definitions(context.Background(), 10, 20)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	}

	expectedLocations := []AdjustedLocation{
		{Dump: uploads[1], Path: "sub2/a.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange1},
		{Dump: uploads[1], Path: "sub2/b.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange2},
		{Dump: uploads[1], Path: "sub2/a.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange3},
		{Dump: uploads[1], Path: "sub2/b.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange4},
		{Dump: uploads[1], Path: "sub2/c.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange5},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}

func TestDefinitionsWithSubRepoPermissions(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()
	mockPositionAdjuster := noopPositionAdjuster()

	locations := []lsifstore.Location{
		{DumpID: 51, Path: "a.go", Range: testRange1},
		{DumpID: 51, Path: "b.go", Range: testRange2},
		{DumpID: 51, Path: "a.go", Range: testRange3},
		{DumpID: 51, Path: "b.go", Range: testRange4},
		{DumpID: 51, Path: "c.go", Range: testRange5},
	}
	mockLSIFStore.DefinitionsFunc.PushReturn(nil, 0, nil)
	mockLSIFStore.DefinitionsFunc.PushReturn(locations, len(locations), nil)

	uploads := []dbstore.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}

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

	resolver := newQueryResolver(
		database.NewMockDB(),
		mockDBStore,
		mockLSIFStore,
		newCachedCommitChecker(mockGitserverClient),
		mockPositionAdjuster,
		42,
		"deadbeef",
		"s1/main.go",
		uploads,
		newOperations(&observation.TestContext),
		checker,
		50,
	)

	ctx := context.Background()
	adjustedLocations, err := resolver.Definitions(actor.WithActor(ctx, &actor.Actor{UID: 1}), 10, 20)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	}

	expectedLocations := []AdjustedLocation{
		{Dump: uploads[1], Path: "sub2/a.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange1},
		{Dump: uploads[1], Path: "sub2/a.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange3},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}

func TestDefinitionsRemote(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()
	mockPositionAdjuster := noopPositionAdjuster()

	remoteUploads := []dbstore.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockDBStore.DefinitionDumpsFunc.PushReturn(remoteUploads, nil)

	// upload #150's commit no longer exists; all others do
	mockGitserverClient.CommitsExistFunc.SetDefaultHook(func(ctx context.Context, rcs []gitserver.RepositoryCommit) (exists []bool, _ error) {
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
	mockLSIFStore.MonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[0]}}, nil)
	mockLSIFStore.MonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[1]}}, nil)
	mockLSIFStore.MonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[2]}}, nil)
	mockLSIFStore.MonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[3]}}, nil)

	packageInformation1 := precise.PackageInformationData{Name: "leftpad", Version: "0.1.0"}
	packageInformation2 := precise.PackageInformationData{Name: "leftpad", Version: "0.2.0"}
	mockLSIFStore.PackageInformationFunc.PushReturn(packageInformation1, true, nil)
	mockLSIFStore.PackageInformationFunc.PushReturn(packageInformation2, true, nil)

	locations := []lsifstore.Location{
		{DumpID: 151, Path: "a.go", Range: testRange1},
		{DumpID: 151, Path: "b.go", Range: testRange2},
		{DumpID: 151, Path: "a.go", Range: testRange3},
		{DumpID: 151, Path: "b.go", Range: testRange4},
		{DumpID: 151, Path: "c.go", Range: testRange5},
	}
	mockLSIFStore.BulkMonikerResultsFunc.PushReturn(locations, 0, nil)
	mockLSIFStore.BulkMonikerResultsFunc.PushReturn(locations, len(locations), nil)

	uploads := []dbstore.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	resolver := newQueryResolver(
		database.NewMockDB(),
		mockDBStore,
		mockLSIFStore,
		newCachedCommitChecker(mockGitserverClient),
		mockPositionAdjuster,
		42,
		"deadbeef",
		"s1/main.go",
		uploads,
		newOperations(&observation.TestContext),
		authz.NewMockSubRepoPermissionChecker(),
		50,
	)
	adjustedLocations, err := resolver.Definitions(context.Background(), 10, 20)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	}

	expectedLocations := []AdjustedLocation{
		{Dump: remoteUploads[0], Path: "sub2/a.go", AdjustedCommit: "deadbeef2", AdjustedRange: testRange1},
		{Dump: remoteUploads[0], Path: "sub2/b.go", AdjustedCommit: "deadbeef2", AdjustedRange: testRange2},
		{Dump: remoteUploads[0], Path: "sub2/a.go", AdjustedCommit: "deadbeef2", AdjustedRange: testRange3},
		{Dump: remoteUploads[0], Path: "sub2/b.go", AdjustedCommit: "deadbeef2", AdjustedRange: testRange4},
		{Dump: remoteUploads[0], Path: "sub2/c.go", AdjustedCommit: "deadbeef2", AdjustedRange: testRange5},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}

	if history := mockDBStore.DefinitionDumpsFunc.History(); len(history) != 1 {
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

	if history := mockLSIFStore.BulkMonikerResultsFunc.History(); len(history) != 1 {
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
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()
	mockPositionAdjuster := noopPositionAdjuster()

	remoteUploads := []dbstore.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockDBStore.DefinitionDumpsFunc.PushReturn(remoteUploads, nil)

	// upload #150's commit no longer exists; all others do
	mockGitserverClient.CommitsExistFunc.SetDefaultHook(func(ctx context.Context, rcs []gitserver.RepositoryCommit) (exists []bool, _ error) {
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
	mockLSIFStore.MonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[0]}}, nil)
	mockLSIFStore.MonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[1]}}, nil)
	mockLSIFStore.MonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[2]}}, nil)
	mockLSIFStore.MonikersByPositionFunc.PushReturn([][]precise.MonikerData{{monikers[3]}}, nil)

	packageInformation1 := precise.PackageInformationData{Name: "leftpad", Version: "0.1.0"}
	packageInformation2 := precise.PackageInformationData{Name: "leftpad", Version: "0.2.0"}
	mockLSIFStore.PackageInformationFunc.PushReturn(packageInformation1, true, nil)
	mockLSIFStore.PackageInformationFunc.PushReturn(packageInformation2, true, nil)

	locations := []lsifstore.Location{
		{DumpID: 151, Path: "a.go", Range: testRange1},
		{DumpID: 151, Path: "b.go", Range: testRange2},
		{DumpID: 151, Path: "a.go", Range: testRange3},
		{DumpID: 151, Path: "b.go", Range: testRange4},
		{DumpID: 151, Path: "c.go", Range: testRange5},
	}
	mockLSIFStore.BulkMonikerResultsFunc.PushReturn(locations, 0, nil)
	mockLSIFStore.BulkMonikerResultsFunc.PushReturn(locations, len(locations), nil)

	uploads := []dbstore.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}

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

	resolver := newQueryResolver(
		database.NewMockDB(),
		mockDBStore,
		mockLSIFStore,
		newCachedCommitChecker(mockGitserverClient),
		mockPositionAdjuster,
		42,
		"deadbeef",
		"s1/main.go",
		uploads,
		newOperations(&observation.TestContext),
		checker,
		50,
	)

	ctx := context.Background()
	adjustedLocations, err := resolver.Definitions(actor.WithActor(ctx, &actor.Actor{UID: 1}), 10, 20)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	}

	expectedLocations := []AdjustedLocation{
		{Dump: remoteUploads[0], Path: "sub2/b.go", AdjustedCommit: "deadbeef2", AdjustedRange: testRange2},
		{Dump: remoteUploads[0], Path: "sub2/b.go", AdjustedCommit: "deadbeef2", AdjustedRange: testRange4},
	}
	if diff := cmp.Diff(expectedLocations, adjustedLocations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}

	if history := mockDBStore.DefinitionDumpsFunc.History(); len(history) != 1 {
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

	if history := mockLSIFStore.BulkMonikerResultsFunc.History(); len(history) != 1 {
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
