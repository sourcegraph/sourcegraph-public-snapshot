package graphqlbackend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/zoekt"
	zoektrpc "github.com/google/zoekt/rpc"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
	searchbackend "github.com/sourcegraph/sourcegraph/pkg/search/backend"
)

func TestSearchResults(t *testing.T) {
	limitOffset := &db.LimitOffset{Limit: maxReposToSearch() + 1}

	getResults := func(t *testing.T, query string) []string {
		r, err := (&schemaResolver{}).Search(&struct{ Query string }{Query: query})
		if err != nil {
			t.Fatal("Search:", err)
		}
		results, err := r.Results(context.Background())
		if err != nil {
			t.Fatal("Results:", err)
		}
		resultDescriptions := make([]string, len(results.results))
		for i, result := range results.results {
			// NOTE: Only supports one match per line. If we need to test other cases,
			// just remove that assumption in the following line of code.
			switch m := result.(type) {
			case *repositoryResolver:
				resultDescriptions[i] = fmt.Sprintf("repo:%s", m.repo.Name)
			case *fileMatchResolver:
				resultDescriptions[i] = fmt.Sprintf("%s:%d", m.JPath, m.JLineMatches[0].JLineNumber)
			default:
				t.Fatal("unexpected result type", result)
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
		db.Mocks.Repos.MinimalList = func(_ context.Context, op db.ReposListOptions) ([]*db.MinimalRepo, error) {
			calledReposList = true
			if want := (db.ReposListOptions{Enabled: true, IncludePatterns: []string{"r", "p"}, LimitOffset: limitOffset}); !reflect.DeepEqual(op, want) {
				t.Fatalf("got %+v, want %+v", op, want)
			}
			return []*db.MinimalRepo{{ID: 1, Name: "repo"}}, nil
		}
		db.Mocks.Repos.MockGetByName(t, "repo", 1)
		db.Mocks.Repos.MockGet(t, 1)

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
		db.Mocks.Repos.MinimalList = func(_ context.Context, op db.ReposListOptions) ([]*db.MinimalRepo, error) {
			calledReposList = true
			if want := (db.ReposListOptions{Enabled: true, LimitOffset: limitOffset}); !reflect.DeepEqual(op, want) {
				t.Fatalf("got %+v, want %+v", op, want)
			}
			return []*db.MinimalRepo{{ID: 1, Name: "repo"}}, nil
		}
		defer func() { db.Mocks = db.MockStores{} }()
		db.Mocks.Repos.MockGetByName(t, "repo", 1)
		db.Mocks.Repos.MockGet(t, 1)

		calledSearchRepositories := false
		mockSearchRepositories = func(args *search.Args) ([]searchResultResolver, *searchResultsCommon, error) {
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
				{
					uri:          "git://repo?rev#dir/file",
					JPath:        "dir/file",
					JLineMatches: []*lineMatch{{JLineNumber: 123}},
					repo:         &db.MinimalRepo{ID: 1},
				},
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

func BenchmarkSearchResults(b *testing.B) {
	minimalRepos, _, zoektRepos := generateRepos(5000)
	zoektFileMatches := generateZoektMatches(50)

	z := &searchbackend.Zoekt{
		Client: &fakeSearcher{
			repos:  &zoekt.RepoList{Repos: zoektRepos},
			result: &zoekt.SearchResult{Files: zoektFileMatches},
		},
		DisableCache: true,
	}

	ctx := context.Background()
	db.Mocks.Repos.MinimalList = func(_ context.Context, op db.ReposListOptions) ([]*db.MinimalRepo, error) {
		return minimalRepos, nil
	}
	defer func() { db.Mocks = db.MockStores{} }()

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		q, err := query.ParseAndCheck(`print index:only count:350`)
		if err != nil {
			b.Fatal(err)
		}
		resolver := &searchResolver{query: q, zoekt: z}
		results, err := resolver.Results(ctx)
		if err != nil {
			b.Fatal("Results:", err)
		}
		if int(results.MatchCount()) != len(zoektFileMatches) {
			b.Fatalf("wrong results length. want=%d, have=%d\n", len(zoektFileMatches), results.MatchCount())
		}
	}
}

func BenchmarkIntegrationSearchResults(b *testing.B) {
	ctx := dbtesting.TestContext(b)

	_, repos, zoektRepos := generateRepos(5000)
	zoektFileMatches := generateZoektMatches(50)

	zoektClient, cleanup := zoektRPC(&fakeSearcher{
		repos:  &zoekt.RepoList{Repos: zoektRepos},
		result: &zoekt.SearchResult{Files: zoektFileMatches},
	})
	defer cleanup()
	z := &searchbackend.Zoekt{
		Client:       zoektClient,
		DisableCache: true,
	}

	for _, r := range repos {
		err := db.Repos.Upsert(ctx, api.InsertRepoOp{
			Name:         r.Name,
			Description:  r.Description,
			Fork:         r.Fork,
			Archived:     false,
			Enabled:      true,
			ExternalRepo: r.ExternalRepo,
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		q, err := query.ParseAndCheck(`print index:only count:350`)
		if err != nil {
			b.Fatal(err)
		}
		resolver := &searchResolver{query: q, zoekt: z}
		results, err := resolver.Results(ctx)
		if err != nil {
			b.Fatal("Results:", err)
		}
		if int(results.MatchCount()) != len(zoektFileMatches) {
			b.Fatalf("wrong results length. want=%d, have=%d\n", len(zoektFileMatches), results.MatchCount())
		}
	}
}

func generateRepos(count int) ([]*db.MinimalRepo, []*types.Repo, []*zoekt.RepoListEntry) {
	var minimalRepos []*db.MinimalRepo
	var repos []*types.Repo
	var zoektRepos []*zoekt.RepoListEntry

	for i := 1; i <= count; i++ {
		name := fmt.Sprintf("repo-%d", i)

		minimalRepo := &db.MinimalRepo{
			ID:   api.RepoID(i),
			Name: api.RepoName(name),
			ExternalRepo: api.ExternalRepoSpec{
				ID:          name,
				ServiceType: "github",
				ServiceID:   "https://github.com",
			},
		}

		minimalRepos = append(minimalRepos, minimalRepo)

		repos = append(repos, &types.Repo{
			ID:           minimalRepo.ID,
			Name:         minimalRepo.Name,
			ExternalRepo: &minimalRepo.ExternalRepo,
			URI:          fmt.Sprintf("https://github.com/foobar/%s", minimalRepo.Name),
			Description:  "this repositoriy contains a side project that I haven't maintained in 2 years",
			Language:     "v-language",
		})

		zoektRepos = append(zoektRepos, &zoekt.RepoListEntry{
			Repository: zoekt.Repository{
				Name:     name,
				Branches: []zoekt.RepositoryBranch{{Name: "HEAD", Version: "deadbeef"}},
			},
		})
	}
	return minimalRepos, repos, zoektRepos
}

func generateZoektMatches(count int) []zoekt.FileMatch {
	var zoektFileMatches []zoekt.FileMatch
	for i := 1; i <= count; i++ {
		repoName := fmt.Sprintf("repo-%d", i)
		fileName := fmt.Sprintf("foobar-%d.go", i)

		zoektFileMatches = append(zoektFileMatches, zoekt.FileMatch{
			Score:      5.0,
			FileName:   fileName,
			Repository: repoName, // Important: this needs to match a name in `repos`
			Branches:   []string{"master"},
			LineMatches: []zoekt.LineMatch{
				{
					Line: nil,
				},
			},
			Checksum: []byte{0, 1, 2},
		})
	}
	return zoektFileMatches
}

// zoektRPC starts zoekts rpc interface and returns a client to
// searcher. Useful for capturing CPU/memory usage when benchmarking the zoekt
// client.
func zoektRPC(s zoekt.Searcher) (zoekt.Searcher, func()) {
	mux := http.NewServeMux()
	mux.Handle(zoektrpc.DefaultRPCPath, zoektrpc.Server(s))
	ts := httptest.NewServer(mux)
	cl := zoektrpc.Client(strings.TrimPrefix(ts.URL, "http://"))
	return cl, func() {
		cl.Close()
		ts.Close()
	}
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
	repo := &db.MinimalRepo{
		Name: "testRepo",
	}

	repoMatch := &repositoryResolver{
		repo: repo,
	}

	fileMatch := &fileMatchResolver{
		JPath: "/testFile.md",
		repo:  repo,
	}

	tsFileMatch := &fileMatchResolver{
		JPath: "/testFile.ts",
		repo:  repo,
	}

	tsxFileMatch := &fileMatchResolver{
		JPath: "/testFile.tsx",
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
		searchResults             []searchResultResolver
		expectedDynamicFilterStrs map[string]struct{}
	}

	tests := []testCase{

		{
			descr:         "single repo match",
			searchResults: []searchResultResolver{repoMatch},
			expectedDynamicFilterStrs: map[string]struct{}{
				`repo:^testRepo$`: {},
				`case:yes`:        {},
			},
		},

		{
			descr:         "single file match without revision in query",
			searchResults: []searchResultResolver{fileMatch},
			expectedDynamicFilterStrs: map[string]struct{}{
				`repo:^testRepo$`: {},
				`lang:markdown`:   {},
				`case:yes`:        {},
			},
		},

		{
			descr:         "single file match with specified revision",
			searchResults: []searchResultResolver{fileMatchRev},
			expectedDynamicFilterStrs: map[string]struct{}{
				`repo:^testRepo$@develop`: {},
				`lang:markdown`:           {},
				`case:yes`:                {},
			},
		},
		{
			descr:         "file match from a language with two file extensions, using first extension",
			searchResults: []searchResultResolver{tsFileMatch},
			expectedDynamicFilterStrs: map[string]struct{}{
				`repo:^testRepo$`: {},
				`lang:typescript`: {},
				`case:yes`:        {},
			},
		},
		{
			descr:         "file match from a language with two file extensions, using second extension",
			searchResults: []searchResultResolver{tsxFileMatch},
			expectedDynamicFilterStrs: map[string]struct{}{
				`repo:^testRepo$`: {},
				`lang:typescript`: {},
				`case:yes`:        {},
			},
		},

		// If there are no search results, no filters should be displayed.
		{
			descr:                     "no results",
			searchResults:             []searchResultResolver{},
			expectedDynamicFilterStrs: map[string]struct{}{},
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
		a       searchResultResolver
		b       searchResultResolver
		aIsLess bool
	}

	tests := []testCase{{
		// Different repo matches
		a: &repositoryResolver{
			repo: &db.MinimalRepo{
				Name: api.RepoName("a"),
			},
		},
		b: &repositoryResolver{
			repo: &db.MinimalRepo{
				Name: api.RepoName("b"),
			},
		},
		aIsLess: true,
	}, {
		// Repo match vs file match in same repo
		a: &fileMatchResolver{
			repo: &db.MinimalRepo{
				Name: api.RepoName("a"),
			},
			JPath: "a",
		},
		b: &repositoryResolver{
			repo: &db.MinimalRepo{
				Name: api.RepoName("a"),
			},
		},
		aIsLess: false,
	}, {
		// Same repo, different files
		a: &fileMatchResolver{
			repo: &db.MinimalRepo{
				Name: api.RepoName("a"),
			},
			JPath: "a",
		},
		b: &fileMatchResolver{
			repo: &db.MinimalRepo{
				Name: api.RepoName("a"),
			},
			JPath: "b",
		},
		aIsLess: true,
	}, {
		// different repo, same file name
		a: &fileMatchResolver{
			repo: &db.MinimalRepo{
				Name: api.RepoName("a"),
			},
			JPath: "a",
		},
		b: &fileMatchResolver{
			repo: &db.MinimalRepo{
				Name: api.RepoName("b"),
			},
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

func Test_longer(t *testing.T) {
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

func Test_roundStr(t *testing.T) {
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

	minimalRepo := &db.MinimalRepo{
		ID:   api.RepoID(id),
		Name: api.RepoName(repoName),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          repoName,
			ServiceType: "github",
			ServiceID:   "https://github.com",
		},
	}

	hydratedRepo := &types.Repo{
		ID:           minimalRepo.ID,
		ExternalRepo: &(minimalRepo.ExternalRepo),
		Name:         minimalRepo.Name,
		URI:          fmt.Sprintf("github.com/my-org/%s", minimalRepo.Name),
		Description:  "This is a description of a repository",
		Language:     "monkey",
		Fork:         false,
	}

	db.Mocks.Repos.Get = func(ctx context.Context, id api.RepoID) (*types.Repo, error) {
		return hydratedRepo, nil
	}

	db.Mocks.Repos.MinimalList = func(_ context.Context, op db.ReposListOptions) ([]*db.MinimalRepo, error) {
		return []*db.MinimalRepo{minimalRepo}, nil
	}

	defer func() { db.Mocks = db.MockStores{} }()

	zoektRepo := &zoekt.RepoListEntry{
		Repository: zoekt.Repository{
			Name:     string(minimalRepo.Name),
			Branches: []zoekt.RepositoryBranch{{Name: "HEAD", Version: "deadbeef"}},
		},
	}

	zoektFileMatches := []zoekt.FileMatch{{
		Score:      5.0,
		FileName:   fileName,
		Repository: string(minimalRepo.Name), // Important: this needs to match a name in `repos`
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
			repos:  &zoekt.RepoList{Repos: []*zoekt.RepoListEntry{zoektRepo}},
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
		case *fileMatchResolver:
			assertRepoResolverHydrated(ctx, t, r.Repository(), hydratedRepo)

		case *repositoryResolver:
			assertRepoResolverHydrated(ctx, t, r, hydratedRepo)
		}
	}
}
