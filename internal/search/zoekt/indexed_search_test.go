package zoekt

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grafana/regexp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/zoekt"
	zoektquery "github.com/sourcegraph/zoekt/query"
	"github.com/stretchr/testify/require"

	"github.com/RoaringBitmap/roaring"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestIndexedSearch(t *testing.T) {
	zeroTimeoutCtx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	type args struct {
		ctx             context.Context
		query           string
		fileMatchLimit  int32
		selectPath      filter.SelectPath
		repos           []*search.RepositoryRevisions
		useFullDeadline bool
		results         []zoekt.FileMatch
		since           func(time.Time) time.Duration
	}

	reposHEAD := makeRepositoryRevisions("foo/bar", "foo/foobar")
	zoektRepos := []*zoekt.RepoListEntry{{
		Repository: zoekt.Repository{
			ID:       uint32(reposHEAD[0].Repo.ID),
			Name:     "foo/bar",
			Branches: []zoekt.RepositoryBranch{{Name: "HEAD", Version: "barHEADSHA"}, {Name: "dev", Version: "bardevSHA"}, {Name: "main", Version: "barmainSHA"}},
		},
	}, {
		Repository: zoekt.Repository{
			ID:       uint32(reposHEAD[1].Repo.ID),
			Name:     "foo/foobar",
			Branches: []zoekt.RepositoryBranch{{Name: "HEAD", Version: "foobarHEADSHA"}},
		},
	}}

	fooSlashBar := zoektRepos[0].Repository
	fooSlashFooBar := zoektRepos[1].Repository

	tests := []struct {
		name               string
		args               args
		wantMatchCount     int
		wantMatchKeys      []result.Key
		wantMatchInputRevs []string
		wantUnindexed      []*search.RepositoryRevisions
		wantCommon         streaming.Stats
		wantErr            bool
	}{
		{
			name: "no matches",
			args: args{
				ctx:             context.Background(),
				repos:           reposHEAD,
				useFullDeadline: false,
				since:           func(time.Time) time.Duration { return time.Second - time.Millisecond },
			},
			wantErr: false,
		},
		{
			name: "no matches timeout",
			args: args{
				ctx:             context.Background(),
				repos:           reposHEAD,
				useFullDeadline: false,
				since:           func(time.Time) time.Duration { return time.Minute },
			},
			wantCommon: streaming.Stats{
				Status: mkStatusMap(map[string]search.RepoStatus{
					"foo/bar":    search.RepoStatusTimedout,
					"foo/foobar": search.RepoStatusTimedout,
				}),
			},
		},
		{
			name: "context timeout",
			args: args{
				ctx:             zeroTimeoutCtx,
				repos:           reposHEAD,
				useFullDeadline: true,
				since:           func(time.Time) time.Duration { return 0 },
			},
			wantErr: true,
		},
		{
			name: "results",
			args: args{
				ctx:             context.Background(),
				fileMatchLimit:  100,
				repos:           makeRepositoryRevisions("foo/bar", "foo/foobar"),
				useFullDeadline: false,
				results: []zoekt.FileMatch{
					{
						Repository:   "foo/bar",
						RepositoryID: fooSlashBar.ID,
						Branches:     []string{"HEAD"},
						Version:      "1",
						FileName:     "baz.go",
						ChunkMatches: []zoekt.ChunkMatch{{
							Content: []byte("I'm like 1.5+ hours into writing this test :'("),
							Ranges: []zoekt.Range{{
								Start: zoekt.Location{0, 1, 1},
								End:   zoekt.Location{5, 1, 6},
							}},
						}, {
							Content: []byte("I'm ready for the rain to stop."),
							Ranges: []zoekt.Range{{
								Start: zoekt.Location{0, 1, 1},
								End:   zoekt.Location{5, 1, 6},
							}, {
								Start: zoekt.Location{5, 1, 6},
								End:   zoekt.Location{15, 1, 16},
							}},
						}},
					},
					{
						Repository:   "foo/foobar",
						RepositoryID: fooSlashFooBar.ID,
						Branches:     []string{"HEAD"},
						Version:      "2",
						FileName:     "baz.go",
						ChunkMatches: []zoekt.ChunkMatch{{
							Content: []byte("s/rain/pain"),
							Ranges: []zoekt.Range{{
								Start: zoekt.Location{0, 1, 1},
								End:   zoekt.Location{5, 1, 6},
							}, {
								Start: zoekt.Location{5, 1, 6},
								End:   zoekt.Location{7, 1, 8},
							}},
						}},
					},
				},
				since: func(time.Time) time.Duration { return 0 },
			},
			wantMatchCount: 5,
			wantMatchKeys: []result.Key{
				{Repo: "foo/bar", Rev: "HEAD", Commit: "1", Path: "baz.go"},
				{Repo: "foo/foobar", Rev: "HEAD", Commit: "2", Path: "baz.go"},
			},
			wantMatchInputRevs: []string{
				"HEAD",
				"HEAD",
			},
			wantErr: false,
		},
		{
			name: "results multi-branch",
			args: args{
				ctx:             context.Background(),
				fileMatchLimit:  100,
				repos:           makeRepositoryRevisions("foo/bar@HEAD:dev:main"),
				useFullDeadline: false,
				results: []zoekt.FileMatch{
					{
						Repository:   "foo/bar",
						RepositoryID: fooSlashBar.ID,
						// baz.go is the same in HEAD and dev
						Branches: []string{"HEAD", "dev"},
						FileName: "baz.go",
						Version:  "1",
					},
					{
						Repository:   "foo/bar",
						RepositoryID: fooSlashBar.ID,
						Branches:     []string{"dev"},
						FileName:     "bam.go",
						Version:      "2",
					},
				},
				since: func(time.Time) time.Duration { return 0 },
			},
			wantMatchCount: 3,
			wantMatchKeys: []result.Key{
				{Repo: "foo/bar", Rev: "HEAD", Commit: "1", Path: "baz.go"},
				{Repo: "foo/bar", Rev: "dev", Commit: "1", Path: "baz.go"},
				{Repo: "foo/bar", Rev: "dev", Commit: "2", Path: "bam.go"},
			},
			wantMatchInputRevs: []string{
				"HEAD",
				"dev",
				"dev",
			},
			wantErr: false,
		},
		{
			// if we search a branch that is indexed and unindexed, we should
			// split the repository revision into the indexed and unindexed
			// parts.
			name: "split branch",
			args: args{
				ctx:             context.Background(),
				fileMatchLimit:  100,
				repos:           makeRepositoryRevisions("foo/bar@HEAD:unindexed"),
				useFullDeadline: false,
				results: []zoekt.FileMatch{
					{
						Repository:   "foo/bar",
						RepositoryID: fooSlashBar.ID,
						Branches:     []string{"HEAD"},
						FileName:     "baz.go",
						Version:      "1",
					},
				},
			},
			wantUnindexed: makeRepositoryRevisions("foo/bar@unindexed"),
			wantMatchKeys: []result.Key{
				{Repo: "foo/bar", Rev: "HEAD", Commit: "1", Path: "baz.go"},
			},
			wantMatchCount:     1,
			wantMatchInputRevs: []string{"HEAD"},
		},
		{
			// Fallback to unindexed search if the query contains ref-globs.
			name: "ref-glob with explicit /*",
			args: args{
				ctx:             context.Background(),
				query:           "repo:foo/bar@*refs/heads/*",
				fileMatchLimit:  100,
				repos:           makeRepositoryRevisions("foo/bar@HEAD"),
				useFullDeadline: false,
				results:         []zoekt.FileMatch{},
			},
			wantUnindexed:      makeRepositoryRevisions("foo/bar@HEAD"),
			wantMatchKeys:      nil,
			wantMatchInputRevs: nil,
		},
		{
			name: "ref-glob with implicit /*",
			args: args{
				ctx:             context.Background(),
				query:           "repo:foo/bar@*refs/tags",
				fileMatchLimit:  100,
				repos:           makeRepositoryRevisions("foo/bar@HEAD"),
				useFullDeadline: false,
				results:         []zoekt.FileMatch{},
			},
			wantUnindexed:      makeRepositoryRevisions("foo/bar@HEAD"),
			wantMatchKeys:      nil,
			wantMatchInputRevs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := query.ParseLiteral(tt.args.query)
			if err != nil {
				t.Fatal(err)
			}

			fakeZoekt := &searchbackend.FakeStreamer{
				Results: []*zoekt.SearchResult{{Files: tt.args.results}},
				Repos:   zoektRepos,
			}

			var resultTypes result.Types
			zoektQuery, err := QueryToZoektQuery(query.Basic{}, resultTypes, &search.Features{}, search.TextRequest)
			if err != nil {
				t.Fatal(err)
			}

			// This is a quick fix which will break once we enable the zoekt client for true streaming.
			// Once we return more than one event we have to account for the proper order of results
			// in the tests.
			agg := streaming.NewAggregatingStream()

			indexed, unindexed, err := PartitionRepos(
				context.Background(),
				logtest.Scoped(t),
				tt.args.repos,
				fakeZoekt,
				search.TextRequest,
				query.Yes,
				query.ContainsRefGlobs(q),
			)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tt.wantUnindexed, unindexed, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("unindexed mismatch (-want +got):\n%s", diff)
			}

			zoektParams := &search.ZoektParameters{
				FileMatchLimit: tt.args.fileMatchLimit,
				Select:         tt.args.selectPath,
			}

			zoektJob := &RepoSubsetTextSearchJob{
				Repos:       indexed,
				Query:       zoektQuery,
				Typ:         search.TextRequest,
				ZoektParams: zoektParams,
				Since:       tt.args.since,
			}

			_, err = zoektJob.Run(tt.args.ctx, job.RuntimeClients{Zoekt: fakeZoekt}, agg)
			if (err != nil) != tt.wantErr {
				t.Errorf("zoektSearchHEAD() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			gotFm, err := matchesToFileMatches(agg.Results)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(&tt.wantCommon, &agg.Stats, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("common mismatch (-want +got):\n%s", diff)
			}

			var gotMatchCount int
			var gotMatchKeys []result.Key
			var gotMatchInputRevs []string
			for _, m := range gotFm {
				gotMatchCount += m.ResultCount()
				gotMatchKeys = append(gotMatchKeys, m.Key())
				if m.InputRev != nil {
					gotMatchInputRevs = append(gotMatchInputRevs, *m.InputRev)
				}
			}
			if diff := cmp.Diff(tt.wantMatchKeys, gotMatchKeys); diff != "" {
				t.Errorf("match URLs mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantMatchInputRevs, gotMatchInputRevs); diff != "" {
				t.Errorf("match InputRevs mismatch (-want +got):\n%s", diff)
			}
			if gotMatchCount != tt.wantMatchCount {
				t.Errorf("gotMatchCount = %v, want %v", gotMatchCount, tt.wantMatchCount)
			}
		})
	}
}

