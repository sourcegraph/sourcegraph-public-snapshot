package graphqlbackend

import (
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

func TestOffsetBasedCursorSlice(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	int2 := 2
	string1 := "1"
	string8 := "8"

	testCases := []struct {
		name string
		args *database.PaginationArgs
		want autogold.Value
	}{
		{
			"first page",
			&database.PaginationArgs{First: &int2},
			autogold.Expect([]int{1, 2}),
		},
		{
			"next page",
			&database.PaginationArgs{First: &int2, After: &string1},
			autogold.Expect([]int{3, 4}),
		},
		{
			"last page",
			&database.PaginationArgs{Last: &int2},
			autogold.Expect([]int{9, 10}),
		},
		{
			"previous page",
			&database.PaginationArgs{Last: &int2, Before: &string8},
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
