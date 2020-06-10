package indexer

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	dbmocks "github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
	gitservermocks "github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver/mocks"
)

func TestFetchRepository(t *testing.T) {
	tarfile, err := os.Open("./testdata/clone.tar")
	if err != nil {
		t.Fatalf("unexpected opening test tarfile: %s", err)
	}
	defer tarfile.Close()

	mockDB := dbmocks.NewMockDB()
	mockGitserverClient := gitservermocks.NewMockClient()
	mockGitserverClient.ArchiveFunc.SetDefaultReturn(tarfile, nil)

	tempDir, err := fetchRepository(context.Background(), mockDB, mockGitserverClient, 50, "deadbeef")
	if err != nil {
		t.Fatalf("unexpected error fetching repository: %s", err)
	}
	defer os.RemoveAll(tempDir)

	sizes, err := readFiles(tempDir)
	if err != nil {
		t.Fatalf("unexpected reading directory: %s", err)
	}

	if diff := cmp.Diff(expectedCloneTarSizes, sizes); diff != "" {
		t.Errorf("unexpected commits (-want +got):\n%s", diff)
	}
}