func mkStatusMap(m map[string]search.RepoStatus) search.RepoStatusMap {
	var rsm search.RepoStatusMap
	for name, status := range m {
		rsm.Update(mkRepos(name)[0].ID, status)
	}
	return rsm
}

func TestZoektIndexedRepos(t *testing.T) {
	repos := makeRepositoryRevisions(
		"foo/indexed-one@HEAD",
		"foo/indexed-two@HEAD",
		"foo/indexed-three@HEAD",
		"foo/partially-indexed@HEAD:bad-rev",
		"foo/unindexed-one",
		"foo/unindexed-two",
	)

	zoektRepos := zoekt.ReposMap{}
	for i, branches := range [][]zoekt.RepositoryBranch{
		{
			{Name: "HEAD", Version: "deadbeef"},
		},
		{
			{Name: "HEAD", Version: "deadbeef"},
		},
		{
			{Name: "HEAD", Version: "deadbeef"},
			{Name: "foobar", Version: "deadcow"},
		},
		{
			{Name: "HEAD", Version: "deadbeef"},
		},
	} {
		r := repos[i]
		branches := branches
		zoektRepos[uint32(r.Repo.ID)] = zoekt.MinimalRepoListEntry{Branches: branches}
	}

	cases := []struct {
		name      string
		repos     []*search.RepositoryRevisions
		indexed   []*search.RepositoryRevisions
		unindexed []*search.RepositoryRevisions
	}{{
		name:  "all",
		repos: repos,
		indexed: []*search.RepositoryRevisions{
			repos[0], repos[1], repos[2],
			{Repo: repos[3].Repo, Revs: repos[3].Revs[:1]},
		},
		unindexed: []*search.RepositoryRevisions{
			{Repo: repos[3].Repo, Revs: repos[3].Revs[1:]},
			repos[4], repos[5],
		},
	}, {
		name:      "one unindexed",
		repos:     repos[4:5],
		indexed:   repos[:0],
		unindexed: repos[4:5],
	}, {
		name:      "one indexed",
		repos:     repos[:1],
		indexed:   repos[:1],
		unindexed: repos[:0],
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			indexed, unindexed := zoektIndexedRepos(zoektRepos, tc.repos, nil)

			if diff := cmp.Diff(repoRevsSliceToMap(tc.indexed), indexed.RepoRevs); diff != "" {
				t.Error("unexpected indexed:", diff)
			}
			if diff := cmp.Diff(tc.unindexed, unindexed); diff != "" {
				t.Error("unexpected unindexed:", diff)
			}
		})
	}
}

