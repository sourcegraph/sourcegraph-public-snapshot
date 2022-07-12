package querybuilder

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func TestWithDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     string
		defaults query.Parameters
	}{
		{
			name:     "no defaults",
			input:    "repo:myrepo testquery",
			want:     "repo:myrepo /testquery/",
			defaults: []query.Parameter{},
		},
		{
			name:     "no defaults with fork archived",
			input:    "repo:myrepo testquery fork:no archived:no",
			want:     "repo:myrepo fork:no archived:no /testquery/",
			defaults: []query.Parameter{},
		},
		{
			name:  "default archived",
			input: "repo:myrepo testquery fork:no",
			want:  "archived:yes repo:myrepo fork:no /testquery/",
			defaults: []query.Parameter{{
				Field:      query.FieldArchived,
				Value:      string(query.Yes),
				Negated:    false,
				Annotation: query.Annotation{},
			}},
		},
		{
			name:  "default fork and archived",
			input: "repo:myrepo testquery",
			want:  "archived:no fork:no repo:myrepo /testquery/",
			defaults: []query.Parameter{{
				Field:      query.FieldArchived,
				Value:      string(query.No),
				Negated:    false,
				Annotation: query.Annotation{},
			}, {
				Field:      query.FieldFork,
				Value:      string(query.No),
				Negated:    false,
				Annotation: query.Annotation{},
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := withDefaults(BasicQuery(test.input), test.defaults)
			if err != nil {
				t.Fatal(err)
			}
			println(got)
			if diff := cmp.Diff(test.want, string(got)); diff != "" {
				t.Fatalf("%s failed (want/got): %s", test.name, diff)
			}
		})
	}
}

func TestMultiRepoQuery(t *testing.T) {
	tests := []struct {
		name     string
		repos    []string
		want     string
		defaults query.Parameters
	}{
		{
			name:     "single repo",
			repos:    []string{"repo1"},
			want:     `count:99999999 /testquery/ repo:^(repo1)$`,
			defaults: []query.Parameter{},
		},
		{
			name:  "multiple repo",
			repos: []string{"repo1", "repo2"},
			want:  `archived:no fork:no count:99999999 /testquery/ repo:^(repo1|repo2)$`,
			defaults: []query.Parameter{{
				Field:      query.FieldArchived,
				Value:      string(query.No),
				Negated:    false,
				Annotation: query.Annotation{},
			}, {
				Field:      query.FieldFork,
				Value:      string(query.No),
				Negated:    false,
				Annotation: query.Annotation{},
			}},
		},
		{
			name:  "multiple repo",
			repos: []string{"github.com/myrepos/repo1", "github.com/myrepos/repo2"},
			want:  `archived:no fork:no count:99999999 /testquery/ repo:^(github\.com/myrepos/repo1|github\.com/myrepos/repo2)$`,
			defaults: []query.Parameter{{
				Field:      query.FieldArchived,
				Value:      string(query.No),
				Negated:    false,
				Annotation: query.Annotation{},
			}, {
				Field:      query.FieldFork,
				Value:      string(query.No),
				Negated:    false,
				Annotation: query.Annotation{},
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MultiRepoQuery("testquery", test.repos, test.defaults)
			if err != nil {
				t.Fatal(err)
			}
			println(got)
			if diff := cmp.Diff(test.want, string(got)); diff != "" {
				t.Fatalf("%s failed (want/got): %s", test.name, diff)
			}
		})
	}
}

func TestDefaults(t *testing.T) {
	tests := []struct {
		name  string
		input bool
		want  query.Parameters
	}{
		{
			name:  "all repos",
			input: true,
			want: query.Parameters{{
				Field:      query.FieldFork,
				Value:      string(query.No),
				Negated:    false,
				Annotation: query.Annotation{},
			}, {
				Field:      query.FieldArchived,
				Value:      string(query.No),
				Negated:    false,
				Annotation: query.Annotation{},
			}},
		},
		{
			name:  "some repos",
			input: false,
			want: query.Parameters{{
				Field:      query.FieldFork,
				Value:      string(query.Yes),
				Negated:    false,
				Annotation: query.Annotation{},
			}, {
				Field:      query.FieldArchived,
				Value:      string(query.Yes),
				Negated:    false,
				Annotation: query.Annotation{},
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CodeInsightsQueryDefaults(test.input)

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Fatalf("%s failed (want/got): %s", test.name, diff)
			}
		})
	}
}

func TestComputeInsightCommandQuery(t *testing.T) {
	tests := []struct {
		name       string
		inputQuery string
		mapType    MapType
		want       string
	}{
		{
			name:       "verify archive fork map to lang",
			inputQuery: "repo:abc123@12346f fork:yes archived:yes findme",
			mapType:    Lang,
			want:       "repo:abc123@12346f fork:yes archived:yes content:output.extra(findme -> $lang)",
		}, {
			name:       "verify archive fork map to repo",
			inputQuery: "repo:abc123@12346f fork:yes archived:yes findme",
			mapType:    Repo,
			want:       "repo:abc123@12346f fork:yes archived:yes content:output.extra(findme -> $repo)",
		}, {
			name:       "verify archive fork map to path",
			inputQuery: "repo:abc123@12346f fork:yes archived:yes findme",
			mapType:    Path,
			want:       "repo:abc123@12346f fork:yes archived:yes content:output.extra(findme -> $path)",
		}, {
			name:       "verify archive fork map to author",
			inputQuery: "repo:abc123@12346f fork:yes archived:yes findme",
			mapType:    Author,
			want:       "repo:abc123@12346f fork:yes archived:yes content:output.extra(findme -> $author)",
		}, {
			name:       "verify archive fork map to date",
			inputQuery: "repo:abc123@12346f fork:yes archived:yes findme",
			mapType:    Date,
			want:       "repo:abc123@12346f fork:yes archived:yes content:output.extra(findme -> $date)",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ComputeInsightCommandQuery(BasicQuery(test.inputQuery), test.mapType)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(test.want, string(got)); diff != "" {
				t.Errorf("%s failed (want/got): %s", test.name, diff)
			}
		})
	}
}
