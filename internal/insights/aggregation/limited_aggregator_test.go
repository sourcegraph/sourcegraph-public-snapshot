package aggregation

import (
	"testing"

	"github.com/hexops/autogold/v2"
)

func TestAddAggregate(t *testing.T) {
	testCases := []struct {
		name  string
		have  limitedAggregator
		value string
		count int32
		want  limitedAggregator
	}{
		{
			name: "invalid buffer size does nothing",
			have: limitedAggregator{
				resultBufferSize: -1,
			},
			value: "B",
			count: 9,
			want: limitedAggregator{
				resultBufferSize: -1,
			},
		},
		{
			name: "adds up other count",
			have: limitedAggregator{
				resultBufferSize: 1,
				Results:          map[string]int32{"A": 12},
				smallestResult:   &Aggregate{"A", 12},
			},
			value: "B",
			count: 9,
			want: limitedAggregator{
				resultBufferSize: 1,
				Results:          map[string]int32{"A": 12},
				smallestResult:   &Aggregate{"A", 12},
				OtherCount:       OtherCount{ResultCount: 9, GroupCount: 1},
			},
		},
		{
			name: "adds new result",
			have: limitedAggregator{
				resultBufferSize: 2,
				Results:          map[string]int32{"A": 24},
				smallestResult:   &Aggregate{"A", 24},
			},
			value: "B",
			count: 32,
			want: limitedAggregator{
				resultBufferSize: 2,
				Results:          map[string]int32{"A": 24, "B": 32},
				smallestResult:   &Aggregate{"A", 24},
			},
		},
		{
			name: "updates existing results",
			have: limitedAggregator{
				resultBufferSize: 1,
				Results:          map[string]int32{"C": 5},
				smallestResult:   &Aggregate{"C", 5},
			},
			value: "C",
			count: 11,
			want: limitedAggregator{
				resultBufferSize: 1,
				Results:          map[string]int32{"C": 16},
				smallestResult:   &Aggregate{"C", 16},
			},
		},
		{
			name: "ejects smallest result",
			have: limitedAggregator{
				resultBufferSize: 1,
				Results:          map[string]int32{"C": 5},
				smallestResult:   &Aggregate{"C", 5},
			},
			value: "A",
			count: 15,
			want: limitedAggregator{
				resultBufferSize: 1,
				Results:          map[string]int32{"A": 15},
				smallestResult:   &Aggregate{"A", 15},
				OtherCount:       OtherCount{ResultCount: 5, GroupCount: 1},
			},
		},
		{
			name: "adds up other group count",
			have: limitedAggregator{
				resultBufferSize: 1,
				Results:          map[string]int32{"A": 12},
				smallestResult:   &Aggregate{"A", 12},
				OtherCount:       OtherCount{ResultCount: 9, GroupCount: 1},
			},
			value: "B",
			count: 9,
			want: limitedAggregator{
				resultBufferSize: 1,
				Results:          map[string]int32{"A": 12},
				smallestResult:   &Aggregate{"A", 12},
				OtherCount:       OtherCount{ResultCount: 18, GroupCount: 2},
			},
		},
		{
			name: "first result becomes smallest result",
			have: limitedAggregator{
				resultBufferSize: 1,
				Results:          map[string]int32{},
			},
			value: "new",
			count: 1,
			want: limitedAggregator{
				resultBufferSize: 1,
				Results:          map[string]int32{"new": 1},
				smallestResult:   &Aggregate{"new", 1},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.have.Add(tc.value, tc.count)
			autogold.Expect(tc.want).Equal(t, tc.have)
		})
	}
}

func TestFindSmallestAggregate(t *testing.T) {
	testCases := []struct {
		name string
		have limitedAggregator
		want *Aggregate
	}{
		{
			name: "returns nil for empty results",
			want: nil,
		},
		{
			name: "one result is smallest",
			have: limitedAggregator{
				Results: map[string]int32{"myresult": 20},
			},
			want: &Aggregate{"myresult", 20},
		},
		{
			name: "finds smallest result by count",
			have: limitedAggregator{
				Results: map[string]int32{"high": 20, "low": 5, "mid": 10},
			},
			want: &Aggregate{"low", 5},
		},
		{
			name: "finds smallest result by label",
			have: limitedAggregator{
				Results: map[string]int32{"outsider": 5, "abc/1": 5, "abcd": 5, "abc/2": 5},
			},
			want: &Aggregate{"abc/1", 5},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.have.findSmallestAggregate()
			autogold.Expect(tc.want).Equal(t, got)
		})
	}
}

func TestSortAggregate(t *testing.T) {
	a := limitedAggregator{
		Results:          make(map[string]int32),
		resultBufferSize: 5,
	}

	// Add 5 distinct elements. Update 1 existing.
	a.Add("sg/1", 5)
	a.Add("sg/2", 10)
	a.Add("sg/3", 8)
	a.Add("sg/1", 3)
	a.Add("sg/4", 22)
	a.Add("sg/5", 60)

	// Add two more elements.
	a.Add("sg/will-eject", 12)
	a.Add("sg/lost", 1)

	// Update another one.
	a.Add("sg/2", 5)

	// Update the smallest result, and then not.
	a.Add("sg/3", 1)
	a.Add("sg/will-eject", 1)

	autogold.Expect(int32(9)).Equal(t, a.OtherCount.ResultCount)
	autogold.Expect(int32(2)).Equal(t, a.OtherCount.GroupCount)

	want := []*Aggregate{
		{"sg/5", 60},
		{"sg/4", 22},
		{"sg/2", 15},
		{"sg/will-eject", 13},
		{"sg/3", 9},
	}
	autogold.Expect(want).Equal(t, a.SortAggregate())
}
