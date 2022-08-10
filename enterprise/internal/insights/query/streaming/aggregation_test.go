package streaming

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestSortAggregate(t *testing.T) {
	testCases := []struct {
		name    string
		results map[string]int32
		max     int
		want    []*Aggregate
	}{
		{
			name: "returns sorted list",
			results: map[string]int32{
				"sg/sg":       5,
				"sg/handbook": 20,
				"sg/test":     2,
			},
			max: 5,
			want: []*Aggregate{
				{"sg/handbook", 20},
				{"sg/sg", 5},
				{"sg/test", 2},
			},
		},
		{
			name: "returns top 3 sorted list",
			results: map[string]int32{
				"sg/sg":       5,
				"sg/handbook": 20,
				"sg/test":     2,
				"sg/deploy":   12,
				"sg/haha":     4,
			},
			max: 3,
			want: []*Aggregate{
				{"sg/handbook", 20},
				{"sg/deploy", 12},
				{"sg/sg", 5},
			},
		},
		{
			name: "returns alphabetically sorted",
			results: map[string]int32{
				"sg/sg":       5,
				"sg/handbook": 20,
				"sg/test":     20,
			},
			max: 5,
			want: []*Aggregate{
				{"sg/handbook", 20},
				{"sg/test", 20},
				{"sg/sg", 5},
			},
		},
		{
			name: "returns top 3 alphabetically sorted",
			results: map[string]int32{
				"sg/sg":       5,
				"sg/handbook": 20,
				"sg/test":     12,
				"sg/deploy":   12,
				"sg/haha":     4,
			},
			max: 3,
			want: []*Aggregate{
				{"sg/handbook", 20},
				{"sg/deploy", 12},
				{"sg/test", 12},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := aggregated{}
			for val, count := range tc.results {
				a.Add(val, count)
			}
			autogold.Want(tc.name, tc.want).Equal(t, a.SortAggregate(tc.max))
		})
	}
}
