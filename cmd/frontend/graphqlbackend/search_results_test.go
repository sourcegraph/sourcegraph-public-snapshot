package graphqlbackend

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/zoekt"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchquerytypes "github.com/sourcegraph/sourcegraph/internal/search/query/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

var mockCount = func(_ context.Context, options db.ReposListOptions) (int, error) { return 0, nil }

func assertEqual(t *testing.T, got, want interface{}) {
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("(-want +got):\n%s", diff)
	}
}

func TestSearchResults(t *testing.T) {
	limitOffset := &db.LimitOffset{Limit: maxReposToSearch() + 1}

	getResults := func(t *testing.T, query, version string) []string {
		r, err := (&schemaResolver{}).Search(context.Background(), &SearchArgs{Query: query, Version: version})
		if err != nil {
			t.Fatal("Search:", err)
		}
		results, err := r.Results(context.Background())
		if err != nil {
			t.Fatal("Results:", err)
		}
		resultDescriptions := make([]string, len(results.SearchResults))
		for i, result := range results.SearchResults {
			// NOTE: Only supports one match per line. If we need to test other cases,
			// just remove that assumption in the following line of code.
			switch m := result.(type) {
			case *RepositoryResolver:
				resultDescriptions[i] = fmt.Sprintf("repo:%s", m.repo.Name)
			case *FileMatchResolver:
				resultDescriptions[i] = fmt.Sprintf("%s:%d", m.JPath, m.JLineMatches[0].JLineNumber)
			default:
				t.Fatal("unexpected result type", result)
			}
		}
		return resultDescriptions
	}
	testCallResults := func(t *testing.T, query, version string, want []string) {
		results := getResults(t, query, version)
		if !reflect.DeepEqual(results, want) {
			t.Errorf("got %v, want %v", results, want)
		}
	}

	searchVersions := []string{"V1", "V2"}

	t.Run("repo: only", func(t *testing.T) {
		mockDecodedViewerFinalSettings = &schema.Settings{}
		defer func() { mockDecodedViewerFinalSettings = nil }()

		var calledReposList bool
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
			calledReposList = true

			// Validate that the following options are invariant
			// when calling the DB through Repos.List, no matter how
			// many times it is called for a single Search(...) operation.
			assertEqual(t, op.OnlyRepoIDs, true)
			assertEqual(t, op.LimitOffset, limitOffset)
			assertEqual(t, op.IncludePatterns, []string{"r", "p"})

			return []*types.Repo{{ID: 1, Name: "repo"}}, nil
		}
		db.Mocks.Repos.MockGetByName(t, "repo", 1)
		db.Mocks.Repos.MockGet(t, 1)
		db.Mocks.Repos.Count = mockCount

		mockSearchFilesInRepos = func(args *search.TextParameters) ([]*FileMatchResolver, *searchResultsCommon, error) {
			return nil, &searchResultsCommon{repos: []*types.Repo{{ID: 1, Name: "repo"}}}, nil
		}
		defer func() { mockSearchFilesInRepos = nil }()

		for _, v := range searchVersions {
			testCallResults(t, `repo:r repo:p`, v, []string{"repo:repo"})
			if !calledReposList {
				t.Error("!calledReposList")
			}
		}

	})

	t.Run("multiple terms regexp", func(t *testing.T) {
		mockDecodedViewerFinalSettings = &schema.Settings{}
		defer func() { mockDecodedViewerFinalSettings = nil }()

		var calledReposList bool
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
			calledReposList = true

			// Validate that the following options are invariant
			// when calling the DB through Repos.List, no matter how
			// many times it is called for a single Search(...) operation.
			assertEqual(t, op.OnlyRepoIDs, true)
			assertEqual(t, op.LimitOffset, limitOffset)

			return []*types.Repo{{ID: 1, Name: "repo"}}, nil
		}
		defer func() { db.Mocks = db.MockStores{} }()
		db.Mocks.Repos.MockGetByName(t, "repo", 1)
		db.Mocks.Repos.MockGet(t, 1)
		db.Mocks.Repos.Count = mockCount

		calledSearchRepositories := false
		mockSearchRepositories = func(args *search.TextParameters) ([]SearchResultResolver, *searchResultsCommon, error) {
			calledSearchRepositories = true
			return nil, &searchResultsCommon{}, nil
		}
		defer func() { mockSearchRepositories = nil }()

		calledSearchSymbols := false
		mockSearchSymbols = func(ctx context.Context, args *search.TextParameters, limit int) (res []*FileMatchResolver, common *searchResultsCommon, err error) {
			calledSearchSymbols = true
			if want := `(foo\d).*?(bar\*)`; args.PatternInfo.Pattern != want {
				t.Errorf("got %q, want %q", args.PatternInfo.Pattern, want)
			}
			// TODO return mock results here and assert that they are output as results
			return nil, nil, nil
		}
		defer func() { mockSearchSymbols = nil }()

		calledSearchFilesInRepos := false
		mockSearchFilesInRepos = func(args *search.TextParameters) ([]*FileMatchResolver, *searchResultsCommon, error) {
			calledSearchFilesInRepos = true
			if want := `(foo\d).*?(bar\*)`; args.PatternInfo.Pattern != want {
				t.Errorf("got %q, want %q", args.PatternInfo.Pattern, want)
			}
			return []*FileMatchResolver{
				{
					uri:          "git://repo?rev#dir/file",
					JPath:        "dir/file",
					JLineMatches: []*lineMatch{{JLineNumber: 123}},
					Repo:         &RepositoryResolver{repo: &types.Repo{ID: 1}},
				},
			}, &searchResultsCommon{repos: []*types.Repo{{ID: 1}}}, nil

		}
		defer func() { mockSearchFilesInRepos = nil }()

		testCallResults(t, `foo\d "bar*"`, "V1", []string{"dir/file:123"})
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

	t.Run("multiple terms literal", func(t *testing.T) {
		mockDecodedViewerFinalSettings = &schema.Settings{}
		defer func() { mockDecodedViewerFinalSettings = nil }()

		var calledReposList bool
		db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
			calledReposList = true

			// Validate that the following options are invariant
			// when calling the DB through Repos.List, no matter how
			// many times it is called for a single Search(...) operation.
			assertEqual(t, op.OnlyRepoIDs, true)
			assertEqual(t, op.LimitOffset, limitOffset)

			return []*types.Repo{{ID: 1, Name: "repo"}}, nil
		}
		defer func() { db.Mocks = db.MockStores{} }()
		db.Mocks.Repos.MockGetByName(t, "repo", 1)
		db.Mocks.Repos.MockGet(t, 1)
		db.Mocks.Repos.Count = mockCount

		calledSearchRepositories := false
		mockSearchRepositories = func(args *search.TextParameters) ([]SearchResultResolver, *searchResultsCommon, error) {
			calledSearchRepositories = true
			return nil, &searchResultsCommon{}, nil
		}
		defer func() { mockSearchRepositories = nil }()

		calledSearchSymbols := false
		mockSearchSymbols = func(ctx context.Context, args *search.TextParameters, limit int) (res []*FileMatchResolver, common *searchResultsCommon, err error) {
			calledSearchSymbols = true
			if want := `"foo\\d \"bar*\""`; args.PatternInfo.Pattern != want {
				t.Errorf("got %q, want %q", args.PatternInfo.Pattern, want)
			}
			// TODO return mock results here and assert that they are output as results
			return nil, nil, nil
		}
		defer func() { mockSearchSymbols = nil }()

		calledSearchFilesInRepos := false
		mockSearchFilesInRepos = func(args *search.TextParameters) ([]*FileMatchResolver, *searchResultsCommon, error) {
			calledSearchFilesInRepos = true
			if want := `foo\\d "bar\*"`; args.PatternInfo.Pattern != want {
				t.Errorf("got %q, want %q", args.PatternInfo.Pattern, want)
			}
			return []*FileMatchResolver{
				{
					uri:          "git://repo?rev#dir/file",
					JPath:        "dir/file",
					JLineMatches: []*lineMatch{{JLineNumber: 123}},
					Repo:         &RepositoryResolver{repo: &types.Repo{ID: 1}},
				},
			}, &searchResultsCommon{repos: []*types.Repo{{ID: 1}}}, nil
		}
		defer func() { mockSearchFilesInRepos = nil }()

		testCallResults(t, `foo\d "bar*"`, "V2", []string{"dir/file:123"})
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

	t.Run("test start time is not null when alert thrown", func(t *testing.T) {
		mockDecodedViewerFinalSettings = &schema.Settings{}
		defer func() { mockDecodedViewerFinalSettings = nil }()

		for _, v := range searchVersions {
			r, err := (&schemaResolver{}).Search(context.Background(), &SearchArgs{Query: `repo:*`, Version: v})
			if err != nil {
				t.Fatal("Search:", err)
			}

			results, err := r.Results(context.Background())
			if err != nil {
				t.Fatal("Search: ", err)
			}

			if results.start.IsZero() {
				t.Error("Start value is not set")
			}
		}
	})
}

func TestOrderedFuzzyRegexp(t *testing.T) {
	got := orderedFuzzyRegexp([]string{})
	if want := ""; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got = orderedFuzzyRegexp([]string{"a"})
	if want := "a"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got = orderedFuzzyRegexp([]string{"a", "b|c"})
	if want := "(a).*?(b|c)"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestProcessSearchPattern(t *testing.T) {
	cases := []struct {
		Name    string
		Pattern string
		Opts    *getPatternInfoOptions
		Want    string
	}{
		{
			Name:    "Regexp, no content field",
			Pattern: `search me`,
			Opts:    &getPatternInfoOptions{},
			Want:    "(search).*?(me)",
		},
		{
			Name:    "Regexp with content field",
			Pattern: `content:search`,
			Opts:    &getPatternInfoOptions{},
			Want:    "search",
		},
		{
			Name:    "Regexp with quoted content field",
			Pattern: `content:"search me"`,
			Opts:    &getPatternInfoOptions{},
			Want:    "search me",
		},
		{
			Name:    "Regexp with content field ignores default pattern",
			Pattern: `content:"search me" ignored`,
			Opts:    &getPatternInfoOptions{},
			Want:    "search me",
		},
		{
			Name:    "Literal with quoted content field means double quotes are not part of the pattern",
			Pattern: `content:"content:"`,
			Opts:    &getPatternInfoOptions{performLiteralSearch: true},
			Want:    "content:",
		},
		{
			Name:    "Literal with quoted content field containing quotes",
			Pattern: `content:"\"content:\""`,
			Opts:    &getPatternInfoOptions{performLiteralSearch: true},
			Want:    "\"content:\"",
		},
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			q, _ := query.ParseAndCheck(tt.Pattern)
			got, _, _, _ := processSearchPattern(q, tt.Opts)
			if got != tt.Want {
				t.Fatalf("got %s\nwant %s", got, tt.Want)
			}
		})
	}
}

