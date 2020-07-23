package graphqlbackend

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
		want []*searchQueryDescription
	}{
		{
			name: "empty",
			args: args{
				rawQuery: "",
			},
			want: []*searchQueryDescription{
				{
					description: wholeMsg,
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
					description: wholeMsg,
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
					description: partsMsg,
					query:       `r:hammer "[s]++"`,
				},
				{
					description: wholeMsg,
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

func TestCapFirst(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "a", in: "a", want: "A"},
		{name: "ab", in: "ab", want: "Ab"},
		{name: "хлеб", in: "хлеб", want: "Хлеб"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := capFirst(tt.in); got != tt.want {
				t.Errorf("makeTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}
