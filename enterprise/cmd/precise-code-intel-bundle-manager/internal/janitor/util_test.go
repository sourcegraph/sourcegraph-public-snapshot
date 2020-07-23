package janitor

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBatchIntSlice(t *testing.T) {
	batches := batchIntSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9}, 2)
	expected := [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9}}

	if diff := cmp.Diff(expected, batches); diff != "" {
		t.Errorf("unexpected batch layout (-want +got):\n%s", diff)
	}
}
