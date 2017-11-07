package graphqlbackend

import (
	"fmt"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/search2"
)

func TestAddQueryRegexpField(t *testing.T) {
	tests := []struct {
		query      string
		addField   search2.Field
		addPattern string
		want       string
	}{
		{
			query:      "",
			addField:   "repo",
			addPattern: "p",
			want:       "repo:p",
		},
		{
			query:      "foo",
			addField:   "repo",
			addPattern: "p",
			want:       "foo repo:p",
		},
		{
			query:      "foo repo:q",
			addField:   "repo",
			addPattern: "p",
			want:       "foo repo:q repo:p",
		},
		{
			query:      "foo repo:p",
			addField:   "repo",
			addPattern: "p",
			want:       "foo repo:p",
		},
		{
			query:      "foo repo:p",
			addField:   "repo",
			addPattern: "pp",
			want:       "foo repo:pp",
		},
		{
			query:      "foo repo:p",
			addField:   "repo",
			addPattern: "^p",
			want:       "foo repo:^p",
		},
		{
			query:      "foo repo:p",
			addField:   "repo",
			addPattern: "p$",
			want:       "foo repo:p$",
		},
		{
			query:      "foo repo:^p",
			addField:   "repo",
			addPattern: "^pq",
			want:       "foo repo:^pq",
		},
		{
			query:      "foo repo:p$",
			addField:   "repo",
			addPattern: "qp$",
			want:       "foo repo:qp$",
		},
		{
			query:      "foo repo:^p",
			addField:   "repo",
			addPattern: "x$",
			want:       "foo repo:^p repo:x$",
		},
		{
			query:      "foo repo:p|q",
			addField:   "repo",
			addPattern: "pq",
			want:       "foo repo:p|q repo:pq",
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s, add %s:%s", test.query, test.addField, test.addPattern), func(t *testing.T) {
			query, err := resolveQuery(test.query)
			if err != nil {
				t.Fatal(err)
			}
			got := addQueryRegexpField(query.tokens, test.addField, test.addPattern)
			if got.String() != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}
