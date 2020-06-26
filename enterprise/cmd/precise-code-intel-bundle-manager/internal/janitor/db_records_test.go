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

func TestRemoveCompletedRecordsWithoutBundleFile(t *testing.T) {
	bundleDir := testRoot(t)

	for _, id := range []int{1, 3, 5, 7, 9} {
		path := filepath.Join(bundleDir, "dbs", fmt.Sprintf("%d", id), "sqlite.db")
		if err := makeFile(path, time.Now().Local()); err != nil {
			t.Fatalf("unexpected error creating file %s: %s", path, err)
		}
	}

	mockStore := storemocks.NewMockStore()
	mockStore.GetUploadsFunc.PushReturn([]store.Upload{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}}, 10, nil)
	mockStore.GetUploadsFunc.PushReturn([]store.Upload{{ID: 6}, {ID: 7}, {ID: 8}, {ID: 9}, {ID: 10}}, 10, nil)

	j := &Janitor{
		store:     mockStore,
		bundleDir: bundleDir,
		metrics:   NewJanitorMetrics(metrics.TestRegisterer),
	}

	if err := j.removeCompletedRecordsWithoutBundleFile(); err != nil {
		t.Fatalf("unexpected error removing completed uploads without bundle files: %s", err)
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

		if diff := cmp.Diff([]int{2, 4, 6, 8, 10}, ids); diff != "" {
			t.Errorf("unexpected dump ids (-want +got):\n%s", diff)
		}
	}
}

func TestRemoveOldUploadingRecords(t *testing.T) {
	bundleDir := testRoot(t)

	mockStore := storemocks.NewMockStore()
	mockStore.GetUploadsFunc.PushReturn([]store.Upload{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}}, 10, nil)
	mockStore.GetUploadsFunc.PushReturn([]store.Upload{{ID: 6}, {ID: 7}, {ID: 8}, {ID: 9}, {ID: 10}}, 10, nil)

	j := &Janitor{
		store:     mockStore,
		bundleDir: bundleDir,
		metrics:   NewJanitorMetrics(metrics.TestRegisterer),
	}

	if err := j.removeOldUploadingRecords(); err != nil {
		t.Fatalf("unexpected error removing old records that have not finished uploading: %s", err)
	}

	if len(mockStore.DeleteUploadByIDFunc.History()) != 10 {
		t.Errorf("unexpected number of DeleteUploadByID calls. want=%d have=%d", 10, len(mockStore.DeleteUploadByIDFunc.History()))
	} else {
		ids := []int{
			mockStore.DeleteUploadByIDFunc.History()[0].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[1].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[2].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[3].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[4].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[5].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[6].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[7].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[8].Arg1,
			mockStore.DeleteUploadByIDFunc.History()[9].Arg1,
		}
		sort.Ints(ids)

		if diff := cmp.Diff([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, ids); diff != "" {
			t.Errorf("unexpected dump ids (-want +got):\n%s", diff)
		}
	}
}