func TestZoektIndexedRepos_single(t *testing.T) {
	branchesRepos := func(branch string, repo api.RepoID) map[string]*zoektquery.BranchRepos {
		return map[string]*zoektquery.BranchRepos{
			branch: {
				Branch: branch,
				Repos:  roaring.BitmapOf(uint32(repo)),
			},
		}
	}
	repoRev := func(revSpec string) *search.RepositoryRevisions {
		return &search.RepositoryRevisions{
			Repo: types.MinimalRepo{ID: api.RepoID(1), Name: "test/repo"},
			Revs: []string{revSpec},
		}
	}
	zoektRepos := zoekt.ReposMap{
		1: {
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
	}
	cmpRoaring := func(a, b *roaring.Bitmap) bool {
		arrayA, arrayB := a.ToArray(), b.ToArray()
		if len(arrayA) != len(arrayB) {
			return false
		}
		for i := range arrayA {
			if arrayA[i] != arrayB[i] {
				return false
			}
		}
		return true
	}
	cases := []struct {
		rev               string
		wantIndexed       []*search.RepositoryRevisions
		wantBranchesRepos map[string]*zoektquery.BranchRepos
		wantUnindexed     []*search.RepositoryRevisions
	}{
		{
			rev:               "",
			wantIndexed:       []*search.RepositoryRevisions{repoRev("")},
			wantBranchesRepos: branchesRepos("HEAD", 1),
			wantUnindexed:     []*search.RepositoryRevisions{},
		},
		{
			rev:               "HEAD",
			wantIndexed:       []*search.RepositoryRevisions{repoRev("HEAD")},
			wantBranchesRepos: branchesRepos("HEAD", 1),
			wantUnindexed:     []*search.RepositoryRevisions{},
		},
		{
			rev:               "df3f4e499698e48152b39cd655d8901eaf583fa5",
			wantIndexed:       []*search.RepositoryRevisions{repoRev("df3f4e499698e48152b39cd655d8901eaf583fa5")},
			wantBranchesRepos: branchesRepos("HEAD", 1),
			wantUnindexed:     []*search.RepositoryRevisions{},
		},
		{
			rev:               "df3f4e",
			wantIndexed:       []*search.RepositoryRevisions{repoRev("df3f4e")},
			wantBranchesRepos: branchesRepos("HEAD", 1),
			wantUnindexed:     []*search.RepositoryRevisions{},
		},
		{
			rev:               "d",
			wantIndexed:       []*search.RepositoryRevisions{},
			wantBranchesRepos: map[string]*zoektquery.BranchRepos{},
			wantUnindexed:     []*search.RepositoryRevisions{repoRev("d")},
		},
		{
			rev:               "HEAD^1",
			wantIndexed:       []*search.RepositoryRevisions{},
			wantBranchesRepos: map[string]*zoektquery.BranchRepos{},
			wantUnindexed:     []*search.RepositoryRevisions{repoRev("HEAD^1")},
		},
		{
			rev:               "8ec975423738fe7851676083ebf660a062ed1578",
			wantIndexed:       []*search.RepositoryRevisions{repoRev("8ec975423738fe7851676083ebf660a062ed1578")},
			wantBranchesRepos: branchesRepos("NOT-HEAD", 1),
			wantUnindexed:     []*search.RepositoryRevisions{},
		},
	}

	type ret struct {
		Indexed     map[api.RepoID]*search.RepositoryRevisions
		BranchRepos map[string]*zoektquery.BranchRepos
		Unindexed   []*search.RepositoryRevisions
	}

	for _, tt := range cases {
		indexed, unindexed := zoektIndexedRepos(zoektRepos, []*search.RepositoryRevisions{repoRev(tt.rev)}, nil)
		got := ret{
			Indexed:     indexed.RepoRevs,
			BranchRepos: indexed.branchRepos,
			Unindexed:   unindexed,
		}
		want := ret{
			Indexed:     repoRevsSliceToMap(tt.wantIndexed),
			BranchRepos: tt.wantBranchesRepos,
			Unindexed:   tt.wantUnindexed,
		}
		if !cmp.Equal(want, got, cmp.Comparer(cmpRoaring)) {
			t.Errorf("%s mismatch (-want +got):\n%s", tt.rev, cmp.Diff(want, got))
		}
	}
}

func TestZoektFileMatchToSymbolResults(t *testing.T) {
	symbolInfo := func(sym string) *zoekt.Symbol {
		return &zoekt.Symbol{
			Sym:        sym,
			Kind:       "kind",
			Parent:     "parent",
			ParentKind: "parentkind",
		}
	}

	file := &zoekt.FileMatch{
		FileName:   "bar.go",
		Repository: "foo",
		Language:   "go",
		Version:    "deadbeef",
		ChunkMatches: []zoekt.ChunkMatch{{
			// Skips missing symbol info (shouldn't happen in practice).
			Content:      []byte(""),
			ContentStart: zoekt.Location{LineNumber: 5, Column: 1},
			Ranges: []zoekt.Range{{
				Start: zoekt.Location{LineNumber: 5, Column: 8},
			}},
		}, {
			Content:      []byte("symbol a symbol b"),
			ContentStart: zoekt.Location{LineNumber: 10, Column: 1},
			Ranges: []zoekt.Range{{
				Start: zoekt.Location{LineNumber: 10, Column: 8},
			}, {
				Start: zoekt.Location{LineNumber: 10, Column: 18},
			}},
			SymbolInfo: []*zoekt.Symbol{symbolInfo("a"), symbolInfo("b")},
		}, {
			Content:      []byte("symbol c"),
			ContentStart: zoekt.Location{LineNumber: 15, Column: 1},
			Ranges: []zoekt.Range{{
				Start: zoekt.Location{LineNumber: 15, Column: 8},
			}},
			SymbolInfo: []*zoekt.Symbol{symbolInfo("c")},
		}, {
			Content:      []byte(`bar() { var regex = /.*\//; function baz() { }  } `),
			ContentStart: zoekt.Location{LineNumber: 20, Column: 1},
			Ranges: []zoekt.Range{{
				Start: zoekt.Location{LineNumber: 20, Column: 38},
			}},
			SymbolInfo: []*zoekt.Symbol{symbolInfo("baz")},
		}},
	}

	results := zoektFileMatchToSymbolResults(types.MinimalRepo{Name: "foo"}, "master", file)
	var symbols []result.Symbol
	for _, res := range results {
		symbols = append(symbols, res.Symbol)
	}

	want := []result.Symbol{{
		Name:      "a",
		Line:      10,
		Character: 7,
	}, {
		Name:      "b",
		Line:      10,
		Character: 17,
	}, {
		Name:      "c",
		Line:      15,
		Character: 7,
	}, {
		Name:      "baz",
		Line:      20,
		Character: 37,
	},
	}
	for i := range want {
		want[i].Kind = "kind"
		want[i].Parent = "parent"
		want[i].ParentKind = "parentkind"
		want[i].Path = "bar.go"
		want[i].Language = "go"
	}

	if diff := cmp.Diff(want, symbols); diff != "" {
		t.Fatalf("symbol mismatch (-want +got):\n%s", diff)
	}
}

func repoRevsSliceToMap(rs []*search.RepositoryRevisions) map[api.RepoID]*search.RepositoryRevisions {
	m := map[api.RepoID]*search.RepositoryRevisions{}
	for _, r := range rs {
		m[r.Repo.ID] = r
	}
	return m
}

func TestZoektGlobalQueryScope(t *testing.T) {
	cases := []struct {
		name    string
		opts    search.RepoOptions
		priv    []types.MinimalRepo
		want    string
		wantErr string
	}{{
		name: "any",
		opts: search.RepoOptions{
			Visibility: query.Any,
		},
		want: `(and branch="HEAD" rawConfig:RcOnlyPublic)`,
	}, {
		name: "normal",
		opts: search.RepoOptions{
			Visibility: query.Any,
			NoArchived: true,
			NoForks:    true,
		},
		priv: []types.MinimalRepo{{ID: 1}, {ID: 2}},
		want: `(or (and branch="HEAD" rawConfig:RcOnlyPublic|RcNoForks|RcNoArchived) (branchesrepos HEAD:2))`,
	}, {
		name: "private",
		opts: search.RepoOptions{
			Visibility: query.Private,
		},
		priv: []types.MinimalRepo{{ID: 1}, {ID: 2}},
		want: `(branchesrepos HEAD:2)`,
	}, {
		name: "minusrepofilter",
		opts: search.RepoOptions{
			Visibility:       query.Public,
			MinusRepoFilters: []string{"java"},
		},
		want: `(and branch="HEAD" rawConfig:RcOnlyPublic (not reporegex:"(?i)java"))`,
	}, {
		name: "bad minusrepofilter",
		opts: search.RepoOptions{
			Visibility:       query.Any,
			MinusRepoFilters: []string{"())"},
		},
		wantErr: "invalid regex for -repo filter",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			includePrivate := tc.opts.Visibility == query.Private || tc.opts.Visibility == query.Any
			defaultScope, err := DefaultGlobalQueryScope(tc.opts)
			if err != nil || tc.wantErr != "" {
				if got := fmt.Sprintf("%s", err); !strings.Contains(got, tc.wantErr) {
					t.Fatalf("expected error to contain %q: %s", tc.wantErr, got)
				}
				if tc.wantErr == "" {
					t.Fatalf("unexpected error: %s", err)
				}
				return
			}
			zoektGlobalQuery := NewGlobalZoektQuery(&zoektquery.Const{Value: true}, defaultScope, includePrivate)
			zoektGlobalQuery.ApplyPrivateFilter(tc.priv)
			q := zoektGlobalQuery.Generate()
			if got := zoektquery.Simplify(q).String(); got != tc.want {
				t.Fatalf("unexpected scoped query:\nwant: %s\ngot:  %s", tc.want, got)
			}
		})
	}
}

