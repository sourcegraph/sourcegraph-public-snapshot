package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestQueryToZoektQuery(t *testing.T) {
	cases := []struct {
		Name    string
		Pattern *search.TextPatternInfo
		Query   string
	}{
		{
			Name: "substr",
			Pattern: &search.TextPatternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              false,
				Pattern:                      "foo",
				IncludePatterns:              nil,
				ExcludePattern:               "",
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: false,
			},
			Query: "foo case:no",
		},
		{
			Name: "regex",
			Pattern: &search.TextPatternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              false,
				Pattern:                      "(foo).*?(bar)",
				IncludePatterns:              nil,
				ExcludePattern:               "",
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: false,
			},
			Query: "(foo).*?(bar) case:no",
		},
		{
			Name: "path",
			Pattern: &search.TextPatternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              false,
				Pattern:                      "foo",
				IncludePatterns:              []string{`\.go$`, `\.yaml$`},
				ExcludePattern:               `\bvendor\b`,
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: false,
			},
			Query: `foo case:no f:\.go$ f:\.yaml$ -f:\bvendor\b`,
		},
		{
			Name: "case",
			Pattern: &search.TextPatternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              true,
				Pattern:                      "foo",
				IncludePatterns:              []string{`\.go$`, `yaml`},
				ExcludePattern:               "",
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: true,
			},
			Query: `foo case:yes f:\.go$ f:yaml`,
		},
		{
			Name: "casepath",
			Pattern: &search.TextPatternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              true,
				Pattern:                      "foo",
				IncludePatterns:              []string{`\.go$`, `\.yaml$`},
				ExcludePattern:               `\bvendor\b`,
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: true,
			},
			Query: `foo case:yes f:\.go$ f:\.yaml$ -f:\bvendor\b`,
		},
		{
			Name: "path matches only",
			Pattern: &search.TextPatternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              false,
				Pattern:                      "test",
				IncludePatterns:              []string{},
				ExcludePattern:               ``,
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: true,
				PatternMatchesContent:        false,
				PatternMatchesPath:           true,
			},
			Query: `f:test`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			q, err := zoektquery.Parse(tt.Query)
			if err != nil {
				t.Fatalf("failed to parse %q: %v", tt.Query, err)
			}
			got, err := queryToZoektQuery(tt.Pattern, false)
			if err != nil {
				t.Fatal("queryToZoektQuery failed:", err)
			}
			if !queryEqual(got, q) {
				t.Fatalf("mismatched queries\ngot  %s\nwant %s", got.String(), q.String())
			}
		})
	}
}

