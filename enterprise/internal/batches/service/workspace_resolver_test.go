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
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
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

func TestService_ResolveWorkspacesForBatchSpec(t *testing.T) {
	ctx := context.Background()

	db := database.NewDB(dbtest.NewDB(t))
	s := store.New(db, &observation.TestContext, nil)

	u := ct.CreateTestUser(t, db, false)

	rs, _ := ct.CreateTestRepos(t, ctx, db, 5)
	unsupported, _ := ct.CreateAWSCodeCommitTestRepos(t, ctx, db, 1)
	// Allow access to all repos but rs[4].
	ct.MockRepoPermissions(t, db, u.ID, rs[0].ID, rs[1].ID, rs[2].ID, rs[3].ID, unsupported[0].ID)

	defaultBranches := map[api.RepoName]defaultBranch{
		rs[0].Name:          {branch: "branch-1", commit: api.CommitID("6f152ece24b9424edcd4da2b82989c5c2bea64c3")},
		rs[1].Name:          {branch: "branch-2", commit: api.CommitID("2840a42c7809c22b16fda7099c725d1ef197961c")},
		rs[2].Name:          {branch: "branch-3", commit: api.CommitID("aead85d33485e115b33ec4045c55bac97e03fd26")},
		rs[3].Name:          {branch: "branch-4", commit: api.CommitID("26ac0350471daac3401a9314fd64e714370837a6")},
		rs[4].Name:          {branch: "branch-6", commit: api.CommitID("010b133ece7a79187cad209b27099232485a5476")},
		unsupported[0].Name: {branch: "branch-5", commit: api.CommitID("c167bd633e2868585b86ef129d07f63dee46b84a")},
	}
	steps := []batcheslib.Step{{Run: "echo 1"}}
	buildRepoWorkspace := func(repo *types.Repo, branch, commit string, fileMatches []string) *RepoWorkspace {
		sort.Strings(fileMatches)
		if branch == "" {
			branch = defaultBranches[repo.Name].branch
		}
		if commit == "" {
			commit = string(defaultBranches[repo.Name].commit)
		}
		return &RepoWorkspace{
			RepoRevision: &RepoRevision{
				Repo:        repo,
				Branch:      branch,
				Commit:      api.CommitID(commit),
				FileMatches: fileMatches,
			},
			Path:               "",
			Steps:              steps,
			OnlyFetchWorkspace: false,
		}
	}
	buildIgnoredRepoWorkspace := func(repo *types.Repo, branch, commit string, fileMatches []string) *RepoWorkspace {
		ws := buildRepoWorkspace(repo, branch, commit, fileMatches)
		ws.Ignored = true
		return ws
	}
	buildUnsupportedRepoWorkspace := func(repo *types.Repo, branch, commit string, fileMatches []string) *RepoWorkspace {
		ws := buildRepoWorkspace(repo, branch, commit, fileMatches)
		ws.Unsupported = true
		return ws
	}

	mockDefaultBranches(t, defaultBranches)

	t.Run("repositoriesMatchingQuery", func(t *testing.T) {
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				{RepositoriesMatchingQuery: "repohasfile:horse.txt"},
				// In our test the search API returns the same results for both
				{RepositoriesMatchingQuery: "repohasfile:horse.txt duplicate"},
			},
			Steps: steps,
		}

		mockBatchIgnores(t, map[api.CommitID]bool{
			defaultBranches[rs[0].Name].commit:          false,
			defaultBranches[rs[1].Name].commit:          true,
			defaultBranches[rs[2].Name].commit:          true,
			defaultBranches[rs[3].Name].commit:          false,
			defaultBranches[unsupported[0].Name].commit: false,
		})

		searchMatches := []streamhttp.EventMatch{
			&streamhttp.EventContentMatch{
				Type:         streamhttp.ContentMatchType,
				Path:         "repo-0/test",
				RepositoryID: int32(rs[0].ID),
			},
			&streamhttp.EventContentMatch{
				Type:         streamhttp.ContentMatchType,
				Path:         "repo-0/duplicate-test",
				RepositoryID: int32(rs[0].ID),
			},
			&streamhttp.EventRepoMatch{
				Type:         streamhttp.RepoMatchType,
				RepositoryID: int32(rs[1].ID),
			},
			&streamhttp.EventPathMatch{
				Type:         streamhttp.PathMatchType,
				Path:         "repo-2/readme",
				RepositoryID: int32(rs[2].ID),
			},
			&streamhttp.EventSymbolMatch{
				Type:         streamhttp.SymbolMatchType,
				Path:         "repo-3/readme",
				RepositoryID: int32(rs[3].ID),
			},
			&streamhttp.EventPathMatch{
				Type:         streamhttp.PathMatchType,
				Path:         "unsupported/path",
				RepositoryID: int32(unsupported[0].ID),
			},
		}

		want := []*RepoWorkspace{
			buildRepoWorkspace(rs[0], "", "", []string{"repo-0/test", "repo-0/duplicate-test"}),
			buildIgnoredRepoWorkspace(rs[1], "", "", []string{}),
			buildIgnoredRepoWorkspace(rs[2], "", "", []string{"repo-2/readme"}),
			buildRepoWorkspace(rs[3], "", "", []string{"repo-3/readme"}),
			buildUnsupportedRepoWorkspace(unsupported[0], "", "", []string{"unsupported/path"}),
		}
		resolveWorkspacesAndCompare(t, s, u, searchMatches, batchSpec, want)
	})

	t.Run("repositories", func(t *testing.T) {
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				{Repository: string(rs[0].Name)},
				{Repository: string(rs[1].Name), Branch: "non-default-branch"},
				{Repository: string(rs[2].Name), Branches: []string{"other-non-default-branch", "yet-another-non-default-branch"}},
				{Repository: string(rs[3].Name)},
				{Repository: string(unsupported[0].Name)},
			},
			Steps: steps,
		}

		mockResolveRevision(t, map[string]api.CommitID{
			defaultBranches[rs[0].Name].branch:          defaultBranches[rs[0].Name].commit,
			"non-default-branch":                        api.CommitID("d34db33f"),
			"other-non-default-branch":                  api.CommitID("c0ff33"),
			"yet-another-non-default-branch":            api.CommitID("b33a"),
			defaultBranches[rs[3].Name].branch:          defaultBranches[rs[3].Name].commit,
			defaultBranches[unsupported[0].Name].branch: defaultBranches[unsupported[0].Name].commit,
		})

		mockBatchIgnores(t, map[api.CommitID]bool{
			defaultBranches[rs[0].Name].commit:          false,
			api.CommitID("d34db33f"):                    false,
			api.CommitID("c0ff33"):                      false,
			api.CommitID("b33a"):                        false,
			defaultBranches[rs[3].Name].commit:          true,
			defaultBranches[unsupported[0].Name].commit: false,
		})

		searchMatches := []streamhttp.EventMatch{}

		want := []*RepoWorkspace{
			buildRepoWorkspace(rs[0], "", "", []string{}),
			buildRepoWorkspace(rs[1], "non-default-branch", "d34db33f", []string{}),
			buildRepoWorkspace(rs[2], "other-non-default-branch", "c0ff33", []string{}),
			buildRepoWorkspace(rs[2], "yet-another-non-default-branch", "b33a", []string{}),
			buildIgnoredRepoWorkspace(rs[3], "", "", []string{}),
			buildUnsupportedRepoWorkspace(unsupported[0], "", "", []string{}),
		}

		resolveWorkspacesAndCompare(t, s, u, searchMatches, batchSpec, want)
	})

	t.Run("repositories overriding previous queries", func(t *testing.T) {
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				// This query is just a placeholder; we'll set up the search
				// results further down to return rs[2].
				{RepositoriesMatchingQuery: "r:rs-2"},
				{Repository: string(rs[0].Name)},
				{Repository: string(rs[1].Name), Branch: "non-default-branch"},
				{Repository: string(rs[1].Name), Branch: "a-different-non-default-branch"},
				{Repository: string(rs[2].Name), Branches: []string{"other-non-default-branch", "yet-another-non-default-branch"}},
				{Repository: string(rs[3].Name)},
				{Repository: string(unsupported[0].Name)},
			},
			Steps: steps,
		}

		mockResolveRevision(t, map[string]api.CommitID{
			defaultBranches[rs[0].Name].branch:          defaultBranches[rs[0].Name].commit,
			"non-default-branch":                        api.CommitID("d34db33f"),
			"a-different-non-default-branch":            api.CommitID("c4a1"),
			"other-non-default-branch":                  api.CommitID("c0ff33"),
			"yet-another-non-default-branch":            api.CommitID("b33a"),
			defaultBranches[rs[3].Name].branch:          defaultBranches[rs[3].Name].commit,
			defaultBranches[unsupported[0].Name].branch: defaultBranches[unsupported[0].Name].commit,
		})

		mockBatchIgnores(t, map[api.CommitID]bool{
			defaultBranches[rs[0].Name].commit:          false,
			api.CommitID("d34db33f"):                    false,
			api.CommitID("c4a1"):                        false,
			api.CommitID("c0ff33"):                      false,
			api.CommitID("b33a"):                        false,
			defaultBranches[rs[3].Name].commit:          true,
			defaultBranches[unsupported[0].Name].commit: false,
		})

		searchMatches := []streamhttp.EventMatch{
			&streamhttp.EventPathMatch{
				Type:         streamhttp.PathMatchType,
				Path:         "repo-2/readme",
				RepositoryID: int32(rs[2].ID),
				Branches:     []string{"main"},
			},
		}

		want := []*RepoWorkspace{
			buildRepoWorkspace(rs[0], "", "", []string{}),
			// Note that only the last rs[1] result is included.
			buildRepoWorkspace(rs[1], "a-different-non-default-branch", "c4a1", []string{}),
			// Note that this doesn't include rs[2] "main".
			buildRepoWorkspace(rs[2], "other-non-default-branch", "c0ff33", []string{}),
			buildRepoWorkspace(rs[2], "yet-another-non-default-branch", "b33a", []string{}),
			buildIgnoredRepoWorkspace(rs[3], "", "", []string{}),
			buildUnsupportedRepoWorkspace(unsupported[0], "", "", []string{}),
		}

		resolveWorkspacesAndCompare(t, s, u, searchMatches, batchSpec, want)
	})

	t.Run("repositoriesMatchingQuery and repositories", func(t *testing.T) {
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				{RepositoriesMatchingQuery: "repohasfile:horse.txt"},
				{Repository: string(rs[2].Name)},
				{Repository: string(rs[3].Name)},
			},
			Steps: steps,
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
			&streamhttp.EventRepoMatch{
				Type:         streamhttp.RepoMatchType,
				RepositoryID: int32(unsupported[0].ID),
			},
		}

		mockResolveRevision(t, map[string]api.CommitID{
			defaultBranches[rs[2].Name].branch: defaultBranches[rs[2].Name].commit,
			defaultBranches[rs[3].Name].branch: defaultBranches[rs[3].Name].commit,
		})

		mockBatchIgnores(t, map[api.CommitID]bool{
			defaultBranches[rs[0].Name].commit:          false,
			defaultBranches[rs[1].Name].commit:          false,
			defaultBranches[rs[2].Name].commit:          false,
			defaultBranches[rs[3].Name].commit:          false,
			defaultBranches[unsupported[0].Name].commit: false,
		})

		want := []*RepoWorkspace{
			buildRepoWorkspace(rs[0], "", "", []string{"test"}),
			buildRepoWorkspace(rs[1], "", "", []string{}),
			buildRepoWorkspace(rs[2], "", "", []string{}),
			buildRepoWorkspace(rs[3], "", "", []string{}),
			buildUnsupportedRepoWorkspace(unsupported[0], "", "", []string{}),
		}

		resolveWorkspacesAndCompare(t, s, u, searchMatches, batchSpec, want)
	})

	t.Run("workspaces with skipped steps", func(t *testing.T) {
		conditionalSteps := []batcheslib.Step{
			// Step should only execute in rs[1]
			{Run: "echo 1", If: fmt.Sprintf(`${{ eq repository.name %q }}`, rs[1].Name)},
		}
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				{Repository: string(rs[0].Name)},
				{Repository: string(rs[1].Name)},
			},
			Steps: conditionalSteps,
		}

		mockResolveRevision(t, map[string]api.CommitID{
			defaultBranches[rs[0].Name].branch: defaultBranches[rs[0].Name].commit,
			defaultBranches[rs[1].Name].branch: defaultBranches[rs[1].Name].commit,
		})

		mockBatchIgnores(t, map[api.CommitID]bool{
			defaultBranches[rs[0].Name].commit: false,
			defaultBranches[rs[1].Name].commit: false,
		})

		searchMatches := []streamhttp.EventMatch{}

		// We want both workspaces, but only one of them has steps that need to run
		ws0 := buildRepoWorkspace(rs[0], "", "", []string{})
		ws0.Steps = conditionalSteps
		ws0.SkippedSteps = []int32{0}
		ws1 := buildRepoWorkspace(rs[1], "", "", []string{})
		ws1.Steps = conditionalSteps

		want := []*RepoWorkspace{ws0, ws1}
		resolveWorkspacesAndCompare(t, s, u, searchMatches, batchSpec, want)
	})
}

