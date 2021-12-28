package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSearchPatternForSuggestion(t *testing.T) {
	cases := []struct {
		Name  string
		Alert searchAlert
		Want  string
	}{
		{
			Name: "with_regex_suggestion",
			Alert: searchAlert{
				title:       "An alert for regex",
				description: "An alert for regex",
				proposedQueries: []*searchQueryDescription{
					{
						description: "Some query description",
						query:       "repo:github.com/sourcegraph/sourcegraph",
						patternType: query.SearchTypeRegex,
					},
				},
			},
			Want: "repo:github.com/sourcegraph/sourcegraph patternType:regexp",
		},
		{
			Name: "with_structural_suggestion",
			Alert: searchAlert{
				title:       "An alert for structural",
				description: "An alert for structural",
				proposedQueries: []*searchQueryDescription{
					{
						description: "Some query description",
						query:       "repo:github.com/sourcegraph/sourcegraph",
						patternType: query.SearchTypeStructural,
					},
				},
			},
			Want: "repo:github.com/sourcegraph/sourcegraph patternType:structural",
		},
	}

	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			got := tt.Alert.ProposedQueries()
			if !reflect.DeepEqual((*got)[0].Query(), tt.Want) {
				t.Errorf("got: %s, want: %s", (*got)[0].query, tt.Want)
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

func TestAlertForDiffCommitSearchLimits(t *testing.T) {
	cases := []struct {
		name                 string
		multiErr             *multierror.Error
		wantAlertDescription string
	}{
		{
			name:                 "diff_search_warns_on_repos_greater_than_search_limit",
			multiErr:             multierror.Append(&multierror.Error{}, &commit.RepoLimitError{ResultType: "diff", Max: 50}),
			wantAlertDescription: `Diff search can currently only handle searching across 50 repositories at a time. Try using the "repo:" filter to narrow down which repositories to search, or using 'after:"1 week ago"'.`,
		},
		{
			name:                 "commit_search_warns_on_repos_greater_than_search_limit",
			multiErr:             multierror.Append(&multierror.Error{}, &commit.RepoLimitError{ResultType: "commit", Max: 50}),
			wantAlertDescription: `Commit search can currently only handle searching across 50 repositories at a time. Try using the "repo:" filter to narrow down which repositories to search, or using 'after:"1 week ago"'.`,
		},
		{
			name:                 "commit_search_warns_on_repos_greater_than_search_limit_with_time_filter",
			multiErr:             multierror.Append(&multierror.Error{}, &commit.TimeLimitError{ResultType: "commit", Max: 10000}),
			wantAlertDescription: `Commit search can currently only handle searching across 10000 repositories at a time. Try using the "repo:" filter to narrow down which repositories to search.`,
		},
	}

	for _, test := range cases {
		alert, _ := (&searchResolver{}).errorToAlert(context.Background(), test.multiErr)
		haveAlertDescription := alert.description
		if diff := cmp.Diff(test.wantAlertDescription, haveAlertDescription); diff != "" {
			t.Fatalf("test %s, mismatched alert (-want, +got):\n%s", test.name, diff)
		}
	}
}

func TestErrorToAlertStructuralSearch(t *testing.T) {
	cases := []struct {
		name           string
		errors         []error
		wantAlertTitle string
	}{
		{
			name:           "multierr_is_unaffected",
			errors:         []error{errors.New("some error")},
			wantAlertTitle: "",
		},
		{
			name: "surface_friendly_alert_on_oom_err_message",
			errors: []error{
				errors.New("some error"),
				errors.New("Worker_oomed"),
				errors.New("some other error"),
			},
			wantAlertTitle: "Structural search needs more memory",
		},
	}
	for _, test := range cases {
		multiErr := &multierror.Error{
			Errors:      test.errors,
			ErrorFormat: multierror.ListFormatFunc,
		}
		haveAlert, _ := (&searchResolver{}).errorToAlert(context.Background(), multiErr)

		if haveAlert != nil && haveAlert.title != test.wantAlertTitle {
			t.Fatalf("test %s, have alert: %q, want: %q", test.name, haveAlert.title, test.wantAlertTitle)
		}

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

func TestAlertForNoResolvedReposWithNonGlobalSearchContext(t *testing.T) {
	db := database.NewDB(nil)

	searchQuery := "context:@user repo:r1 foo"
	wantAlert := &searchAlert{
		prometheusType: "no_resolved_repos__context_none_in_common",
		title:          "No repositories found for your query within the context @user",
		proposedQueries: []*searchQueryDescription{{
			description: "search in the global context",
			query:       "context:global repo:r1 foo",
			patternType: query.SearchTypeRegex,
		}},
	}

	q, err := query.ParseLiteral(searchQuery)
	if err != nil {
		t.Fatal(err)
	}
	sr := searchResolver{
		db: database.NewDB(db),
		SearchInputs: &run.SearchInputs{
			OriginalQuery: searchQuery,
			Query:         q,
			UserSettings:  &schema.Settings{},
		},
	}

	alert := sr.alertForNoResolvedRepos(context.Background(), q)
	require.NoError(t, err)
	require.Equal(t, wantAlert, alert)
}