func TestContextWithoutDeadline(t *testing.T) {
	ctxWithDeadline, cancelWithDeadline := context.WithTimeout(context.Background(), time.Minute)
	defer cancelWithDeadline()

	tr, ctxWithDeadline := trace.New(ctxWithDeadline, "")

	if _, ok := ctxWithDeadline.Deadline(); !ok {
		t.Fatal("expected context to have deadline")
	}

	ctxNoDeadline, cancelNoDeadline := contextWithoutDeadline(ctxWithDeadline)
	defer cancelNoDeadline()

	if _, ok := ctxNoDeadline.Deadline(); ok {
		t.Fatal("expected context to not have deadline")
	}

	// We want to keep trace info
	if tr2 := trace.FromContext(ctxNoDeadline); !tr.SpanContext().Equal(tr2.SpanContext()) {
		t.Error("trace information not propogated")
	}

	// Calling cancelWithDeadline should cancel ctxNoDeadline
	cancelWithDeadline()
	select {
	case <-ctxNoDeadline.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("expected context to be done")
	}
}

func TestContextWithoutDeadline_cancel(t *testing.T) {
	ctxWithDeadline, cancelWithDeadline := context.WithTimeout(context.Background(), time.Minute)
	defer cancelWithDeadline()
	ctxNoDeadline, cancelNoDeadline := contextWithoutDeadline(ctxWithDeadline)

	cancelNoDeadline()
	select {
	case <-ctxNoDeadline.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("expected context to be done")
	}
}

