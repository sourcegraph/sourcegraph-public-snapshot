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

func TestComparePositionSCIP(t *testing.T) {
	testCases := []struct {
		line      int
		character int
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
		if cmp := comparePositionSCIP(scip.NewRange([]int32{5, 11, 13}), testCase.line, testCase.character); cmp != testCase.expected {
			t.Errorf("unexpected ComparePositionSCIP result for %d:%d. want=%d have=%d", testCase.line, testCase.character, testCase.expected, cmp)
		}
	}
}