func TestStructuralPatToZoektQuery(t *testing.T) {
	cases := []struct {
		Name     string
		Pattern  string
		Function func(string) (zoektquery.Q, error)
		Want     string
	}{
		{
			Name:     "Just a hole",
			Pattern:  ":[1]",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()((?s:.))*?()")`,
		},
		{
			Name:     "Adjacent holes",
			Pattern:  ":[1]:[2]:[3]",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()((?s:.))*?()((?s:.))*?()((?s:.))*?()")`,
		},
		{
			Name:     "Substring between holes",
			Pattern:  ":[1] substring :[2]",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()((?s:.))*?([\\t-\\n\\f-\\r ]+substring[\\t-\\n\\f-\\r ]+)((?s:.))*?()")`,
		},
		{
			Name:     "Substring before and after different hole kinds",
			Pattern:  "prefix :[[1]] :[2.] suffix",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(prefix[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+suffix)")`,
		},
		{
			Name:     "Substrings covering all hole kinds.",
			Pattern:  `1. :[1] 2. :[[2]] 3. :[3.] 4. :[4\n] 5. :[ ] 6. :[ 6] done.`,
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(1\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+2\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+3\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+4\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+5\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+6\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+done\\.)")`,
		},
		{
			Name:     "Substrings across multiple lines.",
			Pattern:  ``,
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()")`,
		},
		{
			Name:     "Allow alphanumeric identifiers in holes",
			Pattern:  "sub :[alphanum_ident_123] string",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(sub[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+string)")`,
		},

		{
			Name:     "Whitespace separated holes",
			Pattern:  ":[1] :[2]",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()((?s:.))*?([\\t-\\n\\f-\\r ]+)((?s:.))*?()")`,
		},
		{
			Name:     "Expect newline separated pattern",
			Pattern:  "ParseInt(:[stuff], :[x]) if err ",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(ParseInt\\()((?s:.))*?(,[\\t-\\n\\f-\\r ]+)((?s:.))*?(\\)[\\t-\\n\\f-\\r ]+if[\\t-\\n\\f-\\r ]+err[\\t-\\n\\f-\\r ]+)")`,
		},
		{
			Name: "Contiguous whitespace is replaced by regex",
			Pattern: `ParseInt(:[stuff],    :[x])
             if err `,
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(ParseInt\\()((?s:.))*?(,[\\t-\\n\\f-\\r ]+)((?s:.))*?(\\)[\\t-\\n\\f-\\r ]+if[\\t-\\n\\f-\\r ]+err[\\t-\\n\\f-\\r ]+)")`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			got, _ := tt.Function(tt.Pattern)
			if got.String() != tt.Want {
				t.Fatalf("mismatched queries\ngot  %s\nwant %s", got.String(), tt.Want)
			}
		})
	}
}

func queryEqual(a, b zoektquery.Q) bool {
	sortChildren := func(q zoektquery.Q) zoektquery.Q {
		switch s := q.(type) {
		case *zoektquery.And:
			sort.Slice(s.Children, func(i, j int) bool {
				return s.Children[i].String() < s.Children[j].String()
			})
		case *zoektquery.Or:
			sort.Slice(s.Children, func(i, j int) bool {
				return s.Children[i].String() < s.Children[j].String()
			})
		}
		return q
	}
	return zoektquery.Map(a, sortChildren).String() == zoektquery.Map(b, sortChildren).String()
}

func TestQueryToZoektFileOnlyQueries(t *testing.T) {
	cases := []struct {
		Name    string
		Pattern *search.TextPatternInfo
		Query   []string
		// This should be the same value passed in to either FilePatternsReposMustInclude or FilePatternsReposMustExclude
		ListOfFilePaths []string
	}{
		{
			Name: "single repohasfile filter",
			Pattern: &search.TextPatternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              false,
				Pattern:                      "foo",
				IncludePatterns:              nil,
				ExcludePattern:               "",
				FilePatternsReposMustInclude: []string{"test.md"},
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: false,
			},
			Query:           []string{`f:"test.md"`},
			ListOfFilePaths: []string{"test.md"},
		},
		{
			Name: "multiple repohasfile filters",
			Pattern: &search.TextPatternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              false,
				Pattern:                      "foo",
				IncludePatterns:              nil,
				ExcludePattern:               "",
				FilePatternsReposMustInclude: []string{"t", "d"},
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: false,
			},
			Query:           []string{`f:"t"`, `f:"d"`},
			ListOfFilePaths: []string{"t", "d"},
		},
		{
			Name: "single negated repohasfile filter",
			Pattern: &search.TextPatternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              false,
				Pattern:                      "foo",
				IncludePatterns:              nil,
				ExcludePattern:               "",
				FilePatternsReposMustExclude: []string{"test.md"},
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: false,
			},
			Query:           []string{`f:"test.md"`},
			ListOfFilePaths: []string{"test.md"},
		},
		{
			Name: "multiple negated repohasfile filter",
			Pattern: &search.TextPatternInfo{
				IsRegExp:                     true,
				IsCaseSensitive:              false,
				Pattern:                      "foo",
				IncludePatterns:              nil,
				ExcludePattern:               "",
				FilePatternsReposMustExclude: []string{"t", "d"},
				PathPatternsAreRegExps:       true,
				PathPatternsAreCaseSensitive: false,
			},
			Query:           []string{`f:"t"`, `f:"d"`},
			ListOfFilePaths: []string{"t", "d"},
		},
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			queries := []zoektquery.Q{}
			for _, query := range tt.Query {
				q, err := zoektquery.Parse(query)
				if err != nil {
					t.Fatalf("failed to parse %q: %v", tt.Query, err)
				}
				queries = append(queries, q)
			}

			got, err := queryToZoektFileOnlyQueries(tt.Pattern, tt.ListOfFilePaths)
			if err != nil {
				t.Fatal("queryToZoektQuery failed:", err)
			}
			for i, gotQuery := range got {
				if !queryEqual(gotQuery, queries[i]) {
					t.Fatalf("mismatched queries\ngot  %s\nwant %s", gotQuery.String(), queries[i].String())
				}
			}
		})
	}
}

