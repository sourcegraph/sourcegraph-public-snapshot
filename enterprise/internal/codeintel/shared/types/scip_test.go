package types

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/scip/bindings/go/scip"
)

func TestFindOccurrences(t *testing.T) {
	occurrences := []*scip.Occurrence{
		{Range: []int32{0, 3, 4, 5}},
		{Range: []int32{1, 3, 3, 5}},
		{Range: []int32{2, 3, 5}},
		{Range: []int32{5, 3, 5}},
		{Range: []int32{6, 3, 5}},
	}

	var matchingRanges [][]int32
	for _, occurrence := range FindOccurrences(occurrences, 2, 4) {
		matchingRanges = append(matchingRanges, occurrence.Range)
	}

	expected := [][]int32{
		occurrences[2].Range,
		occurrences[1].Range,
		occurrences[0].Range,
	}
	if diff := cmp.Diff(expected, matchingRanges); diff != "" {
		t.Errorf("unexpected FindOccurrences result (-want +got):\n%s", diff)
	}
}

func TestSortOccurrences(t *testing.T) {
	occurrences := []*scip.Occurrence{
		{Range: []int32{2, 3, 5}},       // rank 2
		{Range: []int32{11, 10, 12}},    // rank 10
		{Range: []int32{6, 3, 5}},       // rank 4
		{Range: []int32{10, 4, 8}},      // rank 6
		{Range: []int32{10, 10, 12}},    // rank 7
		{Range: []int32{0, 3, 4, 5}},    // rank 0
		{Range: []int32{12, 1, 13, 12}}, // rank 11
		{Range: []int32{11, 1, 3}},      // rank 8
		{Range: []int32{5, 3, 5}},       // rank 3
		{Range: []int32{10, 1, 3}},      // rank 5
		{Range: []int32{12, 10, 13, 3}}, // rank 13
		{Range: []int32{11, 4, 8}},      // rank 9
		{Range: []int32{12, 4, 13, 8}},  // rank 12
		{Range: []int32{1, 3, 3, 5}},    // rank 1
	}
	unsorted := make([]*scip.Occurrence, len(occurrences))
	copy(unsorted, occurrences)

	ranges := [][]int32{}
	for _, r := range SortOccurrences(unsorted) {
		ranges = append(ranges, r.Range)
	}

	expected := [][]int32{
		occurrences[5].Range,
		occurrences[13].Range,
		occurrences[0].Range,
		occurrences[8].Range,
		occurrences[2].Range,
		occurrences[9].Range,
		occurrences[3].Range,
		occurrences[4].Range,
		occurrences[7].Range,
		occurrences[11].Range,
		occurrences[1].Range,
		occurrences[6].Range,
		occurrences[12].Range,
		occurrences[10].Range,
	}
	if diff := cmp.Diff(expected, ranges); diff != "" {
		t.Errorf("unexpected occurrence order (-want +got):\n%s", diff)
	}
}

func TestSortRanges(t *testing.T) {
	occurrences := []*scip.Range{
		scip.NewRange([]int32{2, 3, 5}),       // rank 2
		scip.NewRange([]int32{11, 10, 12}),    // rank 10
		scip.NewRange([]int32{6, 3, 5}),       // rank 4
		scip.NewRange([]int32{10, 4, 8}),      // rank 6
		scip.NewRange([]int32{10, 10, 12}),    // rank 7
		scip.NewRange([]int32{0, 3, 4, 5}),    // rank 0
		scip.NewRange([]int32{12, 1, 13, 12}), // rank 11
		scip.NewRange([]int32{11, 1, 3}),      // rank 8
		scip.NewRange([]int32{5, 3, 5}),       // rank 3
		scip.NewRange([]int32{10, 1, 3}),      // rank 5
		scip.NewRange([]int32{12, 10, 13, 3}), // rank 13
		scip.NewRange([]int32{11, 4, 8}),      // rank 9
		scip.NewRange([]int32{12, 4, 13, 8}),  // rank 12
		scip.NewRange([]int32{1, 3, 3, 5}),    // rank 1
	}
	unsorted := make([]*scip.Range, len(occurrences))
	copy(unsorted, occurrences)

	ranges := [][]int32{}
	for _, r := range SortRanges(unsorted) {
		ranges = append(ranges, r.SCIPRange())
	}

	// TODO - better data
	expected := [][]int32{
		occurrences[5].SCIPRange(),
		occurrences[13].SCIPRange(),
		occurrences[0].SCIPRange(),
		occurrences[8].SCIPRange(),
		occurrences[2].SCIPRange(),
		occurrences[9].SCIPRange(),
		occurrences[3].SCIPRange(),
		occurrences[4].SCIPRange(),
		occurrences[7].SCIPRange(),
		occurrences[11].SCIPRange(),
		occurrences[1].SCIPRange(),
		occurrences[6].SCIPRange(),
		occurrences[12].SCIPRange(),
		occurrences[10].SCIPRange(),
	}
	if diff := cmp.Diff(expected, ranges); diff != "" {
		t.Errorf("unexpected occurrence order (-want +got):\n%s", diff)
	}
}

func TestComparePositionToRange(t *testing.T) {
	testCases := []struct {
		line      int32
		character int32
		expected  int
	}{
		{5, 11, 0},
		{5, 12, 0},
		{5, 13, -1},
		{4, 12, +1},
		{5, 10, +1},
		{5, 14, -1},
		{6, 12, -1},
	}

	for _, testCase := range testCases {
		if cmpRes := comparePositionToRange(5, 11, 5, 13, testCase.line, testCase.character); cmpRes != testCase.expected {
			t.Errorf("unexpected ComparePositionSCIP result for %d:%d. want=%d have=%d", testCase.line, testCase.character, testCase.expected, cmpRes)
		}
	}
}
