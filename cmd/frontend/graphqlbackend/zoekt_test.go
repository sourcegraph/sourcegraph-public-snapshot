package graphqlbackend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	zoektrpc "github.com/google/zoekt/rpc"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/schema"
)

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

func TestZoektSearchHEAD(t *testing.T) {
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
	singleRepositoryRevisions := []*search.RepositoryRevisions{rr}

	tests := []struct {
		name              string
		args              args
		wantFm            []*FileMatchResolver
		wantMatchCount    int
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
				since:           func(time.Time) time.Duration { return time.Minute },
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
		{
			name: "returns accurate match count of 5 line fragment matches across two files",
			args: args{
				ctx:             context.Background(),
				query:           &search.TextPatternInfo{PathPatternsAreRegExps: true, FileMatchLimit: 100},
				repos:           makeRepositoryRevisions("foo/bar@master", "foo/foobar@master"),
				useFullDeadline: false,
				searcher: &fakeSearcher{
					repos: &zoekt.RepoList{
						Repos: []*zoekt.RepoListEntry{
							{
								Repository: zoekt.Repository{
									Name: "foo/bar",
								},
							},
							{
								Repository: zoekt.Repository{
									Name: "foo/foobar",
								},
							},
						},
					},
					result: &zoekt.SearchResult{
						Files: []zoekt.FileMatch{
							{
								Repository: "foo/bar",
								FileName:   "baz.go",
								LineMatches: []zoekt.LineMatch{
									{
										Line: []byte("I'm like 1.5+ hours into writing this test :'("),
										LineFragments: []zoekt.LineFragmentMatch{
											{LineOffset: 0, MatchLength: 5},
										},
									},
									{
										Line: []byte("I'm ready for the rain to stop."),
										LineFragments: []zoekt.LineFragmentMatch{
											{LineOffset: 0, MatchLength: 5},
											{LineOffset: 5, MatchLength: 10},
										},
									},
								},
							},
							{
								Repository: "foo/foobar",
								FileName:   "baz.go",
								LineMatches: []zoekt.LineMatch{
									{
										Line: []byte("s/rain/pain"),
										LineFragments: []zoekt.LineFragmentMatch{
											{LineOffset: 0, MatchLength: 5},
											{LineOffset: 5, MatchLength: 2},
										},
									},
								},
							},
						},
					},
				},
				since: func(time.Time) time.Duration { return 0 },
			},
			wantFm:            nil,
			wantLimitHit:      false,
			wantReposLimitHit: map[string]struct{}{},
			wantMatchCount:    5,
			wantErr:           false,
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
			if tt.wantFm != nil {
				if !reflect.DeepEqual(gotFm, tt.wantFm) {
					t.Errorf("zoektSearchHEAD() gotFm = %v, want %v", gotFm, tt.wantFm)
				}
			}
			if gotLimitHit != tt.wantLimitHit {
				t.Errorf("zoektSearchHEAD() gotLimitHit = %v, want %v", gotLimitHit, tt.wantLimitHit)
			}
			if !reflect.DeepEqual(gotReposLimitHit, tt.wantReposLimitHit) {
				t.Errorf("zoektSearchHEAD() gotReposLimitHit = %v, want %v", gotReposLimitHit, tt.wantReposLimitHit)
			}

			var gotMatchCount int
			for _, m := range gotFm {
				gotMatchCount += m.MatchCount
			}
			if gotMatchCount != tt.wantMatchCount {
				t.Errorf("zoektSearchHEAD() gotMatchCount = %v, want %v", gotMatchCount, tt.wantMatchCount)
			}
		})
	}
}

