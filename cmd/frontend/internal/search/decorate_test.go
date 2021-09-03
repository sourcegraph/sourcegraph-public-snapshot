package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
	stream "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

func TestGroupLineMatches(t *testing.T) {
	data := []*result.LineMatch{
		{LineNumber: 1},
		{LineNumber: 2},
		{LineNumber: 3},
		{LineNumber: 5},
		{LineNumber: 6},
		{LineNumber: 8},
	}

	want := []group{
		{
			{LineNumber: 1},
			{LineNumber: 2},
			{LineNumber: 3},
		},
		{
			{LineNumber: 5},
			{LineNumber: 6},
		},
		{
			{LineNumber: 8},
		},
	}

	got := groupLineMatches(data)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("line match group partition wrong (-want +got):\n%s", diff)
	}
}

func TestToHunk(t *testing.T) {
	data := []group{
		{
			{LineNumber: 1, OffsetAndLengths: [][2]int32{{0, 1}}},
			{LineNumber: 2, OffsetAndLengths: [][2]int32{{2, 3}, {4, 5}}},
			{LineNumber: 3, OffsetAndLengths: [][2]int32{{6, 7}, {8, 9}}},
		},
		{
			{LineNumber: 5, OffsetAndLengths: [][2]int32{{0, 1}}},
			{LineNumber: 6, OffsetAndLengths: [][2]int32{{2, 3}, {4, 5}}},
		},
		{
			{LineNumber: 8, OffsetAndLengths: [][2]int32{{6, 7}, {8, 9}}},
		},
	}

	want := [][]stream.Range{
		{
			{
				Start: stream.Location{Offset: -1, Line: 1, Column: 0},
				End:   stream.Location{Offset: -1, Line: 1, Column: 1},
			},
			{
				Start: stream.Location{Offset: -1, Line: 2, Column: 2},
				End:   stream.Location{Offset: -1, Line: 2, Column: 5},
			},
			{
				Start: stream.Location{Offset: -1, Line: 2, Column: 4},
				End:   stream.Location{Offset: -1, Line: 2, Column: 9},
			},
			{
				Start: stream.Location{Offset: -1, Line: 3, Column: 6},
				End:   stream.Location{Offset: -1, Line: 3, Column: 13},
			},
			{
				Start: stream.Location{Offset: -1, Line: 3, Column: 8},
				End:   stream.Location{Offset: -1, Line: 3, Column: 17},
			},
		},
		{
			{
				Start: stream.Location{Offset: -1, Line: 5, Column: 0},
				End:   stream.Location{Offset: -1, Line: 5, Column: 1},
			},
			{
				Start: stream.Location{Offset: -1, Line: 6, Column: 2},
				End:   stream.Location{Offset: -1, Line: 6, Column: 5},
			},
			{
				Start: stream.Location{Offset: -1, Line: 6, Column: 4},
				End:   stream.Location{Offset: -1, Line: 6, Column: 9},
			},
		},
		{
			{
				Start: stream.Location{Offset: -1, Line: 8, Column: 6},
				End:   stream.Location{Offset: -1, Line: 8, Column: 13},
			},
			{
				Start: stream.Location{Offset: -1, Line: 8, Column: 8},
				End:   stream.Location{Offset: -1, Line: 8, Column: 17},
			},
		},
	}

	var got [][]stream.Range
	for _, group := range data {
		got = append(got, toMatchRanges(group))
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("line match group partition wrong (-want +got):\n%s", diff)
	}
}