func makeRepositoryRevisions(repos ...string) []*search.RepositoryRevisions {
	r := make([]*search.RepositoryRevisions, len(repos))
	for i, repospec := range repos {
		repoRevs, err := query.ParseRepositoryRevisions(repospec)
		if err != nil {
			panic(errors.Errorf("unexpected error parsing repo spec %s", repospec))
		}

		revs := make([]string, 0, len(repoRevs.Revs))
		for _, revSpec := range repoRevs.Revs {
			revs = append(revs, revSpec.RevSpec)
		}
		if len(revs) == 0 {
			// treat empty list as HEAD
			revs = []string{"HEAD"}
		}
		r[i] = &search.RepositoryRevisions{Repo: mkRepos(repoRevs.Repo)[0], Revs: revs}
	}
	return r
}

func makeRepositoryRevisionsMap(repos ...string) map[api.RepoID]*search.RepositoryRevisions {
	r := makeRepositoryRevisions(repos...)
	rMap := make(map[api.RepoID]*search.RepositoryRevisions, len(r))
	for _, repoRev := range r {
		rMap[repoRev.Repo.ID] = repoRev
	}
	return rMap
}

func mkRepos(names ...string) []types.MinimalRepo {
	var repos []types.MinimalRepo
	for _, name := range names {
		sum := md5.Sum([]byte(name))
		id := api.RepoID(binary.BigEndian.Uint64(sum[:]))
		if id < 0 {
			id = -(id / 2)
		}
		if id == 0 {
			id++
		}
		repos = append(repos, types.MinimalRepo{ID: id, Name: api.RepoName(name)})
	}
	return repos
}

