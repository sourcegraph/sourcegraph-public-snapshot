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
					description: "query quoted entirely",
					query:       `""`,
				},
				{
					description: "query with parts quoted",
					query:       "",
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
					description: "query quoted entirely",
					query:       `"fmt.Sprintf(\""`,
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
