package janitor

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store/mocks"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

func TestRemoveOrphanedUploadFile(t *testing.T) {
	bundleDir := testRoot(t)
	ids := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for _, id := range ids {
		path := filepath.Join(bundleDir, "uploads", fmt.Sprintf("%d.gz", id))
		if err := makeFile(path, time.Now().Local().Add(-2*time.Minute)); err != nil {
			t.Fatalf("unexpected error creating file %s: %s", path, err)
		}
	}

	// Add a new file that should be skipped
	path := filepath.Join(bundleDir, "uploads", "0.gz")
	if err := makeFile(path, time.Now().Local()); err != nil {
		t.Fatalf("unexpected error creating file %s: %s", path, err)
	}

	mockStore := storemocks.NewMockStore()
	mockStore.GetStatesFunc.SetDefaultHook(func(ctx context.Context, ids []int) (map[int]string, error) {
		sort.Ints(ids)
		return map[int]string{
			1:  "completed",
			2:  "queued",
			3:  "completed",
			4:  "processing",
			5:  "completed",
			9:  "errored",
			10: "errored",
		}, nil
	})

	j := &Janitor{
		store:     mockStore,
		bundleDir: bundleDir,
		metrics:   NewJanitorMetrics(metrics.TestRegisterer),
	}

	if err := j.removeOrphanedUploadFiles(context.Background()); err != nil {
		t.Fatalf("unexpected error removing orphaned upload files: %s", err)
	}

	names, err := getFilenames(filepath.Join(bundleDir, "uploads"))
	if err != nil {
		t.Fatalf("unexpected error listing directory: %s", err)
	}

	expectedNames := []string{"0.gz", "1.gz", "2.gz", "3.gz", "4.gz", "5.gz"}
	if diff := cmp.Diff(expectedNames, names); diff != "" {
		t.Errorf("unexpected directory contents (-want +got):\n%s", diff)
	}

	var idArgs [][]int
	for _, call := range mockStore.GetStatesFunc.History() {
		idArgs = append(idArgs, call.Arg1)
	}

	expectedArgs := [][]int{ids}
	if diff := cmp.Diff(expectedArgs, idArgs); diff != "" {
		t.Errorf("unexpected arguments to statesFn (-want +got):\n%s", diff)
	}
}
