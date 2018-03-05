package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
)

func TestSearchResults(t *testing.T) {
	limitOffset := &db.LimitOffset{Limit: maxReposToSearch + 1}

	createSearchResolver := func(t *testing.T, query string) *searchResolver {
		r, err := (&schemaResolver{}).Search(&struct{ Query string }{Query: query})
		if err != nil {
			t.Fatal("Search:", err)
		}
		return r
	}
	getResults := func(t *testing.T, query string) []string {
		r := createSearchResolver(t, query)
		results, err := r.Results(context.Background())
		if err != nil {
			t.Fatal("Results:", err)
		}
		resultDescriptions := make([]string, len(results.results))
		for i, result := range results.results {
			// NOTE: Only supports one match per line. If we need to test other cases,
			// just remove that assumption in the following line of code.
			switch {
			case result.repo != nil:
				resultDescriptions[i] = fmt.Sprintf("repo:%s", result.repo.repo.URI)
			case result.fileMatch != nil:
				resultDescriptions[i] = fmt.Sprintf("%s:%d", result.fileMatch.JPath, result.fileMatch.JLineMatches[0].JLineNumber)
			}
		}
		return resultDescriptions
	}
	testCallResults := func(t *testing.T, query string, want []string) {
		results := getResults(t, query)
		if !reflect.DeepEqual(results, want) {
			t.Errorf("got %v, want %v", results, want)
		}
	}

	t.Run("repo: only", func(t *testing.T) {
		var calledReposList bool
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
			calledReposList = true
			if want := (db.ReposListOptions{Enabled: true, IncludePatterns: []string{"r", "p"}, LimitOffset: limitOffset}); !reflect.DeepEqual(op, want) {
				t.Fatalf("got %+v, want %+v", op, want)
			}
			return []*types.Repo{{URI: "repo"}}, nil
		}
		db.Mocks.Repos.MockGetByURI(t, "repo", 1)

		mockSearchFilesInRepos = func(args *repoSearchArgs) ([]*fileMatchResolver, *searchResultsCommon, error) {
			return nil, &searchResultsCommon{}, nil
		}
		defer func() { mockSearchFilesInRepos = nil }()

		testCallResults(t, `repo:r repo:p`, []string{"repo:repo"})
		if !calledReposList {
			t.Error("!calledReposList")
		}
	})

	t.Run("multiple terms", func(t *testing.T) {

		var calledReposList bool
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
			calledReposList = true
			if want := (db.ReposListOptions{Enabled: true, LimitOffset: limitOffset}); !reflect.DeepEqual(op, want) {
				t.Fatalf("got %+v, want %+v", op, want)
			}
			return []*types.Repo{{URI: "repo"}}, nil
		}
		defer func() { db.Mocks = db.MockStores{} }()
		db.Mocks.Repos.MockGetByURI(t, "repo", 1)

		calledSearchRepositories := false
		mockSearchRepositories = func(args *repoSearchArgs) ([]*searchResultResolver, *searchResultsCommon, error) {
			calledSearchRepositories = true
			return nil, &searchResultsCommon{}, nil
		}
		defer func() { mockSearchRepositories = nil }()

		calledSearchSymbols := false
		mockSearchSymbols = func(ctx context.Context, args *repoSearchArgs, query searchquery.Query, limit int) (res []*symbolResolver, err error) {
			calledSearchSymbols = true
			if want := `(foo\d).*?(bar\*)`; args.query.Pattern != want {
				t.Errorf("got %q, want %q", args.query.Pattern, want)
			}
			// TODO return mock results here and assert that they are output as results
			return nil, nil
		}
		defer func() { mockSearchSymbols = nil }()

		calledSearchFilesInRepos := false
		mockSearchFilesInRepos = func(args *repoSearchArgs) ([]*fileMatchResolver, *searchResultsCommon, error) {
			calledSearchFilesInRepos = true
			if want := `(foo\d).*?(bar\*)`; args.query.Pattern != want {
				t.Errorf("got %q, want %q", args.query.Pattern, want)
			}
			return []*fileMatchResolver{
				{uri: "git://repo?rev#dir/file", JPath: "dir/file", JLineMatches: []*lineMatch{{JLineNumber: 123}}},
			}, &searchResultsCommon{}, nil
		}
		defer func() { mockSearchFilesInRepos = nil }()

		testCallResults(t, `foo\d "bar*"`, []string{"dir/file:123"})
		if !calledReposList {
			t.Error("!calledReposList")
		}
		if !calledSearchRepositories {
			t.Error("!calledSearchRepositories")
		}
		if !calledSearchFilesInRepos {
			t.Error("!calledSearchFilesInRepos")
		}
		if !calledSearchSymbols {
			t.Error("!calledSearchSymbols")
		}
	})
}