func TestSearchFilesInRepos(t *testing.T) {
	mockSearchFilesInRepo = func(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration) (matches []*FileMatchResolver, limitHit bool, err error) {
		repoName := repo.Name
		switch repoName {
		case "foo/one":
			return []*FileMatchResolver{
				{
					uri: "git://" + string(repoName) + "?" + rev + "#" + "main.go",
				},
			}, false, nil
		case "foo/two":
			return []*FileMatchResolver{
				{
					uri: "git://" + string(repoName) + "?" + rev + "#" + "main.go",
				},
			}, false, nil
		case "foo/empty":
			return nil, false, nil
		case "foo/cloning":
			return nil, false, &vcs.RepoNotExistError{Repo: repoName, CloneInProgress: true}
		case "foo/missing":
			return nil, false, &vcs.RepoNotExistError{Repo: repoName}
		case "foo/missing-db":
			return nil, false, &errcode.Mock{Message: "repo not found: foo/missing-db", IsNotFound: true}
		case "foo/timedout":
			return nil, false, context.DeadlineExceeded
		case "foo/no-rev":
			return nil, false, &gitserver.RevisionNotFoundError{Repo: repoName, Spec: "missing"}
		default:
			return nil, false, errors.New("Unexpected repo")
		}
	}
	defer func() { mockSearchFilesInRepo = nil }()

	zoekt := &searchbackend.Zoekt{Client: &fakeSearcher{repos: &zoekt.RepoList{}}}

	q, err := query.ParseAndCheck("foo")
	if err != nil {
		t.Fatal(err)
	}
	args := &search.TextParameters{
		PatternInfo: &search.TextPatternInfo{
			FileMatchLimit: defaultMaxSearchResults,
			Pattern:        "foo",
		},
		Repos:        makeRepositoryRevisions("foo/one", "foo/two", "foo/empty", "foo/cloning", "foo/missing", "foo/missing-db", "foo/timedout", "foo/no-rev"),
		Query:        q,
		Zoekt:        zoekt,
		SearcherURLs: endpoint.Static("test"),
	}
	results, common, err := searchFilesInRepos(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected two results, got %d", len(results))
	}
	if v := toRepoNames(common.cloning); !reflect.DeepEqual(v, []api.RepoName{"foo/cloning"}) {
		t.Errorf("unexpected cloning: %v", v)
	}
	sort.Slice(common.missing, func(i, j int) bool { return common.missing[i].Name < common.missing[j].Name }) // to make deterministic
	if v := toRepoNames(common.missing); !reflect.DeepEqual(v, []api.RepoName{"foo/missing", "foo/missing-db"}) {
		t.Errorf("unexpected missing: %v", v)
	}
	if v := toRepoNames(common.timedout); !reflect.DeepEqual(v, []api.RepoName{"foo/timedout"}) {
		t.Errorf("unexpected timedout: %v", v)
	}

	// If we specify a rev and it isn't found, we fail the whole search since
	// that should be checked earlier.
	args = &search.TextParameters{
		PatternInfo: &search.TextPatternInfo{
			FileMatchLimit: defaultMaxSearchResults,
			Pattern:        "foo",
		},
		Repos:        makeRepositoryRevisions("foo/no-rev@dev"),
		Query:        q,
		Zoekt:        zoekt,
		SearcherURLs: endpoint.Static("test"),
	}

	_, _, err = searchFilesInRepos(context.Background(), args)
	if !gitserver.IsRevisionNotFound(errors.Cause(err)) {
		t.Fatalf("searching non-existent rev expected to fail with RevisionNotFoundError got: %v", err)
	}
}

