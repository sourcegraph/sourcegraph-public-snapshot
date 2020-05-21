package janitor

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	dbmocks "github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

func TestRemoveOrphanedBundleFile(t *testing.T) {
	bundleDir := testRoot(t)
	ids := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for _, id := range ids {
		path := filepath.Join(bundleDir, "dbs", fmt.Sprintf("%d.lsif.db", id))
		if err := makeFile(path, time.Now().Local()); err != nil {
			t.Fatalf("unexpected error creating file %s: %s", path, err)
		}
	}

	mockDB := dbmocks.NewMockDB()
	mockDB.GetStatesFunc.SetDefaultHook(func(ctx context.Context, ids []int) (map[int]string, error) {
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
		db:        mockDB,
		bundleDir: bundleDir,
		metrics:   NewJanitorMetrics(metrics.TestRegisterer),
	}

	if err := j.removeOrphanedBundleFiles(); err != nil {
		t.Fatalf("unexpected error removing orphaned bundle files: %s", err)
	}

	names, err := getFilenames(filepath.Join(bundleDir, "dbs"))
	if err != nil {
		t.Fatalf("unexpected error listing directory: %s", err)
	}

	expectedNames := []string{"1.lsif.db", "2.lsif.db", "3.lsif.db", "4.lsif.db", "5.lsif.db"}
	if diff := cmp.Diff(expectedNames, names); diff != "" {
		t.Errorf("unexpected directory contents (-want +got):\n%s", diff)
	}

	var idArgs [][]int
	for _, call := range mockDB.GetStatesFunc.History() {
		idArgs = append(idArgs, call.Arg1)
	}

	expectedArgs := [][]int{ids}
	if diff := cmp.Diff(expectedArgs, idArgs); diff != "" {
		t.Errorf("unexpected arguments to statesFn (-want +got):\n%s", diff)
	}
}

func TestRemoveOrphanedBundleFilesMaxRequestBatchSize(t *testing.T) {
	bundleDir := testRoot(t)
	var ids []int
	for i := 1; i <= 225; i++ {
		ids = append(ids, i)
	}

	for _, id := range ids {
		path := filepath.Join(bundleDir, "dbs", fmt.Sprintf("%d.lsif.db", id))
		if err := makeFile(path, time.Now().Local()); err != nil {
			t.Fatalf("unexpected error creating file %s: %s", path, err)
		}
	}

	mockDB := dbmocks.NewMockDB()
	mockDB.GetStatesFunc.SetDefaultHook(func(ctx context.Context, ids []int) (map[int]string, error) {
		states := map[int]string{}
		for _, id := range ids {
			if id%2 == 0 {
				states[id] = "completed"
			}
		}
		return states, nil
	})

	j := &Janitor{
		db:        mockDB,
		bundleDir: bundleDir,
		metrics:   NewJanitorMetrics(metrics.TestRegisterer),
	}

	if err := j.removeOrphanedBundleFiles(); err != nil {
		t.Fatalf("unexpected error removing dead dumps: %s", err)
	}

	names, err := getFilenames(filepath.Join(bundleDir, "dbs"))
	if err != nil {
		t.Fatalf("unexpected error listing directory: %s", err)
	}

	if len(names) != 112 {
		t.Errorf("unexpected directory file count: want=%d have=%d", 112, len(names))
	}

	var allArgs []int
	for _, call := range mockDB.GetStatesFunc.History() {
		if len(call.Arg1) > OrphanedBundleBatchSize {
			t.Errorf("unexpected large slice: want < %d have=%d", OrphanedBundleBatchSize, len(call.Arg1))
		}

		allArgs = append(allArgs, call.Arg1...)
	}
	sort.Ints(allArgs)

	if diff := cmp.Diff(ids, allArgs); diff != "" {
		t.Errorf("unexpected flattened arguments to statesFn (-want +got):\n%s", diff)
	}
}

func TestBatchIntSlice(t *testing.T) {
	batches := batchIntSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9}, 2)
	expected := [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9}}

	if diff := cmp.Diff(expected, batches); diff != "" {
		t.Errorf("unexpected batch layout (-want +got):\n%s", diff)
	}
}
