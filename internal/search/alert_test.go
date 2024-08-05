package search

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func TestMaxPriorityAlert(t *testing.T) {
	t.Run("no alerts", func(t *testing.T) {
		require.Equal(t, (*Alert)(nil), MaxPriorityAlert())
	})

	t.Run("nil alert", func(t *testing.T) {
		require.Equal(t, (*Alert)(nil), MaxPriorityAlert(nil))
	})

	t.Run("one alert", func(t *testing.T) {
		a1 := Alert{Title: "test1"}
		require.Equal(t, &a1, MaxPriorityAlert(&a1))
	})

	t.Run("equal priority alerts", func(t *testing.T) {
		a1 := Alert{Title: "test1"}
		a2 := Alert{Title: "test2"}
		require.Equal(t, &a1, MaxPriorityAlert(&a1, &a2))
	})

	t.Run("higher priority alerts", func(t *testing.T) {
		a1 := Alert{Title: "test1"}
		a2 := Alert{Title: "test2", Priority: 2}
		require.Equal(t, &a2, MaxPriorityAlert(&a1, &a2))
	})

	t.Run("nil and non-nil", func(t *testing.T) {
		a1 := Alert{Title: "test1"}
		require.Equal(t, &a1, MaxPriorityAlert(nil, &a1))
	})

	t.Run("non-nil and nil", func(t *testing.T) {
		a1 := Alert{Title: "test1"}
		require.Equal(t, &a1, MaxPriorityAlert(&a1, nil))
	})
}

func TestSearchPatternForSuggestion(t *testing.T) {
	cases := []struct {
		Name  string
		Alert *Alert
		Want  string
	}{
		{
			Name: "with_regex_suggestion",
			Alert: &Alert{
				Title:       "An alert for regex",
				Description: "An alert for regex",
				ProposedQueries: []*QueryDescription{
					{
						Description: "Some query description",
						Query:       "repo:github.com/sourcegraph/sourcegraph",
						PatternType: query.SearchTypeRegex,
					},
				},
			},
			Want: "repo:github.com/sourcegraph/sourcegraph patternType:regexp",
		},
		{
			Name: "with_structural_suggestion",
			Alert: &Alert{
				Title:       "An alert for structural",
				Description: "An alert for structural",
				ProposedQueries: []*QueryDescription{
					{
						Description: "Some query description",
						Query:       "repo:github.com/sourcegraph/sourcegraph",
						PatternType: query.SearchTypeStructural,
					},
				},
			},
			Want: "repo:github.com/sourcegraph/sourcegraph patternType:structural",
		},
	}

	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			got := tt.Alert.ProposedQueries
			if !reflect.DeepEqual(got[0].QueryString(), tt.Want) {
				t.Errorf("got: %s, want: %s", got[0].QueryString(), tt.Want)
			}
		})
	}
}

func TestAddQueryRegexpField(t *testing.T) {
	tests := []struct {
		query      string
		addField   string
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
			want:       "repo:p foo",
		},
		{
			query:      "foo repo:p",
			addField:   "repo",
			addPattern: "p",
			want:       "repo:p foo",
		},
		{
			query:      "foo repo:q",
			addField:   "repo",
			addPattern: "p",
			want:       "repo:q repo:p foo",
		},
		{
			query:      "foo repo:p",
			addField:   "repo",
			addPattern: "pp",
			want:       "repo:pp foo",
		},
		{
			query:      "foo repo:p",
			addField:   "repo",
			addPattern: "^p",
			want:       "repo:^p foo",
		},
		{
			query:      "foo repo:p",
			addField:   "repo",
			addPattern: "p$",
			want:       "repo:p$ foo",
		},
		{
			query:      "foo repo:^p",
			addField:   "repo",
			addPattern: "^pq",
			want:       "repo:^pq foo",
		},
		{
			query:      "foo repo:p$",
			addField:   "repo",
			addPattern: "qp$",
			want:       "repo:qp$ foo",
		},
		{
			query:      "foo repo:^p",
			addField:   "repo",
			addPattern: "x$",
			want:       "repo:^p repo:x$ foo",
		},
		{
			query:      "foo repo:p|q",
			addField:   "repo",
			addPattern: "pq",
			want:       "repo:p|q repo:pq foo",
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s, add %s:%s", test.query, test.addField, test.addPattern), func(t *testing.T) {
			q, err := query.ParseLiteral(test.query)
			if err != nil {
				t.Fatal(err)
			}
			got := query.AddRegexpField(q, test.addField, test.addPattern)
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
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

func TestQuoteSuggestions(t *testing.T) {
	t.Run("regex error", func(t *testing.T) {
		raw := "*"
		_, err := query.Pipeline(query.InitRegexp(raw))
		if err == nil {
			t.Fatalf("error returned from query.ParseRegexp(%q) is nil", raw)
		}
		alert := AlertForQuery(err)
		if !strings.Contains(alert.Description, "regexp") {
			t.Errorf("description is '%s', want it to contain 'regexp'", alert.Description)
		}
	})
}
