package graphqlbackend

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSearchPagination_unmarshalSearchCursor(t *testing.T) {
	got, err := unmarshalSearchCursor(nil)
	if got != nil || err != nil {
		t.Fatal("expected got == nil && err == nil for nil input")
	}

	want := &run.SearchCursor{
		RepositoryOffset: 1,
		ResultOffset:     2,
	}
	enc := marshalSearchCursor(want)
	if enc == "" {
		t.Fatal("expected encoded string")
	}
	got, err = unmarshalSearchCursor(&enc)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("expected got == want")
	}
}

func TestSearchPagination_sliceSearchResults(t *testing.T) {
	repoName := func(name string) types.RepoName {
		// Backcompat extract ID from name.
		id := name[len(name)-1] - '0'
		return types.RepoName{ID: api.RepoID(id), Name: api.RepoName(name)}
	}
	format := func(r slicedSearchResults) string {
		var b bytes.Buffer
		fmt.Fprintln(&b, "results:")
		for i, match := range r.results {
			fm, _ := match.(*result.FileMatch)
			fmt.Fprintf(&b, "	[%d] %s %s\n", i, fm.Repo.Name, fm.Path)
		}
		fmt.Fprintln(&b, "common.repos:")
		var repos []string
		for _, r := range r.common.Repos {
			repos = append(repos, string(r.Name))
		}
		sort.Strings(repos)
		for _, r := range repos {
			fmt.Fprintf(&b, "	%s\n", r)
		}
		fmt.Fprintf(&b, "resultOffset: %d\n", r.resultOffset)
		fmt.Fprintf(&b, "limitHit: %v\n", r.limitHit)
		return b.String()
	}
	sharedResult := []result.Match{
		mkFileMatch(repoName("org/repo1"), "a.go"),
		mkFileMatch(repoName("org/repo1"), "b.go"),
		mkFileMatch(repoName("org/repo1"), "c.go"),
		mkFileMatch(repoName("org/repo2"), "a.go"),
		mkFileMatch(repoName("org/repo2"), "b.go"),
		mkFileMatch(repoName("org/repo3"), "a.go"),
		mkFileMatch(repoName("org/repo4"), "a.go"),
		mkFileMatch(repoName("org/repo4"), "b.go"),
		mkFileMatch(repoName("org/repo4"), "c.go"),
		mkFileMatch(repoName("org/repo5"), "a.go"),
		mkFileMatch(repoName("org/repo5"), "b.go"),
		mkFileMatch(repoName("org/repo5"), "c.go"),
		mkFileMatch(repoName("org/repo5"), "d.go"),
		mkFileMatch(repoName("org/repo5"), "e.go"),
	}
	sharedCommon := &streaming.Stats{
		// Note: this is an intentionally unordered list to ensure we do not
		// rely on the order of lists in common (which is not guaranteed by
		// tests).
		Repos: reposMap(repoName("org/repo1"), repoName("org/repo3"), repoName("org/repo2")),
	}
	tests := []struct {
		name          string
		results       []result.Match
		common        *streaming.Stats
		offset, limit int
		want          slicedSearchResults
	}{
		{
			name:    "empty result set",
			results: []result.Match{},
			common:  &streaming.Stats{},
			offset:  0,
			limit:   3,
			want: slicedSearchResults{
				results: []result.Match{},
				common: &streaming.Stats{
					Repos: nil,
				},
				resultOffset: 0,
				limitHit:     false,
			},
		},
		{
			name:    "limit repo boundary",
			results: sharedResult,
			common:  sharedCommon,
			offset:  0,
			limit:   3,
			want: slicedSearchResults{
				results: []result.Match{
					mkFileMatch(repoName("org/repo1"), "a.go"),
					mkFileMatch(repoName("org/repo1"), "b.go"),
					mkFileMatch(repoName("org/repo1"), "c.go"),
				},
				common: &streaming.Stats{
					Repos: reposMap(repoName("org/repo1")),
				},
				resultOffset: 0,
				limitHit:     true,
			},
		},
		{
			name:    "limit non repo boundary",
			results: sharedResult,
			common:  sharedCommon,
			offset:  0,
			limit:   2,
			want: slicedSearchResults{
				results: []result.Match{
					mkFileMatch(repoName("org/repo1"), "a.go"),
					mkFileMatch(repoName("org/repo1"), "b.go"),
				},
				common: &streaming.Stats{
					Repos: reposMap(repoName("org/repo1")),
				},
				resultOffset: 2,
				limitHit:     true,
			},
		},
		{
			name:    "offset repo boundary",
			results: sharedResult,
			common:  sharedCommon,
			offset:  3,
			limit:   3,
			want: slicedSearchResults{
				results: []result.Match{
					mkFileMatch(repoName("org/repo2"), "a.go"),
					mkFileMatch(repoName("org/repo2"), "b.go"),
					mkFileMatch(repoName("org/repo3"), "a.go"),
				},
				common: &streaming.Stats{
					Repos: reposMap(repoName("org/repo2"), repoName("org/repo3")),
				},
				resultOffset: 0,
				limitHit:     true,
			},
		},
		{
			name:    "offset non-repo boundary",
			results: sharedResult,
			common:  sharedCommon,
			offset:  2,
			limit:   3,
			want: slicedSearchResults{
				results: []result.Match{
					mkFileMatch(repoName("org/repo1"), "c.go"),
					mkFileMatch(repoName("org/repo2"), "a.go"),
					mkFileMatch(repoName("org/repo2"), "b.go"),
				},
				common: &streaming.Stats{
					Repos: reposMap(repoName("org/repo1"), repoName("org/repo2")),
				},
				resultOffset: 0,
				limitHit:     true,
			},
		},
		{
			name: "offset repo boundary fully consumed",
			results: []result.Match{
				mkFileMatch(repoName("org/repo1"), "a.go"),
				mkFileMatch(repoName("org/repo1"), "b.go"),
				mkFileMatch(repoName("org/repo1"), "c.go"),
				mkFileMatch(repoName("org/repo2"), "a.go"),
				mkFileMatch(repoName("org/repo2"), "b.go"),
				mkFileMatch(repoName("org/repo2"), "c.go"),
			},
			common: &streaming.Stats{
				Repos: reposMap(repoName("org/repo1"), repoName("org/repo2")),
			},
			offset: 3,
			limit:  3,
			want: slicedSearchResults{
				results: []result.Match{
					mkFileMatch(repoName("org/repo2"), "a.go"),
					mkFileMatch(repoName("org/repo2"), "b.go"),
					mkFileMatch(repoName("org/repo2"), "c.go"),
				},
				common: &streaming.Stats{
					Repos: reposMap(repoName("org/repo2")),
				},
				resultOffset: 0,
				limitHit:     false,
			},
		},
		{
			name:    "limit non-repo boundary small",
			results: sharedResult,
			common:  sharedCommon,
			offset:  1,
			limit:   1,
			want: slicedSearchResults{
				results: []result.Match{
					mkFileMatch(repoName("org/repo1"), "b.go"),
				},
				common: &streaming.Stats{
					Repos: reposMap(repoName("org/repo1")),
				},
				resultOffset: 2,
				limitHit:     true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := sliceSearchResults(test.results, test.common, test.offset, test.limit)
			if diff := cmp.Diff(format(test.want), format(got)); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSearchPagination_repoPaginationPlan(t *testing.T) {
	db := new(dbtesting.MockDB)

	revs := func(rev ...string) (revs []search.RevisionSpecifier) {
		for _, r := range rev {
			revs = append(revs, search.RevisionSpecifier{RevSpec: r})
		}
		return revs
	}
	repoName := func(name string) types.RepoName {
		// Backcompat extract ID from name.
		id := name[len(name)-1] - '0'
		return types.RepoName{ID: api.RepoID(id), Name: api.RepoName(name)}
	}
	matchResult := func(repo types.RepoName, path, rev string) *result.FileMatch {
		fm := mkFileMatch(repo, path)
		fm.InputRev = &rev
		return fm
	}
	repoRevs := func(name string, rev ...string) *search.RepositoryRevisions {
		return &search.RepositoryRevisions{
			Repo: repoName(name),
			Revs: revs(rev...),
		}
	}
	searchRepos := []*search.RepositoryRevisions{
		repoRevs("1", "master"),
		repoRevs("2", "master"),
		repoRevs("3", "master", "feature"),
		repoRevs("4", "master"),
		repoRevs("5", "master"),
	}
	var searchedBatches [][]*search.RepositoryRevisions
	resultsExecutor := func(batch []*search.RepositoryRevisions) (results []result.Match, common *streaming.Stats, err error) {
		searchedBatches = append(searchedBatches, batch)
		common = &streaming.Stats{Repos: reposMap()}
		for _, repoRev := range batch {
			for _, rev := range repoRev.Revs {
				rev := rev.RevSpec
				for i := 0; i < 3; i++ {
					results = append(results, matchResult(repoRev.Repo, fmt.Sprintf("some/file%d.go", i), rev))
				}
			}
			common.Repos[repoRev.Repo.ID] = repoRev.Repo
		}
		return
	}
	noResultsExecutor := func(batch []*search.RepositoryRevisions) (results []result.Match, common *streaming.Stats, err error) {
		return nil, &streaming.Stats{}, nil
	}
	ctx := context.Background()

	tests := []struct {
		name                string
		executor            executor
		request             *run.SearchPaginationInfo
		wantSearchedBatches [][]*search.RepositoryRevisions
		wantCursor          *run.SearchCursor
		wantResults         []result.Match
		wantCommon          *streaming.Stats
		wantErr             error
	}{
		{
			name: "first request",
			request: &run.SearchPaginationInfo{
				Cursor: &run.SearchCursor{},
				Limit:  10,
			},
			wantSearchedBatches: [][]*search.RepositoryRevisions{
				{
					repoRevs("1", "master"),
					repoRevs("2", "master"),
					repoRevs("3", "master", "feature"),
					repoRevs("4", "master"),
				},
			},
			wantCursor: &run.SearchCursor{RepositoryOffset: 2, ResultOffset: 4},
			wantResults: []result.Match{
				matchResult(repoName("1"), "some/file0.go", "master"),
				matchResult(repoName("1"), "some/file1.go", "master"),
				matchResult(repoName("1"), "some/file2.go", "master"),
				matchResult(repoName("2"), "some/file0.go", "master"),
				matchResult(repoName("2"), "some/file1.go", "master"),
				matchResult(repoName("2"), "some/file2.go", "master"),
				matchResult(repoName("3"), "some/file0.go", "master"),
				matchResult(repoName("3"), "some/file1.go", "master"),
				matchResult(repoName("3"), "some/file2.go", "master"),
				matchResult(repoName("3"), "some/file0.go", "feature"),
			},
			wantCommon: &streaming.Stats{
				Repos: reposMap(repoName("1"), repoName("2"), repoName("3")),
			},
		},
		{
			name: "second request",
			request: &run.SearchPaginationInfo{
				Cursor: &run.SearchCursor{RepositoryOffset: 2, ResultOffset: 4},
				Limit:  10,
			},
			wantSearchedBatches: [][]*search.RepositoryRevisions{
				{
					repoRevs("3", "master", "feature"),
					repoRevs("4", "master"),
					repoRevs("5", "master"),
				},
			},
			wantCursor: &run.SearchCursor{RepositoryOffset: 5, ResultOffset: 0, Finished: true},
			wantResults: []result.Match{
				matchResult(repoName("3"), "some/file1.go", "feature"),
				matchResult(repoName("3"), "some/file2.go", "feature"),
				matchResult(repoName("4"), "some/file0.go", "master"),
				matchResult(repoName("4"), "some/file1.go", "master"),
				matchResult(repoName("4"), "some/file2.go", "master"),
				matchResult(repoName("5"), "some/file0.go", "master"),
				matchResult(repoName("5"), "some/file1.go", "master"),
				matchResult(repoName("5"), "some/file2.go", "master"),
			},
			wantCommon: &streaming.Stats{
				Repos: reposMap(repoName("3"), repoName("4"), repoName("5")),
			},
		},
		{
			name: "small limit, first request",
			request: &run.SearchPaginationInfo{
				Cursor: &run.SearchCursor{},
				Limit:  1,
			},
			wantSearchedBatches: [][]*search.RepositoryRevisions{
				{
					repoRevs("1", "master"),
					repoRevs("2", "master"),
					repoRevs("3", "master", "feature"),
					repoRevs("4", "master"),
				},
			},
			wantCursor: &run.SearchCursor{RepositoryOffset: 0, ResultOffset: 1},
			wantResults: []result.Match{
				matchResult(repoName("1"), "some/file0.go", "master"),
			},
			wantCommon: &streaming.Stats{
				Repos: reposMap(repoName("1")),
			},
		},
		{
			name: "small limit, second request",
			request: &run.SearchPaginationInfo{
				Cursor: &run.SearchCursor{RepositoryOffset: 0, ResultOffset: 1},
				Limit:  1,
			},
			wantSearchedBatches: [][]*search.RepositoryRevisions{
				{
					repoRevs("1", "master"),
					repoRevs("2", "master"),
					repoRevs("3", "master", "feature"),
					repoRevs("4", "master"),
				},
			},
			wantCursor: &run.SearchCursor{RepositoryOffset: 0, ResultOffset: 2},
			wantResults: []result.Match{
				matchResult(repoName("1"), "some/file1.go", "master"),
			},
			wantCommon: &streaming.Stats{
				Repos: reposMap(repoName("1")),
			},
		},
		{
			name:     "no results",
			executor: noResultsExecutor,
			request: &run.SearchPaginationInfo{
				Cursor: &run.SearchCursor{},
				Limit:  1,
			},
			wantCursor: &run.SearchCursor{RepositoryOffset: 1, ResultOffset: 0, Finished: true},
			wantCommon: &streaming.Stats{
				Repos: reposMap(),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			searchedBatches = nil
			plan := &repoPaginationPlan{
				pagination:          test.request,
				repositories:        searchRepos,
				searchBucketDivisor: 8,
				searchBucketMin:     4,
				searchBucketMax:     10,
				mockNumTotalRepos:   func() int { return len(searchRepos) },
			}
			executor := resultsExecutor
			if test.executor != nil {
				executor = test.executor
			}
			cursor, results, common, err := plan.execute(ctx, database.Repos(db), executor)
			if !cmp.Equal(test.wantCursor, cursor) {
				t.Error("wantCursor != cursor", cmp.Diff(test.wantCursor, cursor))
			}
			if !cmp.Equal(test.wantResults, results) {
				t.Error("wantResults != results", cmp.Diff(test.wantResults, results))
			}
			if diff := cmp.Diff(test.wantCommon, common, cmpopts.EquateEmpty()); diff != "" {
				t.Error("wantCommon != common", diff)
			}
			if !cmp.Equal(test.wantErr, err) {
				t.Error("wantErr != err", cmp.Diff(test.wantErr, err))
			}
			if !cmp.Equal(test.wantSearchedBatches, searchedBatches) {
				t.Error("wantSearchedBatches != searchedBatches", cmp.Diff(test.wantSearchedBatches, searchedBatches))
			}
		})
	}
}

func TestSearchPagination_issue_6287(t *testing.T) {
	db := new(dbtesting.MockDB)

	revs := func(rev ...string) (revs []search.RevisionSpecifier) {
		for _, r := range rev {
			revs = append(revs, search.RevisionSpecifier{RevSpec: r})
		}
		return revs
	}
	repoName := func(name string) types.RepoName {
		// Backcompat extract ID from name.
		id := name[len(name)-1] - '0'
		return types.RepoName{ID: api.RepoID(id), Name: api.RepoName(name)}
	}
	repoRevs := func(name string, rev ...string) *search.RepositoryRevisions {
		return &search.RepositoryRevisions{
			Repo: repoName(name),
			Revs: revs(rev...),
		}
	}
	repoResults := map[string][]result.Match{
		"1": {
			mkFileMatch(repoName("1"), "a.go"),
			mkFileMatch(repoName("1"), "b.go"),
		},
		"2": {
			mkFileMatch(repoName("2"), "a.go"),
			mkFileMatch(repoName("2"), "b.go"),
			mkFileMatch(repoName("2"), "c.go"),
			mkFileMatch(repoName("2"), "d.go"),
			mkFileMatch(repoName("2"), "e.go"),
		},
	}
	searchRepos := []*search.RepositoryRevisions{
		repoRevs("1", "master"),
		repoRevs("2", "master"),
	}
	executor := func(batch []*search.RepositoryRevisions) (results []result.Match, common *streaming.Stats, err error) {
		common = &streaming.Stats{Repos: reposMap()}
		for _, repoRev := range batch {
			results = append(results, repoResults[string(repoRev.Repo.Name)]...)
			common.Repos[repoRev.Repo.ID] = repoRev.Repo
		}
		return
	}
	ctx := context.Background()

	tests := []struct {
		name        string
		request     *run.SearchPaginationInfo
		wantCursor  *run.SearchCursor
		wantResults []result.Match
		wantErr     error
	}{
		{
			name: "request 1",
			request: &run.SearchPaginationInfo{
				Cursor: &run.SearchCursor{},
				Limit:  3,
			},
			wantCursor: &run.SearchCursor{RepositoryOffset: 1, ResultOffset: 1},
			wantResults: []result.Match{
				mkFileMatch(repoName("1"), "a.go"),
				mkFileMatch(repoName("1"), "b.go"),
				mkFileMatch(repoName("2"), "a.go"),
			},
		},
		{
			name: "request 2",
			request: &run.SearchPaginationInfo{
				Cursor: &run.SearchCursor{RepositoryOffset: 1, ResultOffset: 1},
				Limit:  3,
			},
			wantCursor: &run.SearchCursor{RepositoryOffset: 1, ResultOffset: 4},
			wantResults: []result.Match{
				mkFileMatch(repoName("2"), "b.go"),
				mkFileMatch(repoName("2"), "c.go"),
				mkFileMatch(repoName("2"), "d.go"),
			},
		},
		{
			name: "request 3",
			request: &run.SearchPaginationInfo{
				Cursor: &run.SearchCursor{RepositoryOffset: 1, ResultOffset: 4},
				Limit:  3,
			},
			wantCursor: &run.SearchCursor{RepositoryOffset: 2, ResultOffset: 0, Finished: true},
			wantResults: []result.Match{
				mkFileMatch(repoName("2"), "e.go"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := &repoPaginationPlan{
				pagination:          test.request,
				repositories:        searchRepos,
				searchBucketDivisor: 8,
				searchBucketMin:     4,
				searchBucketMax:     10,
				mockNumTotalRepos:   func() int { return len(searchRepos) },
			}
			cursor, results, _, err := plan.execute(ctx, database.Repos(db), executor)
			if !cmp.Equal(test.wantCursor, cursor) {
				t.Error("wantCursor != cursor", cmp.Diff(test.wantCursor, cursor))
			}
			if !cmp.Equal(test.wantResults, results) {
				t.Error("wantResults != results", cmp.Diff(test.wantResults, results))
			}
			if !cmp.Equal(test.wantErr, err) {
				t.Error("wantErr != err", cmp.Diff(test.wantErr, err))
			}
		})
	}
}

// TestSearchPagination_cloning_missing is a joint test for both
// repoPaginationPlan and sliceSearchResults's handling of cloning and missing
// repositories.
func TestSearchPagination_cloning_missing(t *testing.T) {
	db := new(dbtesting.MockDB)

	revs := func(rev ...string) (revs []search.RevisionSpecifier) {
		for _, r := range rev {
			revs = append(revs, search.RevisionSpecifier{RevSpec: r})
		}
		return revs
	}
	repoName := func(name string) types.RepoName {
		// Backcompat extract ID from name.
		id := name[len(name)-1] - 'a' + 1
		return types.RepoName{ID: api.RepoID(id), Name: api.RepoName(name)}
	}
	repoRevs := func(name string, rev ...string) *search.RepositoryRevisions {
		return &search.RepositoryRevisions{
			Repo: repoName(name),
			Revs: revs(rev...),
		}
	}
	repoResults := map[string][]result.Match{
		"a": {
			mkFileMatch(repoName("a"), "a.go"),
		},
		"c": {
			mkFileMatch(repoName("c"), "a.go"),
		},
		"f": {
			mkFileMatch(repoName("f"), "a.go"),
		},
	}
	reposStatus := func(m map[string]search.RepoStatus) search.RepoStatusMap {
		var rsm search.RepoStatusMap
		for name, status := range m {
			rsm.Update(repoName(name).ID, status)
		}
		return rsm
	}
	status := reposStatus(map[string]search.RepoStatus{
		"b": search.RepoStatusMissing,
		"e": search.RepoStatusMissing,
		"d": search.RepoStatusCloning,
	})
	searchRepos := []*search.RepositoryRevisions{
		repoRevs("a", "master"),
		repoRevs("b", "master"),
		repoRevs("c", "master"),
		repoRevs("d", "master"),
		repoRevs("e", "master"),
		repoRevs("f", "master"),
	}
	executor := func(batch []*search.RepositoryRevisions) (results []result.Match, common *streaming.Stats, err error) {
		common = &streaming.Stats{Repos: reposMap()}
		for _, repoRev := range batch {
			if res, ok := repoResults[string(repoRev.Repo.Name)]; ok {
				results = append(results, res...)
			}
			if mask := status.Get(repoRev.Repo.ID); mask != 0 {
				common.Status.Update(repoRev.Repo.ID, mask)
			}
		}
		return
	}
	ctx := context.Background()

	tests := []struct {
		name        string
		request     *run.SearchPaginationInfo
		searchRepos []*search.RepositoryRevisions
		wantCursor  *run.SearchCursor
		wantResults []result.Match
		wantCommon  *streaming.Stats
		wantErr     error
	}{
		{
			name: "repo a",
			request: &run.SearchPaginationInfo{
				Cursor: &run.SearchCursor{},
				Limit:  1,
			},
			wantCursor: &run.SearchCursor{RepositoryOffset: 1, ResultOffset: 0},
			wantResults: []result.Match{
				mkFileMatch(repoName("a"), "a.go"),
			},
			wantCommon: &streaming.Stats{
				Repos: reposMap(repoName("a")),
			},
		},
		{
			name: "missing repo b, repo c",
			request: &run.SearchPaginationInfo{
				Cursor: &run.SearchCursor{RepositoryOffset: 1, ResultOffset: 0},
				Limit:  1,
			},
			wantCursor: &run.SearchCursor{RepositoryOffset: 3, ResultOffset: 0},
			wantResults: []result.Match{
				mkFileMatch(repoName("c"), "a.go"),
			},
			wantCommon: &streaming.Stats{

				Repos: reposMap(repoName("b"), repoName("c")),
				Status: reposStatus(map[string]search.RepoStatus{
					"b": search.RepoStatusMissing,
				}),
			},
		},
		{
			name: "repo a, missing repo b, repo c",
			request: &run.SearchPaginationInfo{
				Cursor: &run.SearchCursor{},
				Limit:  2,
			},
			wantCursor: &run.SearchCursor{RepositoryOffset: 3, ResultOffset: 0},
			wantResults: []result.Match{
				mkFileMatch(repoName("a"), "a.go"),
				mkFileMatch(repoName("c"), "a.go"),
			},
			wantCommon: &streaming.Stats{
				Repos: reposMap(repoName("a"), repoName("b"), repoName("c")),
				Status: reposStatus(map[string]search.RepoStatus{
					"b": search.RepoStatusMissing,
				}),
			},
		},
		{
			name: "all",
			request: &run.SearchPaginationInfo{
				Cursor: &run.SearchCursor{},
				Limit:  3,
			},
			wantCursor: &run.SearchCursor{RepositoryOffset: 6, ResultOffset: 0, Finished: true},
			wantResults: []result.Match{
				mkFileMatch(repoName("a"), "a.go"),
				mkFileMatch(repoName("c"), "a.go"),
				mkFileMatch(repoName("f"), "a.go"),
			},
			wantCommon: &streaming.Stats{
				Repos: reposMap(repoName("a"), repoName("b"), repoName("c"), repoName("d"), repoName("e"), repoName("f")),
				Status: reposStatus(map[string]search.RepoStatus{
					"b": search.RepoStatusMissing,
					"d": search.RepoStatusCloning,
					"e": search.RepoStatusMissing,
				}),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := &repoPaginationPlan{
				pagination:          test.request,
				repositories:        searchRepos,
				searchBucketDivisor: 8,
				searchBucketMin:     4,
				searchBucketMax:     10,
				mockNumTotalRepos:   func() int { return len(test.searchRepos) },
			}
			cursor, results, common, err := plan.execute(ctx, database.Repos(db), executor)
			if !cmp.Equal(test.wantCursor, cursor) {
				t.Error("wantCursor != cursor", cmp.Diff(test.wantCursor, cursor))
			}
			if !cmp.Equal(test.wantResults, results) {
				t.Error("wantResults != results", cmp.Diff(test.wantResults, results))
			}
			if !cmp.Equal(test.wantCommon, common) {
				t.Error("common mismatch (-want +got):\n", cmp.Diff(test.wantCommon, common))
			}
			if !cmp.Equal(test.wantErr, err) {
				t.Error("wantErr != err", cmp.Diff(test.wantErr, err))
			}
		})
	}
}

func reposMap(repos ...types.RepoName) map[api.RepoID]types.RepoName {
	m := make(map[api.RepoID]types.RepoName, len(repos))
	for _, r := range repos {
		m[r.ID] = r
	}
	return m
}