func TestIsPatternNegated(t *testing.T) {
	cases := []struct {
		name    string
		pattern string
		want    bool
	}{
		{
			name:    "simple negated pattern",
			pattern: "-content:foo",
			want:    true,
		},
		{
			name:    "compound query with negated content as first term",
			pattern: "-content:foo and bar",
			want:    false,
		},
		{
			name:    "compound query with negated content as last term",
			pattern: "bar and -content:foo",
			want:    false,
		},
		{
			name:    "simple query with content field but without negation",
			pattern: "content:foo",
			want:    false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			q, err := query.ProcessAndOr(tt.pattern,
				query.ParserOptions{SearchType: query.SearchTypeLiteral, Globbing: false})
			if err != nil {
				t.Fatalf(err.Error())
			}

			if andOrQuery, ok := q.(*query.AndOrQuery); ok {
				got := isPatternNegated(andOrQuery.Query)

				if got != tt.want {
					t.Fatalf("got %t\nwant %t", got, tt.want)
				}
			} else {
				t.Fatalf("the query should have been an AndOrQuery")
			}
		})
	}
}

func TestProcessSearchPatternAndOr(t *testing.T) {
	cases := []struct {
		name                string
		pattern             string
		searchType          query.SearchType
		opts                *getPatternInfoOptions
		wantPattern         string
		wantIsRegExp        bool
		wantIsStructuralPat bool
		wantIsNegated       bool
	}{
		{
			name:                "Simple content",
			pattern:             `content:foo`,
			searchType:          query.SearchTypeLiteral,
			opts:                &getPatternInfoOptions{},
			wantPattern:         "foo",
			wantIsRegExp:        true,
			wantIsStructuralPat: false,
			wantIsNegated:       false,
		},
		{
			name:                "Negated content",
			pattern:             `-content:foo`,
			searchType:          query.SearchTypeLiteral,
			opts:                &getPatternInfoOptions{},
			wantPattern:         "foo",
			wantIsRegExp:        true,
			wantIsStructuralPat: false,
			wantIsNegated:       true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			q, err := query.ProcessAndOr(tt.pattern,
				query.ParserOptions{SearchType: tt.searchType, Globbing: false})
			if err != nil {
				t.Fatalf(err.Error())
			}

			pattern, isRegExp, isStructuralPat, isNegated := processSearchPattern(q, tt.opts)

			if want := tt.wantPattern; pattern != want {
				t.Fatalf("got %s\nwant %s", pattern, want)
			}

			if want := tt.wantIsRegExp; isRegExp != want {
				t.Fatalf("got %t\nwant %t", isRegExp, want)
			}

			if want := tt.wantIsStructuralPat; isStructuralPat != want {
				t.Fatalf("got %t\nwant %t", isStructuralPat, want)
			}

			if want := tt.wantIsNegated; isNegated != want {
				t.Fatalf("got %t\nwant %t", isNegated, want)
			}
		})
	}
}

