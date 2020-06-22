package janitor

import (
	"fmt"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store/mocks"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

func TestRemoveProcessedRecordsWithoutBundleFile(t *testing.T) {
	bundleDir := testRoot(t)
	ids := []int{1, 2, 3, 4, 5}

	for _, id := range []int{1, 3, 5} {
		path := filepath.Join(bundleDir, "dbs", fmt.Sprintf("%d", id), "sqlite.db")
		if err := makeFile(path, time.Now().Local()); err != nil {
			t.Fatalf("unexpected error creating file %s: %s", path, err)
		}
	}

	mockStore := storemocks.NewMockStore()
	mockStore.GetDumpIDsFunc.SetDefaultReturn(ids, nil)

	j := &Janitor{
		store:     mockStore,
		bundleDir: bundleDir,
		metrics:   NewJanitorMetrics(metrics.TestRegisterer),
	}

	if err := j.removeProcessedRecordsWithoutBundleFile(); err != nil {
		t.Fatalf("unexpected error removing processed uploads without bundle files: %s", err)
	}

	if len(mockStore.DeleteUploadByIDFunc.History()) != 2 {
		t.Errorf("unexpected number of DeleteUploadByID calls. want=%d have=%d", 2, len(mockStore.DeleteUploadByIDFunc.History()))
	} else {
		ids := []int{
			mockStore.DeleteUploadByIDFunc.History()[0].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[1].Arg1,
		}
		sort.Ints(ids)

		if diff := cmp.Diff([]int{2, 4}, ids); diff != "" {
			t.Errorf("unexpected dump ids (-want +got):\n%s", diff)
		}
	}
}

func TestRemoveOldUploadingRecords(t *testing.T) {
	bundleDir := testRoot(t)

	mockStore := storemocks.NewMockStore()
	mockStore.GetUploadsFunc.SetDefaultReturn([]store.Upload{
		{ID: 1},
		{ID: 2},
		{ID: 3},
		{ID: 4},
		{ID: 5},
	}, 0, nil)

	j := &Janitor{
		store:     mockStore,
		bundleDir: bundleDir,
		metrics:   NewJanitorMetrics(metrics.TestRegisterer),
	}

	if err := j.removeOldUploadingRecords(); err != nil {
		t.Fatalf("unexpected error removing processed uploads without bundle files: %s", err)
	}

	if len(mockStore.DeleteUploadByIDFunc.History()) != 5 {
		t.Errorf("unexpected number of DeleteUploadByID calls. want=%d have=%d", 5, len(mockStore.DeleteUploadByIDFunc.History()))
	} else {
		ids := []int{
			mockStore.DeleteUploadByIDFunc.History()[0].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[1].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[2].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[3].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[4].Arg1,
		}
		sort.Ints(ids)

		if diff := cmp.Diff([]int{1, 2, 3, 4, 5}, ids); diff != "" {
			t.Errorf("unexpected dump ids (-want +got):\n%s", diff)
		}
	}
}
