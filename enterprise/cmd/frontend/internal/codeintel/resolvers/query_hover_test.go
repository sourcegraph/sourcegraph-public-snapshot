package resolvers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

func TestHover(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()
	mockPositionAdjuster := noopPositionAdjuster()

	expectedRange := lsifstore.Range{
		Start: lsifstore.Position{Line: 10, Character: 10},
		End:   lsifstore.Position{Line: 15, Character: 25},
	}
	mockLSIFStore.HoverFunc.PushReturn("", lsifstore.Range{}, false, nil)
	mockLSIFStore.HoverFunc.PushReturn("doctext", expectedRange, true, nil)

	uploads := []dbstore.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	resolver := newQueryResolver(
		mockDBStore,
		mockLSIFStore,
		newCachedCommitChecker(mockGitserverClient),
		mockPositionAdjuster,
		42,
		"deadbeef",
		"s1/main.go",
		uploads,
		newOperations(&observation.TestContext),
	)
	text, rn, exists, err := resolver.Hover(context.Background(), 10, 20)
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
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()
	mockPositionAdjuster := noopPositionAdjuster()

	expectedRange := lsifstore.Range{
		Start: lsifstore.Position{Line: 10, Character: 10},
		End:   lsifstore.Position{Line: 15, Character: 25},
	}
	mockLSIFStore.HoverFunc.PushReturn("", lsifstore.Range{}, false, nil)
	mockLSIFStore.HoverFunc.PushReturn("doctext", expectedRange, true, nil)

	remoteUploads := []dbstore.Dump{
		{ID: 150, Commit: "deadbeef1", Root: "sub1/"},
		{ID: 151, Commit: "deadbeef2", Root: "sub2/"},
		{ID: 152, Commit: "deadbeef3", Root: "sub3/"},
		{ID: 153, Commit: "deadbeef4", Root: "sub4/"},
	}
	mockDBStore.DefinitionDumpsFunc.PushReturn(remoteUploads, nil)

	// upload #150's commit no longer exists; all others do
	mockGitserverClient.CommitExistsFunc.PushReturn(false, nil)
	mockGitserverClient.CommitExistsFunc.SetDefaultReturn(true, nil)

	monikers := []semantic.MonikerData{
		{Kind: "import", Scheme: "tsc", Identifier: "padLeft", PackageInformationID: "51"},
		{Kind: "export", Scheme: "tsc", Identifier: "pad_left", PackageInformationID: "52"},
		{Kind: "import", Scheme: "tsc", Identifier: "pad-left", PackageInformationID: "53"},
		{Kind: "import", Scheme: "tsc", Identifier: "left_pad"},
	}
	mockLSIFStore.MonikersByPositionFunc.PushReturn([][]semantic.MonikerData{{monikers[0]}}, nil)
	mockLSIFStore.MonikersByPositionFunc.PushReturn([][]semantic.MonikerData{{monikers[1]}}, nil)
	mockLSIFStore.MonikersByPositionFunc.PushReturn([][]semantic.MonikerData{{monikers[2]}}, nil)
	mockLSIFStore.MonikersByPositionFunc.PushReturn([][]semantic.MonikerData{{monikers[3]}}, nil)

	packageInformation1 := semantic.PackageInformationData{Name: "leftpad", Version: "0.1.0"}
	packageInformation2 := semantic.PackageInformationData{Name: "leftpad", Version: "0.2.0"}
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
		mockDBStore,
		mockLSIFStore,
		newCachedCommitChecker(mockGitserverClient),
		mockPositionAdjuster,
		42,
		"deadbeef",
		"s1/main.go",
		uploads,
		newOperations(&observation.TestContext),
	)
	text, rn, exists, err := resolver.Hover(context.Background(), 10, 20)
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
