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

func TestService_ResolveRepositoriesMatchingQuery(t *testing.T) {
	ctx := context.Background()

	db := dbtest.NewDB(t, "")
	s := store.New(db, &observation.TestContext, nil)

	rs, _ := ct.CreateTestRepos(t, ctx, db, 4)

	defaultBranches := map[api.RepoName]defaultBranch{
		rs[0].Name: {branch: "branch-1", commit: api.CommitID("6f152ece24b9424edcd4da2b82989c5c2bea64c3")},
		rs[1].Name: {branch: "branch-2", commit: api.CommitID("2840a42c7809c22b16fda7099c725d1ef197961c")},
		rs[2].Name: {branch: "branch-3", commit: api.CommitID("aead85d33485e115b33ec4045c55bac97e03fd26")},
		rs[3].Name: {branch: "branch-4", commit: api.CommitID("26ac0350471daac3401a9314fd64e714370837a6")},
	}
	buildRepoRev := func(repo *types.Repo) *RepoRevision {
		return &RepoRevision{Repo: repo, Branch: defaultBranches[repo.Name].branch, Commit: defaultBranches[repo.Name].commit}
	}

	mockDefaultBranches(t, defaultBranches)
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
		&streamhttp.EventPathMatch{
			Type:         streamhttp.PathMatchType,
			RepositoryID: int32(rs[2].ID),
		},
		&streamhttp.EventSymbolMatch{
			Type:         streamhttp.SymbolMatchType,
			RepositoryID: int32(rs[3].ID),
		},
	}

	tsURL := newStreamSearchTestServer(t, searchMatches)

	batchSpec := &batcheslib.BatchSpec{
		On: []batcheslib.OnQueryOrRepository{
			{RepositoriesMatchingQuery: "repohasfile:horse.txt"},
		},
		Workspaces: []batcheslib.WorkspaceConfiguration{},
	}

	wr := &workspaceResolver{store: s, frontendInternalURL: tsURL}

	repoRevs, err := wr.ResolveRepositoriesForBatchSpec(ctx, batchSpec, ResolveRepositoriesForBatchSpecOpts{
		AllowIgnored:     false,
		AllowUnsupported: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if want, have := 4, len(repoRevs); want != have {
		t.Fatalf("wrong number of repositories returned. want=%d, have=%d", want, have)
	}

	want := []*RepoRevision{
		buildRepoRev(rs[0]),
		buildRepoRev(rs[1]),
		buildRepoRev(rs[2]),
		buildRepoRev(rs[3]),
	}

	sort.Slice(want, func(i, j int) bool { return want[i].Repo.ID < want[j].Repo.ID })
	sort.Slice(repoRevs, func(i, j int) bool { return repoRevs[i].Repo.ID < repoRevs[j].Repo.ID })

	if diff := cmp.Diff(want, repoRevs); diff != "" {
		t.Fatalf("returned repoRevisions wrong. (-want +got):\n%s", diff)
	}
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