func TestRegexpPatternMatchingExprsInOrder(t *testing.T) {
	got := regexpPatternMatchingExprsInOrder([]string{})
	if want := ""; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got = regexpPatternMatchingExprsInOrder([]string{"a"})
	if want := "a"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got = regexpPatternMatchingExprsInOrder([]string{"a", "b|c"})
	if want := "(a).*?(b|c)"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSearchResolver_getPatternInfo(t *testing.T) {
	normalize := func(p *patternInfo) {
		if len(p.IncludePatterns) == 0 {
			p.IncludePatterns = nil
		}
		if p.FileMatchLimit == 0 {
			p.FileMatchLimit = defaultMaxSearchResults
		}
	}

	tests := map[string]patternInfo{
		"p": {
			Pattern:                "p",
			IsRegExp:               true,
			PathPatternsAreRegExps: true,
		},
		"p1 p2": {
			Pattern:                "(p1).*?(p2)",
			IsRegExp:               true,
			PathPatternsAreRegExps: true,
		},
		"p case:yes": {
			Pattern:                      "p",
			IsRegExp:                     true,
			IsCaseSensitive:              true,
			PathPatternsAreRegExps:       true,
			PathPatternsAreCaseSensitive: true,
		},
		"p file:f": {
			Pattern:                "p",
			IsRegExp:               true,
			PathPatternsAreRegExps: true,
			IncludePatterns:        []string{"f"},
		},
		"p file:f1 file:f2": {
			Pattern:                "p",
			IsRegExp:               true,
			PathPatternsAreRegExps: true,
			IncludePatterns:        []string{"f1", "f2"},
		},
		"p -file:f": {
			Pattern:                "p",
			IsRegExp:               true,
			PathPatternsAreRegExps: true,
			ExcludePattern:         strptr("f"),
		},
		"p -file:f1 -file:f2": {
			Pattern:                "p",
			IsRegExp:               true,
			PathPatternsAreRegExps: true,
			ExcludePattern:         strptr("f1|f2"),
		},
		"p lang:graphql": {
			Pattern:                "p",
			IsRegExp:               true,
			PathPatternsAreRegExps: true,
			IncludePatterns:        []string{`\.graphql$|\.gql$`},
		},
		"p lang:graphql file:f": {
			Pattern:                "p",
			IsRegExp:               true,
			PathPatternsAreRegExps: true,
			IncludePatterns:        []string{"f", `\.graphql$|\.gql$`},
		},
		"p -lang:graphql file:f": {
			Pattern:                "p",
			IsRegExp:               true,
			PathPatternsAreRegExps: true,
			IncludePatterns:        []string{"f"},
			ExcludePattern:         strptr(`\.graphql$|\.gql$`),
		},
		"p -lang:graphql -file:f": {
			Pattern:                "p",
			IsRegExp:               true,
			PathPatternsAreRegExps: true,
			ExcludePattern:         strptr(`f|(\.graphql$|\.gql$)`),
		},
	}
	for queryStr, want := range tests {
		t.Run(queryStr, func(t *testing.T) {
			query, err := searchquery.ParseAndCheck(queryStr)
			if err != nil {
				t.Fatal(err)
			}
			sr := searchResolver{query: *query}
			p, err := sr.getPatternInfo()
			if err != nil {
				t.Fatal(err)
			}
			normalize(p)
			normalize(&want)
			if !reflect.DeepEqual(*p, want) {
				t.Errorf("\ngot  %+v\nwant %+v", *p, want)
			}
		})
	}
}
