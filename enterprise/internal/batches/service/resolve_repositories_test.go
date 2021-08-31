package service

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/internal/vcs/util"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

func TestSetDefaultQueryCount(t *testing.T) {
	for in, want := range map[string]string{
		"":                     hardCodedCount,
		"count:10":             "count:10",
		"count:all":            "count:all",
		"r:foo":                "r:foo" + hardCodedCount,
		"r:foo count:10":       "r:foo count:10",
		"r:foo count:10 f:bar": "r:foo count:10 f:bar",
		"r:foo count:":         "r:foo count:" + hardCodedCount,
		"r:foo count:xyz":      "r:foo count:xyz" + hardCodedCount,
	} {
		t.Run(in, func(t *testing.T) {
			have := setDefaultQueryCount(in)
			if have != want {
				t.Errorf("unexpected query: have %q; want %q", have, want)
			}
		})
	}
}

func TestSetDefaultQuerySelect(t *testing.T) {
	for in, want := range map[string]string{
		"":                        hardCodedSelectRepo,
		"select:file":             "select:file",
		"select:repo":             "select:repo",
		"r:foo":                   "r:foo" + hardCodedSelectRepo,
		"r:foo select:file":       "r:foo select:file",
		"r:foo select:file f:bar": "r:foo select:file f:bar",
		"r:foo select:":           "r:foo select:" + hardCodedSelectRepo,
		"r:foo select:xyz":        "r:foo select:xyz",
	} {
		t.Run(in, func(t *testing.T) {
			have := setDefaultQuerySelect(in)
			if have != want {
				t.Errorf("unexpected query: have %q; want %q", have, want)
			}
		})
	}
}

