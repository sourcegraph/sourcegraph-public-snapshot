package query

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestProposedQuotedQueries(t *testing.T) {
	type args struct {
		rawQuery string
	}
	tests := []struct {
		name string
		args args
		want []*ProposedQuery
	}{
		{
			name: "empty",
			args: args{
				rawQuery: "",
			},
			want: []*ProposedQuery{
				{
					Description: wholeMsg,
					Query:       `""`,
				},
			},
		},
		{
			name: `fmt.Sprintf("`,
			args: args{
				rawQuery: `fmt.Sprintf("`,
			},
			want: []*ProposedQuery{
				{
					Description: wholeMsg,
					Query:       `"fmt.Sprintf(\""`,
				},
			},
		},
		{
			name: `r:hammer [s]++"`,
			args: args{
				rawQuery: `r:hammer [s]++`,
			},
			want: []*ProposedQuery{
				{
					Description: partsMsg,
					Query:       `r:hammer "[s]++"`,
				},
				{
					Description: wholeMsg,
					Query:       `"r:hammer [s]++"`,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProposedQuotedQueries(tt.args.rawQuery); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("proposedQuotedQueries() = \n%s\nwant\n%s", spew.Sdump(got), spew.Sdump(tt.want))
			}
		})
	}
}