func TestZoektIndexedRepos(t *testing.T) {
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

func TestCreateNewRepoSetWithRepoHasFileInputs(t *testing.T) {
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

func TestZoektResultCountFactor(t *testing.T) {
	cases := []struct {
		name     string
		numRepos int
		pattern  *search.TextPatternInfo
		want     int
	}{
		{
			name:     "One repo implies max scaling factor",
			numRepos: 1,
			pattern:  &search.TextPatternInfo{},
			want:     100,
		},
		{
			name:     "Eleven repos implies a scaling factor between min and max",
			numRepos: 11,
			pattern:  &search.TextPatternInfo{},
			want:     8,
		},
		{
			name:     "More than 500 repos implies a min scaling factor",
			numRepos: 501,
			pattern:  &search.TextPatternInfo{},
			want:     1,
		},
		{
			name:     "Setting a count greater than defautl max results (30) adapts scaling factor",
			numRepos: 501,
			pattern:  &search.TextPatternInfo{FileMatchLimit: 100},
			want:     10,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := zoektResultCountFactor(tt.numRepos, tt.pattern)
			if tt.want != got {
				t.Fatalf("Want scaling factor %d but got %d", tt.want, got)
			}
		})
	}
}

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

	mockDecodedViewerFinalSettings = &schema.Settings{}
	defer func() { mockDecodedViewerFinalSettings = nil }()

	db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
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
	dbtesting.SetupGlobalTestDB(b)

	ctx := context.Background()

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

	rows := make([]*sqlf.Query, 0, len(repos))
	for _, r := range repos {
		rows = append(rows, sqlf.Sprintf(
			"(%s, %s, %s, %s, %s, %s, %s)",
			r.Name,
			r.Description,
			r.Fork,
			true,
			r.ExternalRepo.ServiceType,
			r.ExternalRepo.ServiceID,
			r.ExternalRepo.ID,
		))
	}

	q := sqlf.Sprintf(`
		INSERT INTO repo (
			name,
			description,
			fork,
			enabled,
			external_service_type,
			external_service_id,
			external_id
		)
		VALUES %s`,
		sqlf.Join(rows, ","),
	)

	_, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		b.Fatal(err)
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

func generateRepos(count int) ([]*types.Repo, []*types.Repo, []*zoekt.RepoListEntry) {
	var reposWithIDs []*types.Repo
	var repos []*types.Repo
	var zoektRepos []*zoekt.RepoListEntry

	for i := 1; i <= count; i++ {
		name := fmt.Sprintf("repo-%d", i)

		repoWithIDs := &types.Repo{
			ID:   api.RepoID(i),
			Name: api.RepoName(name),
			ExternalRepo: api.ExternalRepoSpec{
				ID:          name,
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com",
			}}

		reposWithIDs = append(reposWithIDs, repoWithIDs)

		repos = append(repos, &types.Repo{

			ID:           repoWithIDs.ID,
			Name:         repoWithIDs.Name,
			ExternalRepo: repoWithIDs.ExternalRepo,

			RepoFields: &types.RepoFields{
				URI:         fmt.Sprintf("https://github.com/foobar/%s", repoWithIDs.Name),
				Description: "this repositoriy contains a side project that I haven't maintained in 2 years",
				Language:    "v-language",
			}})

		zoektRepos = append(zoektRepos, &zoekt.RepoListEntry{
			Repository: zoekt.Repository{
				Name:     name,
				Branches: []zoekt.RepositoryBranch{{Name: "HEAD", Version: "deadbeef"}},
			},
		})
	}
	return reposWithIDs, repos, zoektRepos
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

func TestZoektSingleIndexedRepo(t *testing.T) {
	repoRev := func(revSpec string) *search.RepositoryRevisions {
		return &search.RepositoryRevisions{
			Repo: &types.Repo{ID: api.RepoID(0), Name: "test/repo"},
			Revs: []search.RevisionSpecifier{
				{RevSpec: revSpec},
			},
		}
	}
	zoektRepos := []*zoekt.RepoListEntry{{
		Repository: zoekt.Repository{
			Name: "test/repo",
			Branches: []zoekt.RepositoryBranch{
				{
					Name:    "HEAD",
					Version: "df3f4e499698e48152b39cd655d8901eaf583fa5",
				},
				{
					Name:    "NOT-HEAD",
					Version: "8ec975423738fe7851676083ebf660a062ed1578",
				},
			},
		},
	}}
	z := &searchbackend.Zoekt{
		Client: &fakeSearcher{
			repos: &zoekt.RepoList{Repos: zoektRepos},
		},
		DisableCache: true,
	}
	cases := []struct {
		rev           string
		wantIndexed   []*search.RepositoryRevisions
		wantUnindexed []*search.RepositoryRevisions
	}{
		{
			rev:           "",
			wantIndexed:   []*search.RepositoryRevisions{repoRev("")},
			wantUnindexed: []*search.RepositoryRevisions{},
		},
		{
			rev:           "HEAD",
			wantIndexed:   []*search.RepositoryRevisions{repoRev("HEAD")},
			wantUnindexed: []*search.RepositoryRevisions{},
		},
		{
			rev:           "df3f4e499698e48152b39cd655d8901eaf583fa5",
			wantIndexed:   []*search.RepositoryRevisions{repoRev("df3f4e499698e48152b39cd655d8901eaf583fa5")},
			wantUnindexed: []*search.RepositoryRevisions{},
		},
		{
			rev:           "df3f4e",
			wantIndexed:   []*search.RepositoryRevisions{repoRev("df3f4e")},
			wantUnindexed: []*search.RepositoryRevisions{},
		},
		{
			rev:           "d",
			wantIndexed:   []*search.RepositoryRevisions{},
			wantUnindexed: []*search.RepositoryRevisions{repoRev("d")},
		},
		{
			rev:           "HEAD^1",
			wantIndexed:   []*search.RepositoryRevisions{},
			wantUnindexed: []*search.RepositoryRevisions{repoRev("HEAD^1")},
		},
		{
			rev:           "8ec975423738fe7851676083ebf660a062ed1578",
			wantUnindexed: []*search.RepositoryRevisions{},
			wantIndexed:   []*search.RepositoryRevisions{repoRev("8ec975423738fe7851676083ebf660a062ed1578")},
		},
	}

	type ret struct {
		Indexed, Unindexed []*search.RepositoryRevisions
	}

	for _, tt := range cases {
		filter := func(*zoekt.Repository) bool { return true }
		indexed, unindexed, err := zoektSingleIndexedRepo(context.Background(), z, repoRev(tt.rev), filter)
		if err != nil {
			t.Fatal(err)
		}
		got := ret{
			Indexed:   indexed,
			Unindexed: unindexed,
		}
		want := ret{
			Indexed:   tt.wantIndexed,
			Unindexed: tt.wantUnindexed,
		}
		if !cmp.Equal(want, got) {
			t.Errorf("%s mismatch (-want +got):\n%s", tt.rev, cmp.Diff(want, got))
		}
	}
}
