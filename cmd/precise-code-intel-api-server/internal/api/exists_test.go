package api

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	bundlemocks "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	dbmocks "github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
	gitservermocks "github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver/mocks"
)

func TestFindClosestDatabase(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient1 := bundlemocks.NewMockBundleClient()
	mockBundleClient2 := bundlemocks.NewMockBundleClient()
	mockBundleClient3 := bundlemocks.NewMockBundleClient()
	mockBundleClient4 := bundlemocks.NewMockBundleClient()
	mockGitserverClient := gitservermocks.NewMockClient()

	setMockDBFindClosestDumps(t, mockDB, 42, testCommit, "s1/main.go", []db.Dump{
		{ID: 50, Root: "s1/"},
		{ID: 51, Root: "s1/"},
		{ID: 52, Root: "s1/"},
		{ID: 53, Root: "s2/"},
	})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{
		50: mockBundleClient1,
		51: mockBundleClient2,
		52: mockBundleClient3,
		53: mockBundleClient4,
	})
	setMockBundleClientExists(t, mockBundleClient1, "main.go", true)
	setMockBundleClientExists(t, mockBundleClient2, "main.go", false)
	setMockBundleClientExists(t, mockBundleClient3, "main.go", true)
	setMockBundleClientExists(t, mockBundleClient4, "s1/main.go", false)

	// Set a different tip commit
	mockGitserverClient.HeadFunc.SetDefaultReturn(makeCommit(30), nil)

	// Return some ancestors for each commit args
	mockGitserverClient.CommitsNearFunc.SetDefaultHook(func(ctx context.Context, db db.DB, repositoryID int, commit string) (map[string][]string, error) {
		offset, err := strconv.ParseInt(commit, 10, 64)
		if err != nil {
			return nil, err
		}

		commits := map[string][]string{}
		for i := 0; i < 10; i++ {
			commits[makeCommit(int(offset)+i)] = []string{makeCommit(int(offset) + i + 1)}
		}

		return commits, nil
	})

	api := testAPI(mockDB, mockBundleManagerClient, mockGitserverClient)
	dumps, err := api.FindClosestDumps(context.Background(), 42, testCommit, "s1/main.go")
	if err != nil {
		t.Fatalf("unexpected error finding closest database: %s", err)
	}

	expected := []db.Dump{
		{ID: 50, Root: "s1/"},
		{ID: 52, Root: "s1/"},
	}
	if diff := cmp.Diff(expected, dumps); diff != "" {
		t.Errorf("unexpected dumps (-want +got):\n%s", diff)
	}

	expectedCommits := map[string][]string{}
	for i := 0; i < 10; i++ {
		expectedCommits[makeCommit(i)] = []string{makeCommit(i + 1)}
	}
	if len(mockDB.UpdateCommitsFunc.History()) != 1 {
		t.Errorf("unexpected number of update UpdateCommits calls. want=%d have=%d", 1, len(mockDB.UpdateCommitsFunc.History()))
	} else if diff := cmp.Diff(expectedCommits, mockDB.UpdateCommitsFunc.History()[0].Arg2); diff != "" {
		t.Errorf("unexpected update UpdateCommitsFunc args (-want +got):\n%s", diff)
	}

	if len(mockDB.UpdateDumpsVisibleFromTipFunc.History()) != 1 {
		t.Errorf("unexpected number of UpdateDumpsVisibleFromTip calls. want=%d have=%d", 1, len(mockDB.UpdateDumpsVisibleFromTipFunc.History()))
	} else if mockDB.UpdateDumpsVisibleFromTipFunc.History()[0].Arg1 != 42 {
		t.Errorf("unexpected value for repository id. want=%d have=%d", 42, mockDB.UpdateDumpsVisibleFromTipFunc.History()[0].Arg1)
	} else if mockDB.UpdateDumpsVisibleFromTipFunc.History()[0].Arg2 != makeCommit(30) {
		t.Errorf("unexpected value for tip commit. want=%s have=%s", makeCommit(30), mockDB.UpdateDumpsVisibleFromTipFunc.History()[0].Arg2)
	}
}

func TestFindClosestSkipsGitserverIfCommitIsKnown(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient := bundlemocks.NewMockBundleClient()
	mockGitserverClient := gitservermocks.NewMockClient()

	setMockDBHasCommit(t, mockDB, 42, testCommit, true)
	setMockDBFindClosestDumps(t, mockDB, 42, testCommit, "main.go", []db.Dump{
		{ID: 50, Root: ""},
	})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{
		50: mockBundleClient,
	})
	setMockBundleClientExists(t, mockBundleClient, "main.go", true)

	api := New(mockDB, mockBundleManagerClient, mockGitserverClient)
	dumps, err := api.FindClosestDumps(context.Background(), 42, testCommit, "main.go")
	if err != nil {
		t.Fatalf("unexpected error finding closest database: %s", err)
	}

	expected := []db.Dump{{ID: 50, Root: ""}}
	if diff := cmp.Diff(expected, dumps); diff != "" {
		t.Errorf("unexpected dumps (-want +got):\n%s", diff)
	}

	if len(mockGitserverClient.CommitsNearFunc.History()) != 0 {
		t.Errorf("expected gitserverClient.CommitsNear not to be called")
	}
}