func matchesToFileMatches(matches []result.Match) ([]*result.FileMatch, error) {
	fms := make([]*result.FileMatch, 0, len(matches))
	for _, match := range matches {
		fm, ok := match.(*result.FileMatch)
		if !ok {
			return nil, errors.Errorf("expected only file match results")
		}
		fms = append(fms, fm)
	}
	return fms, nil
}

func TestZoektFileMatchToMultilineMatches(t *testing.T) {
	cases := []struct {
		input  *zoekt.FileMatch
		output result.ChunkMatches
	}{{
		input: &zoekt.FileMatch{
			ChunkMatches: []zoekt.ChunkMatch{{
				Content:      []byte("testing 1 2 3"),
				ContentStart: zoekt.Location{ByteOffset: 0, LineNumber: 1, Column: 1},
				Ranges: []zoekt.Range{{
					Start: zoekt.Location{8, 1, 9},
					End:   zoekt.Location{9, 1, 10},
				}, {
					Start: zoekt.Location{10, 1, 11},
					End:   zoekt.Location{11, 1, 12},
				}, {
					Start: zoekt.Location{12, 1, 13},
					End:   zoekt.Location{13, 1, 14},
				}},
			}},
		},
		// One chunk per line, not one per fragment
		output: result.ChunkMatches{{
			Content:      "testing 1 2 3",
			ContentStart: result.Location{0, 0, 0},
			Ranges: result.Ranges{{
				Start: result.Location{8, 0, 8},
				End:   result.Location{9, 0, 9},
			}, {
				Start: result.Location{10, 0, 10},
				End:   result.Location{11, 0, 11},
			}, {
				Start: result.Location{12, 0, 12},
				End:   result.Location{13, 0, 13},
			}},
		}},
	}}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := zoektFileMatchToMultilineMatches(tc.input)
			require.Equal(t, tc.output, got)
		})
	}
}

