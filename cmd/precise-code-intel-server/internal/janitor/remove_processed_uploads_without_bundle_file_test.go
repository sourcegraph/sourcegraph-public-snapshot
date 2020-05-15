package janitor

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	bundlemocks "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/mocks"
	dbmocks "github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

func TestRemoveProcessedUploadsWithoutBundleFile(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockDB.GetDumpIDsFunc.SetDefaultReturn([]int{1, 2, 3, 4, 5}, nil)
	mockBundleManagerClient.ExistsFunc.SetDefaultReturn(map[int]bool{
		1: true,
		2: false,
		3: true,
		4: false,
		5: true,
	}, nil)

	j := &Janitor{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
		metrics:             NewJanitorMetrics(metrics.TestRegisterer),
	}

	if err := j.removeProcessedUploadsWithoutBundleFile(); err != nil {
		t.Fatalf("unexpected error removing processed uploads without bundle files: %s", err)
	}

	if len(mockBundleManagerClient.ExistsFunc.History()) != 1 {
		t.Errorf("unexpected number of Exists calls. want=%d have=%d", 1, len(mockBundleManagerClient.ExistsFunc.History()))
	} else {
		call := mockBundleManagerClient.ExistsFunc.History()[0]

		if diff := cmp.Diff([]int{1, 2, 3, 4, 5}, call.Arg1); diff != "" {
			t.Errorf("unexpected dump ids (-want +got):\n%s", diff)
		}
	}

	if len(mockDB.DeleteUploadByIDFunc.History()) != 2 {
		t.Errorf("unexpected number of DeleteUploadByID calls. want=%d have=%d", 2, len(mockDB.DeleteUploadByIDFunc.History()))
	} else {
		ids := []int{
			mockDB.DeleteUploadByIDFunc.History()[0].Arg1,
			mockDB.DeleteUploadByIDFunc.History()[1].Arg1,
		}
		sort.Ints(ids)

		if diff := cmp.Diff([]int{2, 4}, ids); diff != "" {
			t.Errorf("unexpected dump ids (-want +got):\n%s", diff)
		}
	}
}

func TestRemoveProcessedUploadsWithoutBundleMaxRequestBatchSize(t *testing.T) {
	var ids []int
	for i := 1; i < 255; i++ {
		ids = append(ids, i)
	}

	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockDB.GetDumpIDsFunc.SetDefaultReturn(ids, nil)
	mockBundleManagerClient.ExistsFunc.SetDefaultHook(func(ctx context.Context, ids []int) (map[int]bool, error) {
		existsMap := map[int]bool{}
		for _, id := range ids {
			existsMap[id] = id%2 == 0
		}
		return existsMap, nil
	})

	j := &Janitor{
		db:                  mockDB,
		bundleManagerClient: mockBundleManagerClient,
		metrics:             NewJanitorMetrics(metrics.TestRegisterer),
	}

	if err := j.removeProcessedUploadsWithoutBundleFile(); err != nil {
		t.Fatalf("unexpected error removing processed uploads without bundle files: %s", err)
	}

	var idArgs [][]int
	for _, call := range mockBundleManagerClient.ExistsFunc.History() {
		idArgs = append(idArgs, call.Arg1)
	}

	var allArgs []int
	for _, args := range idArgs {
		if len(args) > BundleBatchSize {
			t.Errorf("unexpected large slice: want < %d have=%d", BundleBatchSize, len(args))
		}

		allArgs = append(allArgs, args...)
	}
	sort.Ints(allArgs)

	if diff := cmp.Diff(ids, allArgs); diff != "" {
		t.Errorf("unexpected flattened arguments to statesFn (-want +got):\n%s", diff)
	}

	if len(mockDB.DeleteUploadByIDFunc.History()) != 127 {
		t.Errorf("unexpected number of DeleteUploadByID calls. want=%d have=%d", 128, len(mockDB.DeleteUploadByIDFunc.History()))
	}
}

func TestBatchIntSlice(t *testing.T) {
	batches := batchIntSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9}, 2)
	expected := [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9}}

	if diff := cmp.Diff(expected, batches); diff != "" {
		t.Errorf("unexpected batch layout (-want +got):\n%s", diff)
	}
}
