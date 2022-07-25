package graphql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

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

func TestHover(t *testing.T) {
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

	expectedRange := shared.Range{
		Start: shared.Position{Line: 10, Character: 10},
		End:   shared.Position{Line: 15, Character: 25},
	}
	mockSvc.GetHoverFunc.PushReturn("", shared.Range{}, false, nil)
	mockSvc.GetHoverFunc.PushReturn("doctext", expectedRange, true, nil)

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
		Limit:        50,
	}
	text, rn, exists, err := resolver.Hover(context.Background(), mockRequest)
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
	mockLogger := logtest.Scoped(t)
	mockDB := database.NewDB(mockLogger, dbtest.NewDB(mockLogger, t))
	mockGitServer := gitserver.NewClient(mockDB)
	mockGitserverClient := NewMockGitserverClient()
	mockSvc := NewMockService()

	// Init resolver and set local request context
	resolver := New(mockSvc, 50, &observation.TestContext)
	resolver.SetLocalCommitCache(mockGitserverClient)
	resolver.SetLocalGitTreeTranslator(mockGitServer, &types.Repo{}, mockCommit, mockPath)

	expectedRange := shared.Range{
		Start: shared.Position{Line: 10, Character: 10},
		End:   shared.Position{Line: 15, Character: 25},
	}
	mockSvc.GetHoverFunc.PushReturn("", expectedRange, true, nil)

	remoteRange := shared.Range{
		Start: shared.Position{Line: 30, Character: 30},
		End:   shared.Position{Line: 35, Character: 45},
	}
	mockSvc.GetHoverFunc.PushReturn("doctext", remoteRange, true, nil)

	uploadsWithDefinitions := []shared.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockSvc.GetUploadsWithDefinitionsForMonikersFunc.PushReturn(uploadsWithDefinitions, nil)

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

	mockGitserverClient.CommitsExistFunc.SetDefaultHook(func(ctx context.Context, rcs []codeintelgitserver.RepositoryCommit) (exists []bool, _ error) {
		for range rcs {
			exists = append(exists, true)
		}
		return
	})

	uploads := []dbstore.Dump{
		{ID: 50, Commit: "deadbeef"},
	}
	resolver.SetUploadsDataLoader(uploads)
	mockRequest := shared.RequestArgs{
		RepositoryID: 42,
		Commit:       mockCommit,
		Path:         mockPath,
		Line:         10,
		Character:    20,
		Limit:        50,
	}
	text, rn, exists, err := resolver.Hover(context.Background(), mockRequest)
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
