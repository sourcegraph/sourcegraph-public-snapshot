package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestSearchResults(t *testing.T) {
	limitOffset := &db.LimitOffset{Limit: maxReposToSearch() + 1}

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

		mockSearchFilesInRepos = func(args *search.Args) ([]*fileMatchResolver, *searchResultsCommon, error) {
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
		mockSearchRepositories = func(args *search.Args) ([]*searchResultResolver, *searchResultsCommon, error) {
			calledSearchRepositories = true
			return nil, &searchResultsCommon{}, nil
		}
		defer func() { mockSearchRepositories = nil }()

		calledSearchSymbols := false
		mockSearchSymbols = func(ctx context.Context, args *search.Args, limit int) (res []*fileMatchResolver, common *searchResultsCommon, err error) {
			calledSearchSymbols = true
			if want := `(foo\d).*?(bar\*)`; args.Pattern.Pattern != want {
				t.Errorf("got %q, want %q", args.Pattern.Pattern, want)
			}
			// TODO return mock results here and assert that they are output as results
			return nil, nil, nil
		}
		defer func() { mockSearchSymbols = nil }()

		calledSearchFilesInRepos := false
		mockSearchFilesInRepos = func(args *search.Args) ([]*fileMatchResolver, *searchResultsCommon, error) {
			calledSearchFilesInRepos = true
			if want := `(foo\d).*?(bar\*)`; args.Pattern.Pattern != want {
				t.Errorf("got %q, want %q", args.Pattern.Pattern, want)
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
		if calledSearchSymbols {
			t.Error("calledSearchSymbols")
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
	normalize := func(p *search.PatternInfo) {
		if len(p.IncludePatterns) == 0 {
			p.IncludePatterns = nil
		}
		if p.FileMatchLimit == 0 {
			p.FileMatchLimit = defaultMaxSearchResults
		}
	}

	tests := map[string]search.PatternInfo{
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
			ExcludePattern:         "f",
		},
		"p -file:f1 -file:f2": {
			Pattern:                "p",
			IsRegExp:               true,
			PathPatternsAreRegExps: true,
			ExcludePattern:         "f1|f2",
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
			ExcludePattern:         `\.graphql$|\.gql$`,
		},
		"p -lang:graphql -file:f": {
			Pattern:                "p",
			IsRegExp:               true,
			PathPatternsAreRegExps: true,
			ExcludePattern:         `f|(\.graphql$|\.gql$)`,
		},
	}
	for queryStr, want := range tests {
		t.Run(queryStr, func(t *testing.T) {
			query, err := query.ParseAndCheck(queryStr)
			if err != nil {
				t.Fatal(err)
			}
			sr := searchResolver{query: query}
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

func TestSearchResolver_DynamicFilters(t *testing.T) {
	repo := &types.Repo{
		URI: "testRepo",
	}

	repoMatch := &repositoryResolver{
		repo: repo,
	}

	fileMatch := &fileMatchResolver{
		JPath: "/testFile.md",
		repo:  repo,
	}

	rev := "develop"
	fileMatchRev := &fileMatchResolver{
		JPath:    "/testFile.md",
		repo:     repo,
		inputRev: &rev,
	}

	type testCase struct {
		descr                     string
		searchResults             []*searchResultResolver
		expectedDynamicFilterStrs map[string]struct{}
	}

	tests := []testCase{

		testCase{
			descr: "single repo match",
			searchResults: []*searchResultResolver{
				&searchResultResolver{
					repo: repoMatch,
				},
			},
			expectedDynamicFilterStrs: map[string]struct{}{
				`repo:^testRepo$`: struct{}{},
			},
		},

		testCase{
			descr: "single file match without revision in query",
			searchResults: []*searchResultResolver{
				&searchResultResolver{
					fileMatch: fileMatch,
				},
			},
			expectedDynamicFilterStrs: map[string]struct{}{
				`repo:^testRepo$`: struct{}{},
				`file:\.md$`:      struct{}{},
			},
		},

		testCase{
			descr: "single file match with specified revision",
			searchResults: []*searchResultResolver{
				&searchResultResolver{
					fileMatch: fileMatchRev,
				},
			},
			expectedDynamicFilterStrs: map[string]struct{}{
				`repo:^testRepo$@develop`: struct{}{},
				`file:\.md$`:              struct{}{},
			},
		},
	}

	for _, test := range tests {

		t.Run(test.descr, func(t *testing.T) {

			actualDynamicFilters := (&searchResultsResolver{results: test.searchResults}).DynamicFilters()
			actualDynamicFilterStrs := make(map[string]struct{})

			for _, filter := range actualDynamicFilters {
				actualDynamicFilterStrs[filter.Value()] = struct{}{}
			}

			if !reflect.DeepEqual(actualDynamicFilterStrs, test.expectedDynamicFilterStrs) {
				t.Errorf("actual: %v, expected: %v", actualDynamicFilterStrs, test.expectedDynamicFilterStrs)
			}
		})
	}
}

// TestSearchRevspecs tests a repository name against a list of
// repository specs with optional revspecs, and determines whether
// we get the expected error, list of matching rev specs, or list
// of clashing revspecs (if no matching rev specs were found)
func TestSearchRevspecs(t *testing.T) {
	type testCase struct {
		descr    string
		specs    []string
		repo     string
		err      error
		matched  []search.RevisionSpecifier
		clashing []search.RevisionSpecifier
	}

	tests := []testCase{
		testCase{
			descr:    "simple match",
			specs:    []string{"foo"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: ""}},
			clashing: nil,
		},
		testCase{
			descr:    "single revspec",
			specs:    []string{".*o@123456"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "123456"}},
			clashing: nil,
		},
		testCase{
			descr:    "revspec plus unspecified rev",
			specs:    []string{".*o@123456", "foo"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "123456"}},
			clashing: nil,
		},
		testCase{
			descr:    "revspec plus unspecified rev, but backwards",
			specs:    []string{".*o", "foo@123456"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "123456"}},
			clashing: nil,
		},
		testCase{
			descr:    "conflicting revspecs",
			specs:    []string{".*o@123456", "foo@234567"},
			repo:     "foo",
			err:      nil,
			matched:  nil,
			clashing: []search.RevisionSpecifier{{RevSpec: "123456"}, {RevSpec: "234567"}},
		},
		testCase{
			descr:    "overlapping revspecs",
			specs:    []string{".*o@a:b", "foo@b:c"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "b"}},
			clashing: nil,
		},
		testCase{
			descr:    "multiple overlapping revspecs",
			specs:    []string{".*o@a:b:c", "foo@b:c:d"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "b"}, {RevSpec: "c"}},
			clashing: nil,
		},
		testCase{
			descr:    "invalid regexp",
			specs:    []string{"*o@a:b"},
			repo:     "foo",
			err:      fmt.Errorf("%s", "bad request: error parsing regexp: missing argument to repetition operator: `*`"),
			matched:  nil,
			clashing: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.descr, func(t *testing.T) {
			pats, err := findPatternRevs(test.specs)
			if err != nil {
				if test.err == nil {
					t.Errorf("unexpected error: '%s'", err)
				}
				if test.err != nil && err.Error() != test.err.Error() {
					t.Errorf("incorrect error: got '%s', expected '%s'", err, test.err)
				}
				// don't try to use the pattern list if we got an error
				return
			}
			if test.err != nil {
				t.Errorf("missing expected error: wanted '%s'", test.err.Error())
			}
			matched, clashing := getRevsForMatchedRepo(api.RepoURI(test.repo), pats)
			if !reflect.DeepEqual(matched, test.matched) {
				t.Errorf("matched repo mismatch: actual: %#v, expected: %#v", matched, test.matched)
			}
			if !reflect.DeepEqual(clashing, test.clashing) {
				t.Errorf("clashing repo mismatch: actual: %#v, expected: %#v", clashing, test.clashing)
			}
		})
	}
}

func TestCompareSearchResults(t *testing.T) {
	type testCase struct {
		a       *searchResultResolver
		b       *searchResultResolver
		aIsLess bool
	}

	tests := []testCase{
		// Different repo matches
		testCase{
			a: &searchResultResolver{
				repo: &repositoryResolver{
					repo: &types.Repo{
						URI: api.RepoURI("a"),
					},
				},
			},
			b: &searchResultResolver{
				repo: &repositoryResolver{
					repo: &types.Repo{
						URI: api.RepoURI("b"),
					},
				},
			},
			aIsLess: true,
		},
		// Repo match vs file match in same repo
		testCase{
			a: &searchResultResolver{
				fileMatch: &fileMatchResolver{
					repo: &types.Repo{
						URI: api.RepoURI("a"),
					},
					JPath: "a",
				},
			},
			b: &searchResultResolver{
				repo: &repositoryResolver{
					repo: &types.Repo{
						URI: api.RepoURI("a"),
					},
				},
			},
			aIsLess: false,
		},
		// Same repo, different files
		testCase{
			a: &searchResultResolver{
				fileMatch: &fileMatchResolver{
					repo: &types.Repo{
						URI: api.RepoURI("a"),
					},
					JPath: "a",
				},
			},
			b: &searchResultResolver{
				fileMatch: &fileMatchResolver{
					repo: &types.Repo{
						URI: api.RepoURI("a"),
					},
					JPath: "b",
				},
			},
			aIsLess: true,
		},
		// different repo, same file name
		testCase{
			a: &searchResultResolver{
				fileMatch: &fileMatchResolver{
					repo: &types.Repo{
						URI: api.RepoURI("a"),
					},
					JPath: "a",
				},
			},
			b: &searchResultResolver{
				fileMatch: &fileMatchResolver{
					repo: &types.Repo{
						URI: api.RepoURI("b"),
					},
					JPath: "a",
				},
			},
			aIsLess: true,
		},
	}

	for i, test := range tests {
		got := compareSearchResults(test.a, test.b)
		if got != test.aIsLess {
			t.Errorf("[%d] incorrect comparison. got %t, expected %t", i, got, test.aIsLess)
		}
	}
}
