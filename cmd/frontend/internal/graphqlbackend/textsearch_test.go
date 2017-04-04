package graphqlbackend

import (
	"reflect"
	"testing"
)

func TestTruncateMatches(t *testing.T) {
	matches := []*fileMatch{
		{JLineMatches: []*lineMatch{
			{JOffsetAndLengths: [][]int32{{1, 2}, {3, 4}, {5, 6}}},
			{JOffsetAndLengths: nil},
			{JOffsetAndLengths: nil},
		}},
	}

	expected1 := []*fileMatch{
		{JLineMatches: []*lineMatch{
			{JOffsetAndLengths: [][]int32{{1, 2}, {3, 4}}},
		}},
	}
	results1, truncated := truncateMatches(matches, 2)
	if !truncated || !reflect.DeepEqual(results1, expected1) {
		t.Error("expected results to be truncated")
	}

	expected2 := []*fileMatch{
		{JLineMatches: []*lineMatch{
			{JOffsetAndLengths: [][]int32{{1, 2}, {3, 4}, {5, 6}}},
			{JOffsetAndLengths: nil},
			{JOffsetAndLengths: nil},
		}},
	}

	results2, truncated := truncateMatches(matches, 50)
	if truncated || !reflect.DeepEqual(results2, expected2) {
		t.Error("didn't expect reslts to be truncated")
	}

}
