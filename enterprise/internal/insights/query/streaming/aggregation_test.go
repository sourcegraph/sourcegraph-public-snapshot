package streaming

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestAggregateAdd(t *testing.T) {
	testCases := []struct {
		name  string
		have  aggregated
		value string
		count int32
		want  aggregated
	}{
		{
			name: "adds up overflow",
			have: aggregated{
				maxResults: 1,
				Results:    map[string]int32{"A": 12},
			},
			value: "B",
			count: 22,
			want: aggregated{
				maxResults: 1,
				Results:    map[string]int32{"A": 12},
				Overflow:   22,
			},
		},
		{
			name: "adds new result",
			have: aggregated{
				maxResults: 2,
				Results:    map[string]int32{"A": 24},
			},
			value: "B",
			count: 32,
			want: aggregated{
				maxResults: 2,
				Results:    map[string]int32{"A": 24, "B": 32},
			},
		},
		{
			name: "updates existing results",
			have: aggregated{
				maxResults: 1,
				Results:    map[string]int32{"C": 5},
			},
			value: "C",
			count: 11,
			want: aggregated{
				maxResults: 1,
				Results:    map[string]int32{"C": 16},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.have.Add(tc.value, tc.count)
			autogold.Want(tc.name, tc.want).Equal(t, tc.have)
		})
	}
}

func TestSortAggregate(t *testing.T) {
	a := aggregated{
		Results:    make(map[string]int32),
		maxResults: 5,
	}

	// Add 5 distinct elements. Update 1 existing.
	a.Add("sg/1", 5)
	a.Add("sg/2", 10)
	a.Add("sg/3", 8)
	a.Add("sg/1", 3)
	a.Add("sg/4", 22)
	a.Add("sg/5", 60)

	// Add two more elements.
	a.Add("sg/too-much", 12)
	a.Add("sg/lost", 1)

	// Update another one.
	a.Add("sg/2", 5)

	autogold.Want("overflow should be 13", int32(13)).Equal(t, a.Overflow)

	want := []*Aggregate{
		{"sg/5", 60},
		{"sg/4", 22},
		{"sg/2", 15},
		{"sg/1", 8},
		{"sg/3", 8},
	}
	autogold.Want("SortAggregate should return DESC sorted list", want).Equal(t, a.SortAggregate())
}