func TestSearchResolver_getPatternInfo(t *testing.T) {
	normalize := func(p *search.TextPatternInfo) {
		if len(p.IncludePatterns) == 0 {
			p.IncludePatterns = nil
		}
		if p.FileMatchLimit == 0 {
			p.FileMatchLimit = defaultMaxSearchResults
		}
	}

	tests := map[string]search.TextPatternInfo{
		"p": {
			Pattern:  "p",
			IsRegExp: true,
		},
		"p1 p2": {
			Pattern:  "(p1).*?(p2)",
			IsRegExp: true,
		},
		"p case:yes": {
			Pattern:                      "p",
			IsRegExp:                     true,
			IsCaseSensitive:              true,
			PathPatternsAreCaseSensitive: true,
		},
		"p file:f": {
			Pattern:         "p",
			IsRegExp:        true,
			IncludePatterns: []string{"f"},
		},
		"p file:f1 file:f2": {
			Pattern:         "p",
			IsRegExp:        true,
			IncludePatterns: []string{"f1", "f2"},
		},
		"p -file:f": {
			Pattern:        "p",
			IsRegExp:       true,
			ExcludePattern: "f",
		},
		"p -file:f1 -file:f2": {
			Pattern:        "p",
			IsRegExp:       true,
			ExcludePattern: "f1|f2",
		},
		"p lang:graphql": {
			Pattern:         "p",
			IsRegExp:        true,
			IncludePatterns: []string{`\.graphql$|\.gql$|\.graphqls$`},
			Languages:       []string{"graphql"},
		},
		"p lang:graphql file:f": {
			Pattern:         "p",
			IsRegExp:        true,
			IncludePatterns: []string{"f", `\.graphql$|\.gql$|\.graphqls$`},
			Languages:       []string{"graphql"},
		},
		"p -lang:graphql file:f": {
			Pattern:         "p",
			IsRegExp:        true,
			IncludePatterns: []string{"f"},
			ExcludePattern:  `\.graphql$|\.gql$|\.graphqls$`,
		},
		"p -lang:graphql -file:f": {
			Pattern:        "p",
			IsRegExp:       true,
			ExcludePattern: `f|(\.graphql$|\.gql$|\.graphqls$)`,
		},
	}
	for queryStr, want := range tests {
		t.Run(queryStr, func(t *testing.T) {
			query, err := query.ParseAndCheck(queryStr)
			if err != nil {
				t.Fatal(err)
			}
			sr := searchResolver{query: query}
			p, err := sr.getPatternInfo(nil)
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
	repo := &types.Repo{Name: "testRepo"}

	repoMatch := &RepositoryResolver{
		repo: repo,
	}

	fileMatch := &FileMatchResolver{
		JPath: "/testFile.md",
		Repo:  repoMatch,
	}

	tsFileMatch := &FileMatchResolver{
		JPath: "/testFile.ts",
		Repo:  repoMatch,
	}

	tsxFileMatch := &FileMatchResolver{
		JPath: "/testFile.tsx",
		Repo:  repoMatch,
	}

	ignoreListFileMatch := &FileMatchResolver{
		JPath: "/.gitignore",
		Repo:  repoMatch,
	}

	goTestFileMatch := &FileMatchResolver{
		JPath: "/foo_test.go",
		Repo:  repoMatch,
	}

	nodeModulesMatchSub := &FileMatchResolver{
		JPath: "/anything/node_modules/testFile.md",
		Repo:  repoMatch,
	}

	nodeModulesMatchRoot := &FileMatchResolver{
		JPath: "/node_modules/testFile.md",
		Repo:  repoMatch,
	}

	rev := "develop3.0"
	fileMatchRev := &FileMatchResolver{
		JPath:    "/testFile.md",
		Repo:     repoMatch,
		InputRev: &rev,
	}

	type testCase struct {
		descr                             string
		searchResults                     []SearchResultResolver
		expectedDynamicFilterStrsRegexp   map[string]struct{}
		expectedDynamicFilterStrsGlobbing map[string]struct{}
	}

	tests := []testCase{

		{
			descr:         "single repo match",
			searchResults: []SearchResultResolver{repoMatch},
			expectedDynamicFilterStrsRegexp: map[string]struct{}{
				`repo:^testRepo$`: {},
			},
			expectedDynamicFilterStrsGlobbing: map[string]struct{}{
				`repo:testRepo`: {},
			},
		},

		{
			descr:         "single file match without revision in query",
			searchResults: []SearchResultResolver{fileMatch},
			expectedDynamicFilterStrsRegexp: map[string]struct{}{
				`repo:^testRepo$`: {},
				`lang:markdown`:   {},
			},
			expectedDynamicFilterStrsGlobbing: map[string]struct{}{
				`repo:testRepo`: {},
				`lang:markdown`: {},
			},
		},

		{
			descr:         "single file match with specified revision",
			searchResults: []SearchResultResolver{fileMatchRev},
			expectedDynamicFilterStrsRegexp: map[string]struct{}{
				`repo:^testRepo$@develop3.0`: {},
				`lang:markdown`:              {},
			},
			expectedDynamicFilterStrsGlobbing: map[string]struct{}{
				`repo:testRepo@develop3.0`: {},
				`lang:markdown`:            {},
			},
		},
		{
			descr:         "file match from a language with two file extensions, using first extension",
			searchResults: []SearchResultResolver{tsFileMatch},
			expectedDynamicFilterStrsRegexp: map[string]struct{}{
				`repo:^testRepo$`: {},
				`lang:typescript`: {},
			},
			expectedDynamicFilterStrsGlobbing: map[string]struct{}{
				`repo:testRepo`:   {},
				`lang:typescript`: {},
			},
		},
		{
			descr:         "file match from a language with two file extensions, using second extension",
			searchResults: []SearchResultResolver{tsxFileMatch},
			expectedDynamicFilterStrsRegexp: map[string]struct{}{
				`repo:^testRepo$`: {},
				`lang:typescript`: {},
			},
			expectedDynamicFilterStrsGlobbing: map[string]struct{}{
				`repo:testRepo`:   {},
				`lang:typescript`: {},
			},
		},
		{
			descr:         "file match which matches one of the common file filters",
			searchResults: []SearchResultResolver{nodeModulesMatchSub},
			expectedDynamicFilterStrsRegexp: map[string]struct{}{
				`repo:^testRepo$`:          {},
				`-file:(^|/)node_modules/`: {},
				`lang:markdown`:            {},
			},
			expectedDynamicFilterStrsGlobbing: map[string]struct{}{
				`repo:testRepo`: {},
				`-file:node_modules/** -file:**/node_modules/**`: {},
				`lang:markdown`: {},
			},
		},
		{
			descr:         "file match which matches one of the common file filters",
			searchResults: []SearchResultResolver{nodeModulesMatchRoot},
			expectedDynamicFilterStrsRegexp: map[string]struct{}{
				`repo:^testRepo$`:          {},
				`-file:(^|/)node_modules/`: {},
				`lang:markdown`:            {},
			},
			expectedDynamicFilterStrsGlobbing: map[string]struct{}{
				`repo:testRepo`: {},
				`-file:node_modules/** -file:**/node_modules/**`: {},
				`lang:markdown`: {},
			},
		},
		{
			descr:         "file match which matches one of the common file filters",
			searchResults: []SearchResultResolver{goTestFileMatch},
			expectedDynamicFilterStrsRegexp: map[string]struct{}{
				`repo:^testRepo$`:  {},
				`-file:_test\.go$`: {},
				`lang:go`:          {},
			},
			expectedDynamicFilterStrsGlobbing: map[string]struct{}{
				`repo:testRepo`:    {},
				`-file:**_test.go`: {},
				`lang:go`:          {},
			},
		},

		// If there are no search results, no filters should be displayed.
		{
			descr:                             "no results",
			searchResults:                     []SearchResultResolver{},
			expectedDynamicFilterStrsRegexp:   map[string]struct{}{},
			expectedDynamicFilterStrsGlobbing: map[string]struct{}{},
		},
		{
			descr:         "values containing spaces are quoted",
			searchResults: []SearchResultResolver{ignoreListFileMatch},
			expectedDynamicFilterStrsRegexp: map[string]struct{}{
				`repo:^testRepo$`:    {},
				`lang:"ignore list"`: {},
			},
			expectedDynamicFilterStrsGlobbing: map[string]struct{}{
				`repo:testRepo`:      {},
				`lang:"ignore list"`: {},
			},
		},
	}

	mockDecodedViewerFinalSettings = &schema.Settings{}
	defer func() { mockDecodedViewerFinalSettings = nil }()

	var expectedDynamicFilterStrs map[string]struct{}
	for _, test := range tests {
		t.Run(test.descr, func(t *testing.T) {
			for _, globbing := range []bool{true, false} {
				mockDecodedViewerFinalSettings.SearchGlobbing = &globbing
				actualDynamicFilters := (&SearchResultsResolver{SearchResults: test.searchResults}).DynamicFilters(context.Background())
				actualDynamicFilterStrs := make(map[string]struct{})

				for _, filter := range actualDynamicFilters {
					actualDynamicFilterStrs[filter.Value()] = struct{}{}
				}

				if globbing {
					expectedDynamicFilterStrs = test.expectedDynamicFilterStrsGlobbing
				} else {
					expectedDynamicFilterStrs = test.expectedDynamicFilterStrsRegexp
				}

				if diff := cmp.Diff(expectedDynamicFilterStrs, actualDynamicFilterStrs); diff != "" {
					t.Errorf("mismatch (-want, +got):\n%s", diff)
				}
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
		{
			descr:    "simple match",
			specs:    []string{"foo"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: ""}},
			clashing: nil,
		},
		{
			descr:    "single revspec",
			specs:    []string{".*o@123456"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "123456"}},
			clashing: nil,
		},
		{
			descr:    "revspec plus unspecified rev",
			specs:    []string{".*o@123456", "foo"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "123456"}},
			clashing: nil,
		},
		{
			descr:    "revspec plus unspecified rev, but backwards",
			specs:    []string{".*o", "foo@123456"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "123456"}},
			clashing: nil,
		},
		{
			descr:    "conflicting revspecs",
			specs:    []string{".*o@123456", "foo@234567"},
			repo:     "foo",
			err:      nil,
			matched:  nil,
			clashing: []search.RevisionSpecifier{{RevSpec: "123456"}, {RevSpec: "234567"}},
		},
		{
			descr:    "overlapping revspecs",
			specs:    []string{".*o@a:b", "foo@b:c"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "b"}},
			clashing: nil,
		},
		{
			descr:    "multiple overlapping revspecs",
			specs:    []string{".*o@a:b:c", "foo@b:c:d"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "b"}, {RevSpec: "c"}},
			clashing: nil,
		},
		{
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
			matched, clashing := getRevsForMatchedRepo(api.RepoName(test.repo), pats)
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
		a       SearchResultResolver
		b       SearchResultResolver
		aIsLess bool
	}

	tests := []testCase{{
		// Different repo matches
		a: &RepositoryResolver{
			repo: &types.Repo{Name: api.RepoName("a")},
		},
		b: &RepositoryResolver{
			repo: &types.Repo{Name: api.RepoName("b")},
		},
		aIsLess: true,
	}, {
		// Repo match vs file match in same repo
		a: &FileMatchResolver{
			Repo: &RepositoryResolver{repo: &types.Repo{Name: "a"}},

			JPath: "a",
		},
		b: &RepositoryResolver{
			repo: &types.Repo{Name: api.RepoName("a")},
		},
		aIsLess: false,
	}, {
		// Same repo, different files
		a: &FileMatchResolver{
			Repo: &RepositoryResolver{repo: &types.Repo{Name: "a"}},

			JPath: "a",
		},
		b: &FileMatchResolver{
			Repo: &RepositoryResolver{repo: &types.Repo{Name: "a"}},

			JPath: "b",
		},
		aIsLess: true,
	}, {
		// different repo, same file name
		a: &FileMatchResolver{
			Repo: &RepositoryResolver{repo: &types.Repo{Name: "a"}},

			JPath: "a",
		},
		b: &FileMatchResolver{
			Repo: &RepositoryResolver{repo: &types.Repo{Name: "b"}},

			JPath: "a",
		},
		aIsLess: true,
	}}

	for i, test := range tests {
		got := compareSearchResults(test.a, test.b)
		if got != test.aIsLess {
			t.Errorf("[%d] incorrect comparison. got %t, expected %t", i, got, test.aIsLess)
		}
	}
}

func TestLonger(t *testing.T) {
	N := 2
	noise := time.Nanosecond
	for dt := time.Millisecond + noise; dt < time.Hour; dt += time.Millisecond {
		dt2 := longer(N, dt)
		if dt2 < time.Duration(N)*dt {
			t.Fatalf("longer(%v)=%v < 2*%v, want more", dt, dt2, dt)
		}
		if strings.Contains(dt2.String(), ".") {
			t.Fatalf("longer(%v).String() = %q contains an unwanted decimal point, want a nice round duration", dt, dt2)
		}
		lowest := 2 * time.Second
		if dt2 < lowest {
			t.Fatalf("longer(%v) = %v < %s, too short", dt, dt2, lowest)
		}
	}
}

func TestRoundStr(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "empty",
			s:    "",
			want: "",
		},
		{
			name: "simple",
			s:    "19s",
			want: "19s",
		},
		{
			name: "decimal",
			s:    "19.99s",
			want: "20s",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roundStr(tt.s); got != tt.want {
				t.Errorf("roundStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateRepoHasFileUsage(t *testing.T) {
	q, err := query.ParseAndCheck("repohasfile:test type:symbol")
	if err != nil {
		t.Fatal(err)
	}
	err = validateRepoHasFileUsage(q)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}

	validQueries := []string{
		"repohasfile:go",
		"repohasfile:go error",
		"repohasfile:test type:repo .",
		"type:repo",
		"repohasfile",
		"foo bar type:repo",
		"repohasfile:test type:path .",
		"repohasfile:test type:symbol .",
		"foo",
		"bar",
		"\"repohasfile\"",
	}
	for _, validQuery := range validQueries {
		q, err = query.ParseAndCheck(validQuery)
		if err != nil {
			t.Fatal(err)
		}
		err = validateRepoHasFileUsage(q)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
}

func TestSearchResultsHydration(t *testing.T) {
	id := 42
	repoName := "reponame-foobar"
	fileName := "foobar.go"

	repoWithIDs := &types.Repo{

		ID:   api.RepoID(id),
		Name: api.RepoName(repoName),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          repoName,
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		}}

	hydratedRepo := &types.Repo{

		ID:           repoWithIDs.ID,
		ExternalRepo: repoWithIDs.ExternalRepo,
		Name:         repoWithIDs.Name,

		RepoFields: &types.RepoFields{
			URI:         fmt.Sprintf("github.com/my-org/%s", repoWithIDs.Name),
			Description: "This is a description of a repository",
			Language:    "monkey",
			Fork:        false,
		}}

	mockDecodedViewerFinalSettings = &schema.Settings{}
	defer func() { mockDecodedViewerFinalSettings = nil }()

	db.Mocks.Repos.Get = func(ctx context.Context, id api.RepoID) (*types.Repo, error) {
		return hydratedRepo, nil
	}

	db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
		return []*types.Repo{repoWithIDs}, nil
	}
	db.Mocks.Repos.Count = mockCount

	defer func() { db.Mocks = db.MockStores{} }()

	zoektRepo := &zoekt.RepoListEntry{
		Repository: zoekt.Repository{
			Name:     string(repoWithIDs.Name),
			Branches: []zoekt.RepositoryBranch{{Name: "HEAD", Version: "deadbeef"}},
		},
	}

	zoektFileMatches := []zoekt.FileMatch{{
		Score:      5.0,
		FileName:   fileName,
		Repository: string(repoWithIDs.Name), // Important: this needs to match a name in `repos`
		Branches:   []string{"master"},
		LineMatches: []zoekt.LineMatch{
			{
				Line: nil,
			},
		},
		Checksum: []byte{0, 1, 2},
	}}

	z := &searchbackend.Zoekt{
		Client: &fakeSearcher{
			repos:  []*zoekt.RepoListEntry{zoektRepo},
			result: &zoekt.SearchResult{Files: zoektFileMatches},
		},
		DisableCache: true,
	}

	ctx := context.Background()

	q, err := query.ParseAndCheck(`foobar index:only count:350`)
	if err != nil {
		t.Fatal(err)
	}
	resolver := &searchResolver{query: q, zoekt: z}
	results, err := resolver.Results(ctx)
	if err != nil {
		t.Fatal("Results:", err)
	}
	// We want one file match and one repository match
	wantMatchCount := 2
	if int(results.MatchCount()) != wantMatchCount {
		t.Fatalf("wrong results length. want=%d, have=%d\n", wantMatchCount, results.MatchCount())
	}

	for _, r := range results.Results() {
		switch r := r.(type) {
		case *FileMatchResolver:
			assertRepoResolverHydrated(ctx, t, r.Repository(), hydratedRepo)

		case *RepositoryResolver:
			assertRepoResolverHydrated(ctx, t, r, hydratedRepo)
		}
	}
}

func TestDedupSort(t *testing.T) {
	repos := make(types.Repos, 512)
	for i := range repos {
		repos[i] = &types.Repo{ID: api.RepoID(i % 256)}
	}

	rand.Shuffle(len(repos), func(i, j int) {
		repos[i], repos[j] = repos[j], repos[i]
	})

	dedupSort(&repos)

	if have, want := len(repos), 256; have != want {
		t.Fatalf("have %d unique repos, want: %d", have, want)
	}

	for i, r := range repos {
		if have, want := api.RepoID(i), r.ID; have != want {
			t.Errorf("%dth repo id = %d, want %d", i, have, want)
		}
	}
}

func TestCommitAndDiffSearchLimits(t *testing.T) {
	cases := []struct {
		name                 string
		resultTypes          []string
		numRepoRevs          int
		fields               map[string][]*searchquerytypes.Value
		wantResultTypes      []string
		wantAlertDescription string
	}{
		{
			name:                 "diff_search_warns_on_repos_greater_than_search_limit",
			resultTypes:          []string{"diff"},
			numRepoRevs:          51,
			wantResultTypes:      []string{}, // diff is removed from the resultTypes
			wantAlertDescription: `Diff search can currently only handle searching over 50 repositories at a time. Try using the "repo:" filter to narrow down which repositories to search, or using 'after:"1 week ago"'. Tracking issue: https://github.com/sourcegraph/sourcegraph/issues/6826`,
		},
		{
			name:                 "commit_search_warns_on_repos_greater_than_search_limit",
			resultTypes:          []string{"commit"},
			numRepoRevs:          51,
			wantResultTypes:      []string{}, // diff is removed from the resultTypes
			wantAlertDescription: `Commit search can currently only handle searching over 50 repositories at a time. Try using the "repo:" filter to narrow down which repositories to search, or using 'after:"1 week ago"'. Tracking issue: https://github.com/sourcegraph/sourcegraph/issues/6826`,
		},
		{
			name:                 "no_warning_when_commit_search_within_search_limit",
			resultTypes:          []string{"commit"},
			numRepoRevs:          50,
			wantResultTypes:      []string{"commit"}, // commit is preserved in resultTypes
			wantAlertDescription: "",
		},
		{
			name:                 "no_search_limit_on_queries_including_after_filter",
			fields:               map[string][]*searchquerytypes.Value{"after": nil},
			resultTypes:          []string{"file"},
			numRepoRevs:          200,
			wantResultTypes:      []string{"file"},
			wantAlertDescription: "",
		},
		{
			name:                 "no_search_limit_on_queries_including_before_filter",
			fields:               map[string][]*searchquerytypes.Value{"before": nil},
			resultTypes:          []string{"file"},
			numRepoRevs:          200,
			wantResultTypes:      []string{"file"},
			wantAlertDescription: "",
		},
		{
			name:                 "no_search_limit_on_repos_for_file_search",
			resultTypes:          []string{"file"},
			numRepoRevs:          200,
			wantResultTypes:      []string{"file"},
			wantAlertDescription: "",
		},
		{
			name:                 "multiple_result_type_search_is_unaffected",
			resultTypes:          []string{"file", "commit"},
			numRepoRevs:          200,
			wantResultTypes:      []string{"file", "commit"},
			wantAlertDescription: "",
		},
	}

	for _, test := range cases {
		repoRevs := make([]*search.RepositoryRevisions, test.numRepoRevs)
		for i := range repoRevs {
			repoRevs[i] = &search.RepositoryRevisions{
				Repo: &types.Repo{ID: api.RepoID(i)},
			}
		}

		haveResultTypes, alert := alertOnSearchLimit(test.resultTypes, &search.TextParameters{
			Repos: repoRevs,
			Query: &query.OrdinaryQuery{Query: &query.Query{Fields: test.fields}},
		})

		haveAlertDescription := ""
		if alert != nil {
			haveAlertDescription = *alert.Description()
		}

		if haveAlertDescription != test.wantAlertDescription {
			t.Fatalf("test %s, have alert %q, want: %q", test.name, haveAlertDescription, test.wantAlertDescription)
		}
		if !reflect.DeepEqual(haveResultTypes, test.wantResultTypes) {
			haveResultType := "is empty"
			wantResultType := "is empty"
			if len(haveResultTypes) > 0 {
				haveResultType = haveResultTypes[0]
			}
			if len(test.wantResultTypes) > 0 {
				wantResultType = test.wantResultTypes[0]
			}
			t.Fatalf("test %s, have result type: %q, want result type: %q", test.name, haveResultType, wantResultType)
		}
	}
}

func Test_SearchResultsResolver_ApproximateResultCount(t *testing.T) {
	type fields struct {
		results             []SearchResultResolver
		searchResultsCommon searchResultsCommon
		alert               *searchAlert
		start               time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "empty",
			fields: fields{},
			want:   "0",
		},

		{
			name: "file matches",
			fields: fields{
				results: []SearchResultResolver{&FileMatchResolver{}},
			},
			want: "1",
		},

		{
			name: "file matches limit hit",
			fields: fields{
				results:             []SearchResultResolver{&FileMatchResolver{}},
				searchResultsCommon: searchResultsCommon{limitHit: true},
			},
			want: "1+",
		},

		{
			name: "symbol matches",
			fields: fields{
				results: []SearchResultResolver{
					&FileMatchResolver{
						symbols: []*searchSymbolResult{
							// 1
							{},
							// 2
							{},
						},
					},
				},
			},
			want: "2",
		},

		{
			name: "symbol matches limit hit",
			fields: fields{
				results: []SearchResultResolver{
					&FileMatchResolver{
						symbols: []*searchSymbolResult{
							// 1
							{},
							// 2
							{},
						},
					},
				},
				searchResultsCommon: searchResultsCommon{limitHit: true},
			},
			want: "2+",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := &SearchResultsResolver{
				SearchResults:       tt.fields.results,
				searchResultsCommon: tt.fields.searchResultsCommon,
				alert:               tt.fields.alert,
				start:               tt.fields.start,
			}
			if got := sr.ApproximateResultCount(); got != tt.want {
				t.Errorf("searchResultsResolver.ApproximateResultCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchResolver_evaluateWarning(t *testing.T) {
	q, _ := query.ProcessAndOr("file:foo or file:bar", query.ParserOptions{SearchType: query.SearchTypeRegex, Globbing: false})
	wantPrefix := "I'm having trouble understanding that query."
	andOrQuery, _ := q.(*query.AndOrQuery)
	got, _ := (&searchResolver{}).evaluate(context.Background(), andOrQuery.Query)
	t.Run("warn for unsupported and/or query", func(t *testing.T) {
		if !strings.HasPrefix(got.alert.description, wantPrefix) {
			t.Fatalf("got alert description %s, want %s", got.alert.description, wantPrefix)
		}
	})

	_, err := query.ProcessAndOr("file:foo or or or", query.ParserOptions{SearchType: query.SearchTypeRegex, Globbing: false})
	gotAlert := alertForQuery("", err)
	t.Run("warn for unsupported ambiguous and/or query", func(t *testing.T) {
		if !strings.HasPrefix(gotAlert.description, wantPrefix) {
			t.Fatalf("got alert description %s, want %s", got.alert.description, wantPrefix)
		}
	})
}
