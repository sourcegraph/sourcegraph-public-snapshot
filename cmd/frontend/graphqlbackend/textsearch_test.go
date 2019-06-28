package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	searchbackend "github.com/sourcegraph/sourcegraph/pkg/search/backend"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestQueryToZoektQuery(t *testing.T) {
	cases := []struct {
		Name    string
		Pattern *search.PatternInfo
		Query   string
	}{
		{
			Name: "substr",
			Pattern: &search.PatternInfo{
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
			Pattern: &search.PatternInfo{
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
			Pattern: &search.PatternInfo{
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
			Pattern: &search.PatternInfo{
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
			Pattern: &search.PatternInfo{
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
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			q, err := zoektquery.Parse(tt.Query)
			if err != nil {
				t.Fatalf("failed to parse %q: %v", tt.Query, err)
			}
			got, err := queryToZoektQuery(tt.Pattern)
			if err != nil {
				t.Fatal("queryToZoektQuery failed:", err)
			}
			if !queryEqual(got, q) {
				t.Fatalf("mismatched queries\ngot  %s\nwant %s", got.String(), q.String())
			}
		})
	}
}

func queryEqual(a zoektquery.Q, b zoektquery.Q) bool {
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
		Pattern *search.PatternInfo
		Query   []string
		// This should be the same value passed in to either FilePatternsReposMustInclude or FilePatternsReposMustExclude
		ListOfFilePaths []string
	}{
		{
			Name: "single repohasfile filter",
			Pattern: &search.PatternInfo{
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
			Pattern: &search.PatternInfo{
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
			Pattern: &search.PatternInfo{
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
			Pattern: &search.PatternInfo{
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
	mockSearchFilesInRepo = func(ctx context.Context, repo *types.Repo, gitserverRepo gitserver.Repo, rev string, info *search.PatternInfo, fetchTimeout time.Duration) (matches []*fileMatchResolver, limitHit bool, err error) {
		repoName := repo.Name
		switch repoName {
		case "foo/one":
			return []*fileMatchResolver{
				{
					uri: "git://" + string(repoName) + "?" + rev + "#" + "main.go",
				},
			}, false, nil
		case "foo/two":
			return []*fileMatchResolver{
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
			return nil, false, &git.RevisionNotFoundError{Repo: repoName, Spec: "missing"}
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
	args := &search.Args{
		Pattern: &search.PatternInfo{
			FileMatchLimit: defaultMaxSearchResults,
			Pattern:        "foo",
		},
		Repos: makeRepositoryRevisions("foo/one", "foo/two", "foo/empty", "foo/cloning", "foo/missing", "foo/missing-db", "foo/timedout", "foo/no-rev"),
		Query: q,
		Zoekt: zoekt,
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
	args = &search.Args{
		Pattern: &search.PatternInfo{
			FileMatchLimit: defaultMaxSearchResults,
			Pattern:        "foo",
		},
		Repos: makeRepositoryRevisions("foo/no-rev@dev"),
		Query: q,
		Zoekt: zoekt,
	}
	_, _, err = searchFilesInRepos(context.Background(), args)
	if !git.IsRevisionNotFound(errors.Cause(err)) {
		t.Fatalf("searching non-existent rev expected to fail with RevisionNotFoundError got: %v", err)
	}
}

func TestRepoShouldBeSearched(t *testing.T) {
	mockTextSearch = func(ctx context.Context, repo gitserver.Repo, commit api.CommitID, p *search.PatternInfo, fetchTimeout time.Duration) (matches []*fileMatchResolver, limitHit bool, err error) {
		repoName := repo.Name
		switch repoName {
		case "foo/one":
			return []*fileMatchResolver{
				{
					uri: "git://" + string(repoName) + "?1a2b3c#" + "main.go",
				},
			}, false, nil
		case "foo/no-filematch":
			return []*fileMatchResolver{}, false, nil
		default:
			return nil, false, errors.New("Unexpected repo")
		}
	}
	defer func() { mockTextSearch = nil }()
	info := &search.PatternInfo{
		FileMatchLimit:               defaultMaxSearchResults,
		Pattern:                      "foo",
		FilePatternsReposMustInclude: []string{"main"},
	}

	shouldBeSearched, err := repoShouldBeSearched(context.Background(), info, gitserver.Repo{Name: "foo/one", URL: "http://example.com/foo/one"}, "1a2b3c", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !shouldBeSearched {
		t.Errorf("expected repo to be searched, got shouldn't be searched")
	}

	shouldBeSearched, err = repoShouldBeSearched(context.Background(), info, gitserver.Repo{Name: "foo/no-filematch", URL: "http://example.com/foo/no-filematch"}, "1a2b3c", time.Minute)
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
		query           *search.PatternInfo
		indexedCommits  map[*search.RepositoryRevisions]string
		repos           []*search.RepositoryRevisions
		useFullDeadline bool
		searcher        zoekt.Searcher
		opts            zoekt.SearchOptions
		since           func(time.Time) time.Duration
	}

	singleRepositoryRevisions := []*search.RepositoryRevisions{
		{Repo: &types.Repo{}},
	}
	singleIndexedCommits := map[*search.RepositoryRevisions]string{
		singleRepositoryRevisions[0]: "abc",
	}

	tests := []struct {
		name              string
		args              args
		wantFm            []*fileMatchResolver
		wantLimitHit      bool
		wantReposLimitHit map[string]struct{}
		wantErr           bool
	}{
		{
			name: "returns no error if search completed with no matches before timeout",
			args: args{
				ctx:             context.Background(),
				query:           &search.PatternInfo{PathPatternsAreRegExps: true},
				indexedCommits:  singleIndexedCommits,
				repos:           singleRepositoryRevisions,
				useFullDeadline: false,
				searcher:        &fakeSearcher{result: &zoekt.SearchResult{}},
				opts:            zoekt.SearchOptions{MaxWallTime: time.Second},
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
				query:           &search.PatternInfo{PathPatternsAreRegExps: true},
				indexedCommits:  singleIndexedCommits,
				repos:           singleRepositoryRevisions,
				useFullDeadline: false,
				searcher:        &fakeSearcher{result: &zoekt.SearchResult{}},
				opts:            zoekt.SearchOptions{MaxWallTime: time.Second},
				since:           func(time.Time) time.Duration { return time.Second },
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
				query:           &search.PatternInfo{PathPatternsAreRegExps: true},
				indexedCommits:  singleIndexedCommits,
				repos:           singleRepositoryRevisions,
				useFullDeadline: true,
				searcher:        &fakeSearcher{result: &zoekt.SearchResult{}},
				opts:            zoekt.SearchOptions{},
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
				query:           &search.PatternInfo{PathPatternsAreRegExps: true},
				indexedCommits:  singleIndexedCommits,
				repos:           singleRepositoryRevisions,
				useFullDeadline: true,
				searcher:        &errorSearcher{err: errors.New("womp womp")},
				opts:            zoekt.SearchOptions{},
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
			gotFm, gotLimitHit, gotReposLimitHit, err := zoektSearchHEAD(tt.args.ctx, tt.args.query, tt.args.repos, tt.args.indexedCommits, tt.args.useFullDeadline, tt.args.searcher, tt.args.opts, tt.args.since)
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

func Test_createNewRepoSetWithRepoHasFileInputs(t *testing.T) {
	type args struct {
		ctx                             context.Context
		queryPatternInfo                *search.PatternInfo
		searcher                        zoekt.Searcher
		repoSet                         zoektquery.RepoSet
		repoHasFileFlagIsInQuery        bool
		negatedRepoHasFileFlagIsInQuery bool
	}

	tests := []struct {
		name        string
		args        args
		wantRepoSet *zoektquery.RepoSet
	}{
		{
			name: "returns filtered repoSet when repoHasFileFlag is in query",
			args: args{
				queryPatternInfo: &search.PatternInfo{FilePatternsReposMustInclude: []string{"1"}, PathPatternsAreRegExps: true},
				searcher: &fakeSearcher{result: &zoekt.SearchResult{
					Files: []zoekt.FileMatch{{
						FileName:   "1.md",
						Repository: "github.com/test/1",
						LineMatches: []zoekt.LineMatch{{
							FileName: true,
						}}},
					},
					RepoURLs: map[string]string{"github.com/test/1": "github.com/test/1"}}},
				repoSet:                         zoektquery.RepoSet{Set: map[string]bool{"github.com/test/1": true, "github.com/test/2": true}},
				repoHasFileFlagIsInQuery:        true,
				negatedRepoHasFileFlagIsInQuery: false,
			},
			wantRepoSet: &zoektquery.RepoSet{Set: map[string]bool{"github.com/test/1": true}},
		},
		{
			name: "returns filtered repoSet when multiple repoHasFileFlags are in query",
			args: args{
				queryPatternInfo: &search.PatternInfo{FilePatternsReposMustInclude: []string{"1", "2"}, PathPatternsAreRegExps: true},
				searcher: &fakeSearcher{result: &zoekt.SearchResult{
					Files: []zoekt.FileMatch{{
						FileName:   "1.md",
						Repository: "github.com/test/1",
						LineMatches: []zoekt.LineMatch{{
							FileName: true,
						}}},
						{
							FileName:   "1.md",
							Repository: "github.com/test/2",
							LineMatches: []zoekt.LineMatch{{
								FileName: true,
							}}},
						{
							FileName:   "2.md",
							Repository: "github.com/test/2",
							LineMatches: []zoekt.LineMatch{{
								FileName: true,
							}}},
					},
					RepoURLs: map[string]string{"github.com/test/2": "github.com/test/2"}}},
				repoSet:                         zoektquery.RepoSet{Set: map[string]bool{"github.com/test/1": true, "github.com/test/2": true}},
				repoHasFileFlagIsInQuery:        true,
				negatedRepoHasFileFlagIsInQuery: false,
			},
			wantRepoSet: &zoektquery.RepoSet{Set: map[string]bool{"github.com/test/2": true}},
		},
		{
			name: "returns filtered repoSet when negated repoHasFileFlag is in query",
			args: args{
				queryPatternInfo: &search.PatternInfo{FilePatternsReposMustExclude: []string{"1"}, PathPatternsAreRegExps: true},
				searcher: &fakeSearcher{result: &zoekt.SearchResult{
					Files: []zoekt.FileMatch{{
						FileName:   "1.md",
						Repository: "github.com/test/1",
						LineMatches: []zoekt.LineMatch{{
							FileName: true,
						}}},
					},
					RepoURLs: map[string]string{"github.com/test/1": "github.com/test/1"}}},
				repoSet:                         zoektquery.RepoSet{Set: map[string]bool{"github.com/test/1": true, "github.com/test/2": true}},
				repoHasFileFlagIsInQuery:        false,
				negatedRepoHasFileFlagIsInQuery: true,
			},
			wantRepoSet: &zoektquery.RepoSet{Set: map[string]bool{"github.com/test/2": true}},
		},
		{
			name: "returns a new repoSet that includes at most the repos from original repoSet",
			args: args{
				queryPatternInfo: &search.PatternInfo{FilePatternsReposMustInclude: []string{"1"}, PathPatternsAreRegExps: true},
				searcher: &fakeSearcher{result: &zoekt.SearchResult{
					Files: []zoekt.FileMatch{{
						FileName:   "1.md",
						Repository: "github.com/test/1",
						LineMatches: []zoekt.LineMatch{{
							FileName: true,
						}}},
						{
							FileName:   "1.md",
							Repository: "github.com/test/2",
							LineMatches: []zoekt.LineMatch{{
								FileName: true,
							}}},
					},
					RepoURLs: map[string]string{"github.com/test/1": "github.com/test/1", "github.com/test/2": "github.com/test/2"}}},
				repoSet:                         zoektquery.RepoSet{Set: map[string]bool{"github.com/test/1": true}},
				repoHasFileFlagIsInQuery:        false,
				negatedRepoHasFileFlagIsInQuery: true,
			},
			wantRepoSet: &zoektquery.RepoSet{Set: map[string]bool{"github.com/test/1": true}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRepoSet, err := createNewRepoSetWithRepoHasFileInputs(tt.args.ctx, tt.args.queryPatternInfo, tt.args.searcher, tt.args.repoSet)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotRepoSet, tt.wantRepoSet) {
				t.Errorf("createNewRepoSetWithRepoHasFileInputs() gotRepoSet = %v, want %v", gotRepoSet, tt.wantRepoSet)
			}
		})
	}
}

func Test_zoektIndexedRepos(t *testing.T) {
	repos := makeRepositoryRevisions(
		"foo/indexed-one@",
		"foo/indexed-two@",
		"foo/indexed-three-no-HEAD@",
		"foo/unindexed-one",
		"foo/unindexed-two",
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
					Name:     "foo/indexed-three-no-HEAD",
					Branches: []zoekt.RepositoryBranch{{Name: "foobar", Version: "deadbeef"}},
				},
			},
		},
	}

	zoekt := &searchbackend.Zoekt{Client: &fakeSearcher{repos: zoektRepoList}}
	ctx := context.Background()

	indexed, unindexed, indexedCommits, err := zoektIndexedRepos(ctx, zoekt, repos)
	if err != nil {
		t.Fatal(err)
	}

	for _, assertion := range []struct {
		name       string
		have, want []*search.RepositoryRevisions
	}{
		{"indexed", indexed, makeRepositoryRevisions("foo/indexed-one@", "foo/indexed-two@", "foo/indexed-three-no-HEAD@")},
		{"unindexed", unindexed, makeRepositoryRevisions("foo/unindexed-one", "foo/unindexed-two")},
	} {
		sort.Slice(assertion.have, sortRepoRevsByName(assertion.have))
		sort.Slice(assertion.want, sortRepoRevsByName(assertion.want))

		if !reflect.DeepEqual(assertion.have, assertion.want) {
			diff := cmp.Diff(assertion.have, assertion.want)
			t.Fatalf("%s has wrong repo revs. diff=%s", assertion.name, diff)
		}
	}

	wantIndexedCommits := map[*search.RepositoryRevisions]string{
		repos[0]: "deadbeef",
		repos[1]: "deadbeef",
	}

	if !reflect.DeepEqual(indexedCommits, wantIndexedCommits) {
		diff := cmp.Diff(indexedCommits, wantIndexedCommits)
		t.Fatalf("indexedCommits has wrong revisions. diff=%s", diff)
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
				Name:     indexedName,
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
		_, _, _, _ = zoektIndexedRepos(ctx, z, repos)
	}
}

func sortRepoRevsByName(repoRevs []*search.RepositoryRevisions) func(int, int) bool {
	return func(i, j int) bool { return repoRevs[i].Repo.Name < repoRevs[j].Repo.Name }
}

func init() {
	// Set both URLs to something that will fail in tests. We shouldn't be
	// contacting them in tests.
	zoektAddr = "127.0.0.1:101010"
	searcherURL = "http://127.0.0.1:101010"
}