func TestSearchFilesInRepos_multipleRevsPerRepo(t *testing.T) {
	mockSearchFilesInRepo = func(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration) (matches []*FileMatchResolver, limitHit bool, err error) {
		repoName := repo.Name
		switch repoName {
		case "foo":
			return []*FileMatchResolver{
				{
					uri: "git://" + string(repoName) + "?" + rev + "#" + "main.go",
				},
			}, false, nil
		default:
			panic("unexpected repo")
		}
	}
	defer func() { mockSearchFilesInRepo = nil }()

	trueVal := true
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{SearchMultipleRevisionsPerRepository: &trueVal},
	}})
	defer conf.Mock(nil)

	zoekt := &searchbackend.Zoekt{Client: &fakeSearcher{repos: &zoekt.RepoList{}}}

	q, err := query.ParseAndCheck("foo")
	if err != nil {
		t.Fatal(err)
	}
	args := &search.TextParameters{
		PatternInfo: &search.TextPatternInfo{
			FileMatchLimit: defaultMaxSearchResults,
			Pattern:        "foo",
		},
		Repos:        makeRepositoryRevisions("foo@master:mybranch:*refs/heads/"),
		Query:        q,
		Zoekt:        zoekt,
		SearcherURLs: endpoint.Static("test"),
	}
	args.Repos[0].ListRefs = func(context.Context, gitserver.Repo) ([]git.Ref, error) {
		return []git.Ref{{Name: "refs/heads/branch3"}, {Name: "refs/heads/branch4"}}, nil
	}
	results, _, err := searchFilesInRepos(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}

	resultURIs := make([]string, len(results))
	for i, result := range results {
		resultURIs[i] = result.uri
	}
	sort.Strings(resultURIs)

	wantResultURIs := []string{
		"git://foo?branch3#main.go",
		"git://foo?branch4#main.go",
		"git://foo?master#main.go",
		"git://foo?mybranch#main.go",
	}
	if !reflect.DeepEqual(resultURIs, wantResultURIs) {
		t.Errorf("got %v, want %v", resultURIs, wantResultURIs)
	}
}