func TestService_ResolveRepositoriesForBatchSpec(t *testing.T) {
	ctx := context.Background()

	db := dbtest.NewDB(t, "")
	s := store.New(db, &observation.TestContext, nil)

	rs, _ := ct.CreateTestRepos(t, ctx, db, 4)

	unsupported, _ := ct.CreateAWSCodeCommitTestRepos(t, ctx, db, 1)

	defaultBranches := map[api.RepoName]defaultBranch{
		rs[0].Name:          {branch: "branch-1", commit: api.CommitID("6f152ece24b9424edcd4da2b82989c5c2bea64c3")},
		rs[1].Name:          {branch: "branch-2", commit: api.CommitID("2840a42c7809c22b16fda7099c725d1ef197961c")},
		rs[2].Name:          {branch: "branch-3", commit: api.CommitID("aead85d33485e115b33ec4045c55bac97e03fd26")},
		rs[3].Name:          {branch: "branch-4", commit: api.CommitID("26ac0350471daac3401a9314fd64e714370837a6")},
		unsupported[0].Name: {branch: "branch-5", commit: api.CommitID("c167bd633e2868585b86ef129d07f63dee46b84a")},
	}
	buildRepoRev := func(repo *types.Repo) *RepoRevision {
		return &RepoRevision{Repo: repo, Branch: defaultBranches[repo.Name].branch, Commit: defaultBranches[repo.Name].commit}
	}
	mockDefaultBranches(t, defaultBranches)

	defaultOpts := ResolveRepositoriesForBatchSpecOpts{
		AllowIgnored:     false,
		AllowUnsupported: false,
	}

	t.Run("repositoriesMatchingQuery", func(t *testing.T) {
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				{RepositoriesMatchingQuery: "repohasfile:horse.txt"},
				// In our test the search API returns the same results for both
				{RepositoriesMatchingQuery: "repohasfile:horse.txt duplicate"},
			},
		}

		mockBatchIgnores(t, map[api.CommitID]bool{
			defaultBranches[rs[0].Name].commit: false,
			defaultBranches[rs[1].Name].commit: true,
			defaultBranches[rs[2].Name].commit: true,
			defaultBranches[rs[3].Name].commit: false,
		})

		searchMatches := []streamhttp.EventMatch{
			&streamhttp.EventContentMatch{
				Type:         streamhttp.ContentMatchType,
				Path:         "test",
				RepositoryID: int32(rs[0].ID),
			},
			&streamhttp.EventContentMatch{
				Type:         streamhttp.ContentMatchType,
				Path:         "duplicate-test",
				RepositoryID: int32(rs[0].ID),
			},
			&streamhttp.EventRepoMatch{
				Type:         streamhttp.RepoMatchType,
				RepositoryID: int32(rs[1].ID),
			},
			&streamhttp.EventPathMatch{
				Type:         streamhttp.PathMatchType,
				RepositoryID: int32(rs[2].ID),
			},
			&streamhttp.EventSymbolMatch{
				Type:         streamhttp.SymbolMatchType,
				RepositoryID: int32(rs[3].ID),
			},
		}

		want := []*RepoRevision{buildRepoRev(rs[0]), buildRepoRev(rs[3])}
		wantIgnored := []api.RepoID{rs[1].ID, rs[2].ID}
		wantUnsupported := []api.RepoID{}
		resolveRepoRevsAndCompare(t, s, defaultOpts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)
	})

	t.Run("repositories", func(t *testing.T) {
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				{Repository: string(rs[0].Name)},
				{Repository: string(rs[1].Name), Branch: "non-default-branch"},
				{Repository: string(rs[2].Name), Branch: "other-non-default-branch"},
				{Repository: string(rs[3].Name)},
			},
		}

		mockResolveRevision(t, map[string]api.CommitID{
			defaultBranches[rs[0].Name].branch: defaultBranches[rs[0].Name].commit,
			"non-default-branch":               api.CommitID("d34db33f"),
			"other-non-default-branch":         api.CommitID("c0ff33"),
			defaultBranches[rs[3].Name].branch: defaultBranches[rs[3].Name].commit,
		})

		mockBatchIgnores(t, map[api.CommitID]bool{
			defaultBranches[rs[0].Name].commit: false,
			api.CommitID("d34db33f"):           false,
			api.CommitID("c0ff33"):             false,
			defaultBranches[rs[3].Name].commit: true,
		})

		searchMatches := []streamhttp.EventMatch{}

		want := []*RepoRevision{
			buildRepoRev(rs[0]),
			{Repo: rs[1], Branch: "non-default-branch", Commit: api.CommitID("d34db33f")},
			{Repo: rs[2], Branch: "other-non-default-branch", Commit: api.CommitID("c0ff33")},
		}

		wantIgnored := []api.RepoID{rs[3].ID}
		wantUnsupported := []api.RepoID{}
		resolveRepoRevsAndCompare(t, s, defaultOpts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)
	})

	t.Run("repositoriesMatchingQuery and repositories", func(t *testing.T) {
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				{RepositoriesMatchingQuery: "repohasfile:horse.txt"},
				{Repository: string(rs[2].Name)},
				{Repository: string(rs[3].Name)},
			},
		}

		mockBatchIgnores(t, map[api.CommitID]bool{
			defaultBranches[rs[0].Name].commit: false,
			defaultBranches[rs[1].Name].commit: false,
			defaultBranches[rs[2].Name].commit: false,
			defaultBranches[rs[3].Name].commit: false,
		})

		searchMatches := []streamhttp.EventMatch{
			&streamhttp.EventContentMatch{
				Type:         streamhttp.ContentMatchType,
				Path:         "test",
				RepositoryID: int32(rs[0].ID),
			},
			&streamhttp.EventRepoMatch{
				Type:         streamhttp.RepoMatchType,
				RepositoryID: int32(rs[1].ID),
			},
			// Included in the search results and explicitly listed
			&streamhttp.EventRepoMatch{
				Type:         streamhttp.RepoMatchType,
				RepositoryID: int32(rs[2].ID),
			},
		}

		mockResolveRevision(t, map[string]api.CommitID{
			defaultBranches[rs[2].Name].branch: defaultBranches[rs[2].Name].commit,
			defaultBranches[rs[3].Name].branch: defaultBranches[rs[3].Name].commit,
		})

		mockBatchIgnores(t, map[api.CommitID]bool{
			defaultBranches[rs[0].Name].commit: false,
			defaultBranches[rs[1].Name].commit: false,
			defaultBranches[rs[2].Name].commit: false,
			defaultBranches[rs[3].Name].commit: false,
		})

		want := []*RepoRevision{
			buildRepoRev(rs[0]),
			buildRepoRev(rs[1]),
			buildRepoRev(rs[2]),
			buildRepoRev(rs[3]),
		}

		wantIgnored := []api.RepoID{}
		wantUnsupported := []api.RepoID{}
		resolveRepoRevsAndCompare(t, s, defaultOpts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)
	})

	t.Run("allowUnsupported option", func(t *testing.T) {
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				{RepositoriesMatchingQuery: "repohasfile:horse.txt"},
			},
		}

		mockBatchIgnores(t, map[api.CommitID]bool{
			defaultBranches[unsupported[0].Name].commit: false,
		})

		searchMatches := []streamhttp.EventMatch{
			&streamhttp.EventContentMatch{
				Type:         streamhttp.ContentMatchType,
				Path:         "test",
				RepositoryID: int32(unsupported[0].ID),
			},
		}

		want := []*RepoRevision{}
		wantIgnored := []api.RepoID{}
		wantUnsupported := []api.RepoID{unsupported[0].ID}
		resolveRepoRevsAndCompare(t, s, defaultOpts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)

		// with allowUnsupported: true
		opts := defaultOpts
		opts.AllowUnsupported = true

		want = []*RepoRevision{buildRepoRev(unsupported[0])}
		wantUnsupported = []api.RepoID{}
		resolveRepoRevsAndCompare(t, s, opts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)
	})

	t.Run("allowIgnored option", func(t *testing.T) {
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				{Repository: string(rs[0].Name)},
			},
		}

		mockResolveRevision(t, map[string]api.CommitID{
			defaultBranches[rs[0].Name].branch: defaultBranches[rs[0].Name].commit,
		})

		mockBatchIgnores(t, map[api.CommitID]bool{
			defaultBranches[rs[0].Name].commit: true,
		})

		searchMatches := []streamhttp.EventMatch{}

		want := []*RepoRevision{}
		wantIgnored := []api.RepoID{rs[0].ID}
		wantUnsupported := []api.RepoID{}
		resolveRepoRevsAndCompare(t, s, defaultOpts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)

		// with allowIgnored: true
		opts := defaultOpts
		opts.AllowIgnored = true

		want = []*RepoRevision{buildRepoRev(rs[0])}
		wantIgnored = []api.RepoID{}
		resolveRepoRevsAndCompare(t, s, opts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)
	})
}

