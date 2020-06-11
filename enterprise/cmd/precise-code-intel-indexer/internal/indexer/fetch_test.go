package indexer

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	gitservermocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver/mocks"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store/mocks"
)

func TestFetchRepository(t *testing.T) {
	tarfile, err := os.Open("./testdata/clone.tar")
	if err != nil {
		t.Fatalf("unexpected opening test tarfile: %s", err)
	}
	defer tarfile.Close()

	mockStore := storemocks.NewMockStore()
	mockGitserverClient := gitservermocks.NewMockClient()
	mockGitserverClient.ArchiveFunc.SetDefaultReturn(tarfile, nil)

	tempDir, err := fetchRepository(context.Background(), mockStore, mockGitserverClient, 50, "deadbeef")
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