func TestRepoShouldBeSearched(t *testing.T) {
	mockTextSearch = func(ctx context.Context, repo gitserver.Repo, commit api.CommitID, p *search.TextPatternInfo, fetchTimeout time.Duration) (matches []*FileMatchResolver, limitHit bool, err error) {
		repoName := repo.Name
		switch repoName {
		case "foo/one":
			return []*FileMatchResolver{
				{
					uri: "git://" + string(repoName) + "?1a2b3c#" + "main.go",
				},
			}, false, nil
		case "foo/no-filematch":
			return []*FileMatchResolver{}, false, nil
		default:
			return nil, false, errors.New("Unexpected repo")
		}
	}
	defer func() { mockTextSearch = nil }()
	info := &search.TextPatternInfo{
		FileMatchLimit:               defaultMaxSearchResults,
		Pattern:                      "foo",
		FilePatternsReposMustInclude: []string{"main"},
	}

	shouldBeSearched, err := repoShouldBeSearched(context.Background(), nil, info, gitserver.Repo{Name: "foo/one", URL: "http://example.com/foo/one"}, "1a2b3c", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !shouldBeSearched {
		t.Errorf("expected repo to be searched, got shouldn't be searched")
	}

	shouldBeSearched, err = repoShouldBeSearched(context.Background(), nil, info, gitserver.Repo{Name: "foo/no-filematch", URL: "http://example.com/foo/no-filematch"}, "1a2b3c", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if shouldBeSearched {
		t.Errorf("expected repo to not be searched, got should be searched")
	}
}

func makeRepositoryRevisions(repos ...string) []*search.RepositoryRevisions {
	r := make([]*search.RepositoryRevisions, len(repos))
	for i, repospec := range repos {
		repoName, revs := search.ParseRepositoryRevisions(repospec)
		if len(revs) == 0 {
			// treat empty list as preferring master
			revs = []search.RevisionSpecifier{{RevSpec: ""}}
		}
		r[i] = &search.RepositoryRevisions{Repo: &types.Repo{Name: repoName}, Revs: revs}
	}
	return r
}

// fakeSearcher is a zoekt.Searcher that returns a predefined search result.
type fakeSearcher struct {
	result *zoekt.SearchResult

	repos *zoekt.RepoList

	// Default all unimplemented zoekt.Searcher methods to panic.
	zoekt.Searcher
}

func (ss *fakeSearcher) Search(ctx context.Context, q zoektquery.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	return ss.result, nil
}

func (ss *fakeSearcher) List(ctx context.Context, q zoektquery.Q) (*zoekt.RepoList, error) {
	return ss.repos, nil
}

func (ss *fakeSearcher) String() string {
	return fmt.Sprintf("fakeSearcher(result = %v, repos = %v)", ss.result, ss.repos)
}

type errorSearcher struct {
	err error

	// Default all unimplemented zoekt.Searcher methods to panic.
	zoekt.Searcher
}

func (es *errorSearcher) Search(ctx context.Context, q zoektquery.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	return nil, es.err
}

func Test_zoektSearchHEAD(t *testing.T) {
	zeroTimeoutCtx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	type args struct {
		ctx             context.Context
		query           *search.TextPatternInfo
		repos           []*search.RepositoryRevisions
		useFullDeadline bool
		searcher        zoekt.Searcher
		since           func(time.Time) time.Duration
	}

	rr := &search.RepositoryRevisions{Repo: &types.Repo{}}
	rr.SetIndexedHEADCommit("abc")
	singleRepositoryRevisions := []*search.RepositoryRevisions{rr}

	tests := []struct {
		name              string
		args              args
		wantFm            []*FileMatchResolver
		wantLimitHit      bool
		wantReposLimitHit map[string]struct{}
		wantErr           bool
	}{
		{
			name: "returns no error if search completed with no matches before timeout",
			args: args{
				ctx:             context.Background(),
				query:           &search.TextPatternInfo{PathPatternsAreRegExps: true},
				repos:           singleRepositoryRevisions,
				useFullDeadline: false,
				searcher:        &fakeSearcher{result: &zoekt.SearchResult{}},
				since:           func(time.Time) time.Duration { return time.Second - time.Millisecond },
			},
			wantFm:            nil,
			wantLimitHit:      false,
			wantReposLimitHit: nil,
			wantErr:           false,
		},
		{
			name: "returns error if max wall time is exceeded but no matches have been found yet",
			args: args{
				ctx:             context.Background(),
				query:           &search.TextPatternInfo{PathPatternsAreRegExps: true},
				repos:           singleRepositoryRevisions,
				useFullDeadline: false,
				searcher:        &fakeSearcher{result: &zoekt.SearchResult{}},
				since:           func(time.Time) time.Duration { return 4 * time.Second },
			},
			wantFm:            nil,
			wantLimitHit:      false,
			wantReposLimitHit: nil,
			wantErr:           true,
		},
		{
			name: "returns error if context timeout already passed",
			args: args{
				ctx:             zeroTimeoutCtx,
				query:           &search.TextPatternInfo{PathPatternsAreRegExps: true},
				repos:           singleRepositoryRevisions,
				useFullDeadline: true,
				searcher:        &fakeSearcher{result: &zoekt.SearchResult{}},
				since:           func(time.Time) time.Duration { return 0 },
			},
			wantFm:            nil,
			wantLimitHit:      false,
			wantReposLimitHit: nil,
			wantErr:           true,
		},
		{
			name: "returns error if searcher returns an error",
			args: args{
				ctx:             context.Background(),
				query:           &search.TextPatternInfo{PathPatternsAreRegExps: true},
				repos:           singleRepositoryRevisions,
				useFullDeadline: true,
				searcher:        &errorSearcher{err: errors.New("womp womp")},
				since:           func(time.Time) time.Duration { return 0 },
			},
			wantFm:            nil,
			wantLimitHit:      false,
			wantReposLimitHit: nil,
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := &search.TextParameters{
				PatternInfo:     tt.args.query,
				UseFullDeadline: tt.args.useFullDeadline,
				Zoekt:           &searchbackend.Zoekt{Client: tt.args.searcher},
			}
			gotFm, gotLimitHit, gotReposLimitHit, err := zoektSearchHEAD(tt.args.ctx, args, tt.args.repos, false, tt.args.since)
			if (err != nil) != tt.wantErr {
				t.Errorf("zoektSearchHEAD() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFm, tt.wantFm) {
				t.Errorf("zoektSearchHEAD() gotFm = %v, want %v", gotFm, tt.wantFm)
			}
			if gotLimitHit != tt.wantLimitHit {
				t.Errorf("zoektSearchHEAD() gotLimitHit = %v, want %v", gotLimitHit, tt.wantLimitHit)
			}
			if !reflect.DeepEqual(gotReposLimitHit, tt.wantReposLimitHit) {
				t.Errorf("zoektSearchHEAD() gotReposLimitHit = %v, want %v", gotReposLimitHit, tt.wantReposLimitHit)
			}
		})
	}
}

// repoURLsFakeSearcher fakes a searcher for use in
// createNewRepoSetWithRepoHasFileInputs. It only supports setting the
// RepoURLs field in search results, and will only evaluate search queries
// containing RepoSets and file path filters.
//
// It is a map from repo name to list of files.
type repoURLsFakeSearcher map[string][]string

func (repoPaths repoURLsFakeSearcher) Search(ctx context.Context, q zoektquery.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	matchedRepoURLs := map[string]string{}
	for repo, files := range repoPaths {
		// We only expect a subset of query atoms. So we can evaluate them
		// against our repo and files and tell if this repo should be in the
		// result set.
		errS := ""
		eval := zoektquery.Map(q, func(q zoektquery.Q) zoektquery.Q {
			switch r := q.(type) {
			case *zoektquery.RepoSet:
				return &zoektquery.Const{Value: r.Set[repo]}

			case *zoektquery.Substring:
				// Return true if any file name matches pattern
				if r.Content || !r.FileName {
					errS = "content substr"
					return q
				}

				match := func(v string) bool {
					return strings.Contains(v, r.Pattern)
				}
				if !r.CaseSensitive {
					pat := strings.ToLower(r.Pattern)
					match = func(v string) bool {
						return strings.Contains(strings.ToLower(v), pat)
					}
				}

				for _, f := range files {
					if match(f) {
						return &zoektquery.Const{Value: true}
					}
				}
				return &zoektquery.Const{Value: false}

			case *zoektquery.Regexp:
				// Return true if any file name matches regexp
				if r.Content || !r.FileName {
					errS = "content regexp"
					return q
				}

				prefix := ""
				if !r.CaseSensitive {
					prefix = "(?i)"
				}
				re := regexp.MustCompile(prefix + r.Regexp.String())

				for _, f := range files {
					if re.FindStringIndex(f) != nil {
						return &zoektquery.Const{Value: true}
					}
				}
				return &zoektquery.Const{Value: false}

			case *zoektquery.And:
				return q

			default:
				errS = "unexpected query atom: " + q.String()
				return q
			}
		})
		if errS != "" {
			return nil, errors.Errorf("unsupported query %s: %s", q.String(), errS)
		}
		eval = zoektquery.Simplify(eval)
		if eval.(*zoektquery.Const).Value {
			matchedRepoURLs[repo] = repo
		}
	}

	return &zoekt.SearchResult{RepoURLs: matchedRepoURLs}, nil
}

func (repoURLsFakeSearcher) List(ctx context.Context, q zoektquery.Q) (*zoekt.RepoList, error) {
	panic("unimplemented")
}

func (repoURLsFakeSearcher) String() string {
	panic("unimplemented")
}

func (repoURLsFakeSearcher) Close() {
	panic("unimplemented")
}

func Test_createNewRepoSetWithRepoHasFileInputs(t *testing.T) {
	searcher := repoURLsFakeSearcher{
		"github.com/test/1": []string{"1.md"},
		"github.com/test/2": []string{"2.md"},
	}
	allRepos := []string{"github.com/test/1", "github.com/test/2"}

	tests := []struct {
		name        string
		include     []string
		exclude     []string
		repoSet     []string
		wantRepoSet []string
	}{
		{
			name:        "all",
			include:     []string{"md"},
			repoSet:     allRepos,
			wantRepoSet: allRepos,
		},
		{
			name:    "none",
			include: []string{"foo"},
			repoSet: allRepos,
		},
		{
			name:        "one include",
			include:     []string{"1"},
			repoSet:     allRepos,
			wantRepoSet: []string{"github.com/test/1"},
		},
		{
			name:        "two include",
			include:     []string{"md", "2"},
			repoSet:     allRepos,
			wantRepoSet: []string{"github.com/test/2"},
		},
		{
			name:        "include exclude",
			include:     []string{"md"},
			exclude:     []string{"1"},
			repoSet:     allRepos,
			wantRepoSet: []string{"github.com/test/2"},
		},
		{
			name:        "exclude",
			exclude:     []string{"1"},
			repoSet:     allRepos,
			wantRepoSet: []string{"github.com/test/2"},
		},
		{
			name:    "exclude all",
			exclude: []string{"md"},
			repoSet: allRepos,
		},
		{
			name:        "exclude none",
			exclude:     []string{"foo"},
			repoSet:     allRepos,
			wantRepoSet: allRepos,
		},
		{
			name:        "subset of reposet",
			include:     []string{"md"},
			repoSet:     []string{"github.com/test/1"},
			wantRepoSet: []string{"github.com/test/1"},
		},
		{
			name:    "exclude subset of reposet",
			exclude: []string{"1"},
			repoSet: []string{"github.com/test/1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoSet := &zoektquery.RepoSet{Set: map[string]bool{}}
			for _, r := range tt.repoSet {
				repoSet.Set[r] = true
			}

			info := &search.TextPatternInfo{
				FilePatternsReposMustInclude: tt.include,
				FilePatternsReposMustExclude: tt.exclude,
				PathPatternsAreRegExps:       true,
			}

			gotRepoSet, err := createNewRepoSetWithRepoHasFileInputs(context.Background(), info, searcher, repoSet)
			if err != nil {
				t.Fatal(err)
			}

			var got []string
			for r := range gotRepoSet.Set {
				got = append(got, r)
			}

			sort.Strings(got)
			sort.Strings(tt.wantRepoSet)
			if !cmp.Equal(tt.wantRepoSet, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.wantRepoSet, got))
			}
		})
	}
}