func resolveRepoRevsAndCompare(t *testing.T, s *store.Store, opts ResolveRepositoriesForBatchSpecOpts, matches []streamhttp.EventMatch, spec *batcheslib.BatchSpec, want []*RepoRevision, wantIgnored, wantUnsupported []api.RepoID) {
	t.Helper()

	wr := &workspaceResolver{
		store:               s,
		frontendInternalURL: newStreamSearchTestServer(t, matches),
	}
	have, err := wr.ResolveRepositoriesForBatchSpec(context.Background(), spec, opts)
	if len(wantIgnored) > 0 {
		set, ok := err.(IgnoredRepoSet)
		if !ok {
			t.Fatalf("unexpected error: %s", err)
		}

		for _, id := range wantIgnored {
			if !set.includesRepoWithID(id) {
				t.Fatalf("IgnoredRepoSet does not contain repo with ID %d", id)
			}
		}
	} else if len(wantUnsupported) > 0 {
		set, ok := err.(UnsupportedRepoSet)
		if !ok {
			t.Fatalf("unexpected error: %s", err)
		}

		for _, id := range wantUnsupported {
			if !set.includesRepoWithID(id) {
				t.Fatalf("UnsupportedRepoSet does not contain repo with ID %d", id)
			}
		}
	} else if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	sortRepoRevs(want)
	sortRepoRevs(have)
	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("returned repoRevisions wrong. (-want +got):\n%s", diff)
	}
}

func sortRepoRevs(revs []*RepoRevision) {
	sort.Slice(revs, func(i, j int) bool { return revs[i].Repo.ID < revs[j].Repo.ID })
}

func newStreamSearchTestServer(t *testing.T, matches []streamhttp.EventMatch) string {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type ev struct {
			Name  string
			Value interface{}
		}
		ew, err := streamhttp.NewWriter(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ew.Event("progress", ev{
			Name:  "progress",
			Value: &streamapi.Progress{MatchCount: len(matches)},
		})
		ew.Event("matches", matches)
		ew.Event("done", struct{}{})
	}))

	t.Cleanup(ts.Close)

	return ts.URL
}

type defaultBranch struct {
	branch string
	commit api.CommitID
}

func mockDefaultBranches(t *testing.T, defaultBranches map[api.RepoName]defaultBranch) {
	git.Mocks.GetDefaultBranch = func(repo api.RepoName) (refName string, commit api.CommitID, err error) {
		if res, ok := defaultBranches[repo]; ok {
			return res.branch, res.commit, nil
		}
		return "", "", fmt.Errorf("unknown repo: %s", repo)
	}
	t.Cleanup(func() { git.Mocks.GetDefaultBranch = nil })
}

func mockBatchIgnores(t *testing.T, m map[api.CommitID]bool) {
	git.Mocks.Stat = func(commit api.CommitID, name string) (fs.FileInfo, error) {
		hasBatchIgnore, ok := m[commit]
		if !ok {
			return nil, fmt.Errorf("unknown commit: %s", commit)
		}
		if hasBatchIgnore {
			return &util.FileInfo{Name_: ".batchignore", Mode_: 0}, nil
		}
		return nil, os.ErrNotExist
	}
	t.Cleanup(func() { git.Mocks.Stat = nil })
}

func mockResolveRevision(t *testing.T, branches map[string]api.CommitID) {
	git.Mocks.ResolveRevision = func(spec string, opt git.ResolveRevisionOptions) (api.CommitID, error) {
		if commit, ok := branches[spec]; ok {
			return commit, nil
		}
		return "", fmt.Errorf("unknown spec: %s", spec)
	}
	t.Cleanup(func() { git.Mocks.ResolveRevision = nil })
}
