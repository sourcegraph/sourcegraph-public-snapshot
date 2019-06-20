package graphqlbackend

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func Test_proposedQuotedQueries(t *testing.T) {
	type args struct {
		rawQuery string
	}
	tests := []struct {
		name string
		args args
		want []*searchQueryDescription
	}{
		{
			name: "empty",
			args: args{
				rawQuery: "",
			},
			want: []*searchQueryDescription{
				{
					description: "quote the whole thing",
					query:       `""`,
				},
			},
		},
		{
			name: `fmt.Sprintf("`,
			args: args{
				rawQuery: `fmt.Sprintf("`,
			},
			want: []*searchQueryDescription{
				{
					description: "quote the whole thing",
					query:       `"fmt.Sprintf(\""`,
				},
			},
		},
		{
			name: `r:hammer [s]++"`,
			args: args{
				rawQuery: `r:hammer [s]++`,
			},
			want: []*searchQueryDescription{
				{
					description: "quote just the errored parts",
					query:       `r:hammer "[s]++"`,
				},
				{
					description: "quote the whole thing",
					query:       `"r:hammer [s]++"`,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proposedQuotedQueries(tt.args.rawQuery); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("proposedQuotedQueries() = \n%s\nwant\n%s", spew.Sdump(got), spew.Sdump(tt.want))
			}
		})
	}
}
