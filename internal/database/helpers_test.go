package database

import (
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestOffsetBasedCursorSlice(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	testCases := []struct {
		name string
		args *PaginationArgs
		want autogold.Value
	}{
		{
			"first page",
			&PaginationArgs{First: pointers.Ptr(2)},
			autogold.Expect([]int{1, 2}),
		},
		{
			"next page",
			&PaginationArgs{First: pointers.Ptr(2), After: []any{1}},
			autogold.Expect([]int{3, 4}),
		},
		{
			"last page",
			&PaginationArgs{Last: pointers.Ptr(2)},
			autogold.Expect([]int{9, 10}),
		},
		{
			"previous page",
			&PaginationArgs{Last: pointers.Ptr(2), Before: []any{8}},
			autogold.Expect([]int{7, 8}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := OffsetBasedCursorSlice(slice, tc.args)
			if err != nil {
				t.Fatal(err)
			}
			tc.want.Equal(t, result)
		})
	}
}