func TestZoektFileMatchToPathMatchRanges(t *testing.T) {
	zoektQueryRegexps := []*regexp.Regexp{regexp.MustCompile("python.*worker|stuff")}

	cases := []struct {
		name   string
		input  *zoekt.FileMatch
		output []result.Range
	}{
		{
			name: "returns single path match range",
			input: &zoekt.FileMatch{
				FileName: "internal/python/foo/worker.py",
			},
			output: []result.Range{
				{
					Start: result.Location{Offset: 9, Line: 0, Column: 9},
					End:   result.Location{Offset: 26, Line: 0, Column: 26},
				},
			},
		},
		{
			name: "returns multiple path match ranges",
			input: &zoekt.FileMatch{
				FileName: "internal/python/foo/worker/src/dev/python_stuff.py",
			},
			output: []result.Range{
				{
					Start: result.Location{Offset: 9, Line: 0, Column: 9},
					End:   result.Location{Offset: 26, Line: 0, Column: 26},
				},
				{
					Start: result.Location{Offset: 42, Line: 0, Column: 42},
					End:   result.Location{Offset: 47, Line: 0, Column: 47},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := zoektFileMatchToPathMatchRanges(tc.input, zoektQueryRegexps)
			require.Equal(t, tc.output, got)
		})
	}
}

func TestGetRepoRevsFromBranchRepos_SingleRepo(t *testing.T) {
	cases := []struct {
		name            string
		revisions       []string
		indexedBranches []string
		wantRepoRevs    []string
	}{
		{
			name:            "no revisions specified for the indexed branch",
			indexedBranches: []string{"HEAD"},
			wantRepoRevs:    []string{"HEAD"},
		}, {
			name:            "specific revision is the latest commit ID indexed for the default branch of repo",
			revisions:       []string{"latestCommitID"},
			indexedBranches: []string{"HEAD"},
			wantRepoRevs:    []string{"HEAD"},
		}, {
			name:            "specific revision that is also a non default branch which is indexed",
			revisions:       []string{"myIndexedRevision"},
			indexedBranches: []string{"myIndexedRevision"},
			wantRepoRevs:    []string{"myIndexedRevision"},
		}, {
			name:            "specific revision is the latest commit ID indexed for a non default branch which is indexed",
			revisions:       []string{"latestCommitID"},
			indexedBranches: []string{"myIndexedFeatureBranch"},
			wantRepoRevs:    []string{"myIndexedFeatureBranch"},
		}, {
			name:            "specific revision is the latest commit ID indexed for one of multiple indexed branches",
			revisions:       []string{"someCommitID"},
			indexedBranches: []string{"HEAD", "myIndexedFeatureBranch", "myIndexedRevision"},
			wantRepoRevs:    []string{""},
		}, {
			name:            "specific revision is the latest commit ID indexed for one of multiple indexed branches, including the specified revision",
			revisions:       []string{"someCommitID"},
			indexedBranches: []string{"HEAD", "myIndexedFeatureBranch", "someCommitID"},
			wantRepoRevs:    []string{"someCommitID"},
		}, {
			name:            "multiple specified revisions: one is indexed default branch and one is an indexed revision",
			revisions:       []string{"someCommitID0", "someCommitID1"},
			indexedBranches: []string{"HEAD", "someCommitID0"},
			wantRepoRevs:    []string{"someCommitID0", ""},
		}, {
			name:            "multiple specified revisions: one is an indexed revision and the other cannot be matched by branch name so default to empty string",
			revisions:       []string{"someCommitID0", "someCommitID1"},
			indexedBranches: []string{"myIndexedFeatureBranch", "someCommitID0"},
			wantRepoRevs:    []string{"someCommitID0", ""},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repoWithRevs := "foo/indexed-one"

			if len(tc.revisions) > 1 {
				repoWithRevs = fmt.Sprintf("%v@%v", repoWithRevs, strings.Join(tc.revisions, ":"))
			} else if len(tc.revisions) > 0 {
				repoWithRevs = fmt.Sprintf("%v@%v", repoWithRevs, tc.revisions[0])
			}

			repoRevs := makeRepositoryRevisionsMap(repoWithRevs)

			inputBranchRepos := make(map[string]*zoektquery.BranchRepos, len(tc.indexedBranches))

			if len(repoRevs) != 1 {
				t.Fatal("repoRevs map should represent revisions for no more than one repo with ID")
			}

			var wantRepoID api.RepoID
			for repoID := range repoRevs {
				wantRepoID = repoID
				break
			}

			for _, branch := range tc.indexedBranches {
				repos := roaring.New()
				repos.Add(uint32(wantRepoID))
				inputBranchRepos[branch] = &zoektquery.BranchRepos{Branch: branch, Repos: repos}
			}

			indexed := IndexedRepoRevs{
				RepoRevs:    repoRevs,
				branchRepos: inputBranchRepos,
			}

			gotRepoRevs := indexed.GetRepoRevsFromBranchRepos()
			for _, revs := range gotRepoRevs {
				if diff := cmp.Diff(tc.wantRepoRevs, revs.Revs); diff != "" {
					t.Errorf("unindexed mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
