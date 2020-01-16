package graphqlbackend

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/search/query"

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

func Test_didYouMeanQuotedResolver_Results(t *testing.T) {
	t.Run("regex error", func(t *testing.T) {
		raw := "*"
		_, err := query.ParseAndCheck(raw)
		if err == nil {
			t.Fatalf(`error returned from syntax.Parse("%s") is nil`, raw)
		}
		dymqr := didYouMeanQuotedResolver{query: raw, err: err}
		srr, err := dymqr.Results(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		alert := srr.alert
		if !strings.Contains(strings.ToLower(alert.title), "regexp") {
			t.Errorf("title is '%s', want it to contain 'regexp'", alert.title)
		}
		if !strings.Contains(alert.description, "regular expression") {
			t.Errorf("description is '%s', want it to contain 'regular expression'", alert.description)
		}
	})

	t.Run("type error that is not a regex error should show a suggestion", func(t *testing.T) {
		raw := "-foobar"
		_, err := query.ParseAndCheck(raw)
		if err == nil {
			t.Fatalf(`error returned from syntax.Parse("%s") is nil`, raw)
		}
		dymqr := didYouMeanQuotedResolver{query: raw, err: err}
		_, err = dymqr.Results(context.Background())
		if err == nil {
			t.Errorf("got nil error")
		}
	})

	t.Run("query parse error", func(t *testing.T) {
		raw := ":"
		_, err := query.ParseAndCheck(raw)
		dymqr := didYouMeanQuotedResolver{query: raw, err: err}
		srr, err := dymqr.Results(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		alert := srr.alert
		if strings.Contains(strings.ToLower(alert.title), "regexp") {
			t.Errorf("title is '%s', want it not to contain 'regexp'", alert.title)
		}
		if strings.Contains(alert.description, "regular expression") {
			t.Errorf("description is '%s', want it not to contain 'regular expression'", alert.description)
		}
	})

	t.Run("negated file field with an invalid regex", func(t *testing.T) {
		raw := "-f:(a"
		_, err := query.ParseAndCheck(raw)
		if err == nil {
			t.Fatal("ParseAndCheck failed to detect the invalid regex in the f: field")
		}
		dymqr := didYouMeanQuotedResolver{query: raw, err: err}
		srr, err := dymqr.Results(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		alert := srr.alert
		if len(alert.proposedQueries) != 1 {
			t.Fatalf("got %d proposed queries (%v), want exactly 1", len(alert.proposedQueries), alert.proposedQueries)
		}
	})
}

func Test_capFirst(t *testing.T) {
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