func Test_zoektIndexedRepos(t *testing.T) {
	repos := makeRepositoryRevisions(
		"foo/indexed-one@",
		"foo/indexed-two@",
		"foo/indexed-three@",
		"foo/unindexed-one",
		"foo/unindexed-two",
		"foo/multi-rev@a:b",
	)

	zoektRepoList := &zoekt.RepoList{
		Repos: []*zoekt.RepoListEntry{
			{
				Repository: zoekt.Repository{
					Name:     "foo/indexed-one",
					Branches: []zoekt.RepositoryBranch{{Name: "HEAD", Version: "deadbeef"}},
				},
			},
			{
				Repository: zoekt.Repository{
					Name:     "foo/indexed-two",
					Branches: []zoekt.RepositoryBranch{{Name: "HEAD", Version: "deadbeef"}},
				},
			},
			{
				Repository: zoekt.Repository{
					Name: "foo/indexed-three",
					Branches: []zoekt.RepositoryBranch{
						{Name: "HEAD", Version: "deadbeef"},
						{Name: "foobar", Version: "deadcow"},
					},
				},
			},
		},
	}

	zoekt := &searchbackend.Zoekt{Client: &fakeSearcher{repos: zoektRepoList}}
	ctx := context.Background()

	makeIndexed := func(repos []*search.RepositoryRevisions) []*search.RepositoryRevisions {
		var indexed []*search.RepositoryRevisions
		for _, r := range repos {
			rev := &search.RepositoryRevisions{
				Repo: r.Repo,
				Revs: r.Revs,
			}
			rev.SetIndexedHEADCommit("deadbeef")
			indexed = append(indexed, rev)
		}
		return indexed
	}

	cases := []struct {
		name      string
		repos     []*search.RepositoryRevisions
		indexed   []*search.RepositoryRevisions
		unindexed []*search.RepositoryRevisions
	}{{
		name:      "all",
		repos:     repos,
		indexed:   makeIndexed(repos[:3]),
		unindexed: repos[3:],
	}, {
		name:      "one unindexed",
		repos:     repos[3:4],
		indexed:   repos[:0],
		unindexed: repos[3:4],
	}, {
		name:      "one indexed",
		repos:     repos[:1],
		indexed:   makeIndexed(repos[:1]),
		unindexed: repos[:0],
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			indexed, unindexed, err := zoektIndexedRepos(ctx, zoekt, tc.repos, nil)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tc.indexed, indexed) {
				diff := cmp.Diff(tc.indexed, indexed)
				t.Error("unexpected indexed:", diff)
			}
			if !reflect.DeepEqual(tc.unindexed, unindexed) {
				diff := cmp.Diff(tc.unindexed, unindexed)
				t.Error("unexpected unindexed:", diff)
			}
		})
	}
}

func Benchmark_zoektIndexedRepos(b *testing.B) {
	repoNames := []string{}
	zoektRepos := []*zoekt.RepoListEntry{}

	for i := 0; i < 10000; i++ {
		indexedName := fmt.Sprintf("foo/indexed-%d@", i)
		unindexedName := fmt.Sprintf("foo/unindexed-%d@", i)

		repoNames = append(repoNames, indexedName, unindexedName)

		zoektRepos = append(zoektRepos, &zoekt.RepoListEntry{
			Repository: zoekt.Repository{
				Name:     strings.TrimSuffix(indexedName, "@"),
				Branches: []zoekt.RepositoryBranch{{Name: "HEAD", Version: "deadbeef"}},
			},
		})
	}

	repos := makeRepositoryRevisions(repoNames...)
	z := &searchbackend.Zoekt{Client: &fakeSearcher{repos: &zoekt.RepoList{Repos: zoektRepos}}}
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_, _, _ = zoektIndexedRepos(ctx, z, repos, nil)
	}
}