func resolveWorkspacesAndCompare(t *testing.T, s *store.Store, u *types.User, matches []streamhttp.EventMatch, spec *batcheslib.BatchSpec, want []*RepoWorkspace) {
	t.Helper()

	wr := &workspaceResolver{
		store:               s,
		frontendInternalURL: newStreamSearchTestServer(t, matches),
	}
	ctx := actor.WithActor(context.Background(), actor.FromUser(u.ID))
	have, err := wr.ResolveWorkspacesForBatchSpec(ctx, spec)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("returned workspaces wrong. (-want +got):\n%s", diff)
	}
}

func newStreamSearchTestServer(t *testing.T, matches []streamhttp.EventMatch) string {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
	git.Mocks.Stat = func(commit api.CommitID, _ string) (fs.FileInfo, error) {
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
	git.Mocks.ResolveRevision = func(spec string, _ git.ResolveRevisionOptions) (api.CommitID, error) {
		if commit, ok := branches[spec]; ok {
			return commit, nil
		}
		return "", fmt.Errorf("unknown spec: %s", spec)
	}
	t.Cleanup(func() { git.Mocks.ResolveRevision = nil })
}

func TestFindWorkspaces(t *testing.T) {
	repoRevs := []*RepoRevision{
		{Repo: &types.Repo{ID: 1, Name: "github.com/sourcegraph/automation-testing"}},
		{Repo: &types.Repo{ID: 2, Name: "github.com/sourcegraph/sourcegraph"}},
		{Repo: &types.Repo{ID: 3, Name: "bitbucket.sgdev.org/SOUR/automation-testing"}},
	}
	steps := []batcheslib.Step{{Run: "echo 1"}}

	type finderResults map[repoRevKey][]string

	tests := map[string]struct {
		spec          *batcheslib.BatchSpec
		finderResults finderResults

		// workspaces in which repo/path they are executed
		wantWorkspaces []*RepoWorkspace
	}{
		"no workspace configuration": {
			spec:          &batcheslib.BatchSpec{Steps: steps},
			finderResults: finderResults{},
			wantWorkspaces: []*RepoWorkspace{
				{RepoRevision: repoRevs[0], Steps: steps, Path: ""},
				{RepoRevision: repoRevs[1], Steps: steps, Path: ""},
				{RepoRevision: repoRevs[2], Steps: steps, Path: ""},
			},
		},

		"workspace configuration matching no repos": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{In: "this-does-not-match", RootAtLocationOf: "package.json"},
				},
			},
			finderResults: finderResults{},
			wantWorkspaces: []*RepoWorkspace{
				{RepoRevision: repoRevs[0], Steps: steps, Path: ""},
				{RepoRevision: repoRevs[1], Steps: steps, Path: ""},
				{RepoRevision: repoRevs[2], Steps: steps, Path: ""},
			},
		},

		"workspace configuration matching 2 repos with no results": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{In: "*automation-testing", RootAtLocationOf: "package.json"},
				},
			},
			finderResults: finderResults{
				repoRevs[0].Key(): []string{},
				repoRevs[2].Key(): []string{},
			},
			wantWorkspaces: []*RepoWorkspace{
				{RepoRevision: repoRevs[1], Steps: steps, Path: ""},
			},
		},

		"workspace configuration matching 2 repos with 3 results each": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{In: "*automation-testing", RootAtLocationOf: "package.json"},
				},
			},
			finderResults: finderResults{
				repoRevs[0].Key(): {"a/b", "a/b/c", "d/e/f"},
				repoRevs[2].Key(): {"a/b", "a/b/c", "d/e/f"},
			},
			wantWorkspaces: []*RepoWorkspace{
				{RepoRevision: repoRevs[0], Steps: steps, Path: "a/b"},
				{RepoRevision: repoRevs[0], Steps: steps, Path: "a/b/c"},
				{RepoRevision: repoRevs[0], Steps: steps, Path: "d/e/f"},
				{RepoRevision: repoRevs[1], Steps: steps, Path: ""},
				{RepoRevision: repoRevs[2], Steps: steps, Path: "a/b"},
				{RepoRevision: repoRevs[2], Steps: steps, Path: "a/b/c"},
				{RepoRevision: repoRevs[2], Steps: steps, Path: "d/e/f"},
			},
		},

		"workspace configuration matches repo with OnlyFetchWorkspace": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{
						OnlyFetchWorkspace: true,
						In:                 "*automation-testing",
						RootAtLocationOf:   "package.json",
					},
				},
			},
			finderResults: finderResults{
				repoRevs[0].Key(): {"a/b", "a/b/c", "d/e/f"},
				repoRevs[2].Key(): {"a/b", "a/b/c", "d/e/f"},
			},
			wantWorkspaces: []*RepoWorkspace{
				{RepoRevision: repoRevs[0], Steps: steps, Path: "a/b", OnlyFetchWorkspace: true},
				{RepoRevision: repoRevs[0], Steps: steps, Path: "a/b/c", OnlyFetchWorkspace: true},
				{RepoRevision: repoRevs[0], Steps: steps, Path: "d/e/f", OnlyFetchWorkspace: true},
				{RepoRevision: repoRevs[1], Steps: steps, Path: ""},
				{RepoRevision: repoRevs[2], Steps: steps, Path: "a/b", OnlyFetchWorkspace: true},
				{RepoRevision: repoRevs[2], Steps: steps, Path: "a/b/c", OnlyFetchWorkspace: true},
				{RepoRevision: repoRevs[2], Steps: steps, Path: "d/e/f", OnlyFetchWorkspace: true},
			},
		},
		"workspace configuration without 'in' matches all": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{
						RootAtLocationOf: "package.json",
					},
				},
			},
			finderResults: finderResults{
				repoRevs[0].Key(): {"a/b"},
				repoRevs[2].Key(): {"a/b"},
			},
			wantWorkspaces: []*RepoWorkspace{
				{RepoRevision: repoRevs[0], Steps: steps, Path: "a/b"},
				{RepoRevision: repoRevs[2], Steps: steps, Path: "a/b"},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			finder := &mockDirectoryFinder{results: tt.finderResults}
			workspaces, err := findWorkspaces(context.Background(), tt.spec, finder, repoRevs)
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			// Sort by ID, easier than by name for tests.
			sort.Slice(workspaces, func(i, j int) bool {
				if workspaces[i].Repo.ID == workspaces[j].Repo.ID {
					return workspaces[i].Path < workspaces[j].Path
				}
				return workspaces[i].Repo.ID < workspaces[j].Repo.ID
			})

			if diff := cmp.Diff(tt.wantWorkspaces, workspaces); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

type mockDirectoryFinder struct {
	results map[repoRevKey][]string
}

func (m *mockDirectoryFinder) FindDirectoriesInRepos(ctx context.Context, fileName string, repos ...*RepoRevision) (map[repoRevKey][]string, error) {
	return m.results, nil
}

func TestStepsForRepoRevision(t *testing.T) {
	tests := map[string]struct {
		spec *batcheslib.BatchSpec

		wantSteps   []batcheslib.Step
		wantSkipped []int32
	}{
		"no if": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1"},
			},
		},

		"if has static true value": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: "true"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1", If: "true"},
			},
		},

		"one of many steps has if with static true value": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "true"},
					{Run: "echo 3"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1"},
				{Run: "echo 2", If: "true"},
				{Run: "echo 3"},
			},
		},

		"if has static non-true value": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: "this is not true"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1", If: "this is not true"},
			},
			wantSkipped: []int32{0},
		},

		"one of many steps has if with static non-true value": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "every type system needs generics"},
					{Run: "echo 3"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1"},
				{Run: "echo 2", If: "every type system needs generics"},
				{Run: "echo 3"},
			},
			wantSkipped: []int32{1},
		},

		"if expression that can be partially evaluated to true": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: `${{ matches repository.name "github.com/sourcegraph/src*" }}`},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1", If: `${{ matches repository.name "github.com/sourcegraph/src*" }}`},
			},
		},

		"if expression that can be partially evaluated to false": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: `${{ matches repository.name "horse" }}`},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1", If: `${{ matches repository.name "horse" }}`},
			},
			wantSkipped: []int32{0},
		},

		"one of many steps has if expression that can be evaluated to false": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: `${{ matches repository.name "horse" }}`},
					{Run: "echo 3"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1"},
				{Run: "echo 2", If: `${{ matches repository.name "horse" }}`},
				{Run: "echo 3"},
			},
			wantSkipped: []int32{1},
		},

		"if expression that can NOT be partially evaluated": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: `${{ eq outputs.value "foobar" }}`},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1", If: `${{ eq outputs.value "foobar" }}`},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			haveSteps, haveSkipped, err := stepsForRepo(tt.spec, "github.com/sourcegraph/src-cli", []string{})
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			opts := cmpopts.IgnoreUnexported(batcheslib.Step{})
			if diff := cmp.Diff(tt.wantSteps, haveSteps, opts); diff != "" {
				t.Errorf("mismatch in steps (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantSkipped, haveSkipped, opts); diff != "" {
				t.Errorf("mismatch in skipped (-want +got):\n%s", diff)
			}
		})
	}
}
