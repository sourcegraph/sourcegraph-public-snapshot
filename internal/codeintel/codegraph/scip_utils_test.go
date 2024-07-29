package codegraph

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/scip/bindings/go/scip"
)

type testOccurrence struct {
	Ints []int32
}

func (o testOccurrence) GetRange() []int32 {
	return o.Ints
}

type args[Occurrence IOccurrence] struct {
	occurrences []Occurrence
	search      scip.Range
}
type testCase[Occurrence IOccurrence] struct {
	name string
	args args[Occurrence]
	want []Occurrence
}

func Test_findOccurrencesWithEqualRange(t *testing.T) {
	tests := []testCase[testOccurrence]{
		{
			name: "empty",
			args: args[testOccurrence]{
				occurrences: []testOccurrence{},
				search:      scip.NewRangeUnchecked([]int32{1, 1, 4}),
			},
			want: []testOccurrence{},
		},
		{
			name: "exact match",
			args: args[testOccurrence]{
				occurrences: []testOccurrence{
					{Ints: []int32{1, 0, 3}},
					{Ints: []int32{1, 3, 5}},
					{Ints: []int32{1, 5, 6}},
				},
				search: scip.NewRangeUnchecked([]int32{1, 3, 5}),
			},
			want: []testOccurrence{
				{Ints: []int32{1, 3, 5}},
			},
		},
		{
			name: "no match",
			args: args[testOccurrence]{
				occurrences: []testOccurrence{
					{Ints: []int32{1, 0, 3}},
					{Ints: []int32{1, 3, 5}},
					{Ints: []int32{1, 5, 7}},
				},
				search: scip.NewRangeUnchecked([]int32{1, 4, 6}),
			},
			want: []testOccurrence{},
		},
		{
			name: "multi match",
			args: args[testOccurrence]{
				occurrences: []testOccurrence{
					{Ints: []int32{1, 0, 3}},
					{Ints: []int32{1, 3, 7}},
					{Ints: []int32{1, 3, 7}},
					{Ints: []int32{1, 3, 7}},
					{Ints: []int32{1, 7, 10}},
				},
				search: scip.NewRangeUnchecked([]int32{1, 3, 7}),
			},
			want: []testOccurrence{
				{Ints: []int32{1, 3, 7}},
				{Ints: []int32{1, 3, 7}},
				{Ints: []int32{1, 3, 7}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindOccurrencesWithEqualRange(tt.args.occurrences, tt.args.search)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("unexpected ranges (-want +got):\n%s", diff)
			}
		})
	}
}
