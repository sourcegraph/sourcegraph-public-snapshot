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

func TestService_ResolveWorkspacesForBatchSpec(t *testing.T) {
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
	mockDefaultBranches(t, defaultBranches)

	defaultOpts := ResolveWorkspacesForBatchSpecOpts{
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
			Steps: steps,
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
				Path:         "test",
				RepositoryID: int32(rs[2].ID),
			},
			&streamhttp.EventSymbolMatch{
				Type:         streamhttp.SymbolMatchType,
				Path:         "test",
				RepositoryID: int32(rs[3].ID),
			},
		}

		want := []*RepoWorkspace{buildRepoWorkspace(rs[0], "", "", []string{"test", "duplicate-test"}), buildRepoWorkspace(rs[3], "", "", []string{"test"})}
		wantIgnored := []api.RepoID{rs[1].ID, rs[2].ID}
		wantUnsupported := []api.RepoID{}
		resolveWorkspacesAndCompare(t, s, defaultOpts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)
	})

	t.Run("repositories", func(t *testing.T) {
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				{Repository: string(rs[0].Name)},
				{Repository: string(rs[1].Name), Branch: "non-default-branch"},
				{Repository: string(rs[2].Name), Branch: "other-non-default-branch"},
				{Repository: string(rs[3].Name)},
			},
			Steps: steps,
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

		want := []*RepoWorkspace{
			buildRepoWorkspace(rs[0], "", "", []string{}),
			buildRepoWorkspace(rs[1], "non-default-branch", "d34db33f", []string{}),
			buildRepoWorkspace(rs[2], "other-non-default-branch", "c0ff33", []string{}),
		}

		wantIgnored := []api.RepoID{rs[3].ID}
		wantUnsupported := []api.RepoID{}
		resolveWorkspacesAndCompare(t, s, defaultOpts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)
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

		want := []*RepoWorkspace{
			buildRepoWorkspace(rs[0], "", "", []string{"test"}),
			buildRepoWorkspace(rs[1], "", "", []string{}),
			buildRepoWorkspace(rs[2], "", "", []string{}),
			buildRepoWorkspace(rs[3], "", "", []string{}),
		}

		wantIgnored := []api.RepoID{}
		wantUnsupported := []api.RepoID{}
		resolveWorkspacesAndCompare(t, s, defaultOpts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)
	})

	t.Run("allowUnsupported option", func(t *testing.T) {
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				{RepositoriesMatchingQuery: "repohasfile:horse.txt"},
			},
			Steps: steps,
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

		want := []*RepoWorkspace{}
		wantIgnored := []api.RepoID{}
		wantUnsupported := []api.RepoID{unsupported[0].ID}
		resolveWorkspacesAndCompare(t, s, defaultOpts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)

		// with allowUnsupported: true
		// Now we expect the repo to be returned.
		opts := defaultOpts
		opts.AllowUnsupported = true

		want = []*RepoWorkspace{buildRepoWorkspace(unsupported[0], "", "", []string{"test"})}
		wantUnsupported = []api.RepoID{unsupported[0].ID}
		resolveWorkspacesAndCompare(t, s, opts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)
	})

	t.Run("allowIgnored option", func(t *testing.T) {
		batchSpec := &batcheslib.BatchSpec{
			On: []batcheslib.OnQueryOrRepository{
				{Repository: string(rs[0].Name)},
			},
			Steps: steps,
		}

		mockResolveRevision(t, map[string]api.CommitID{
			defaultBranches[rs[0].Name].branch: defaultBranches[rs[0].Name].commit,
		})

		mockBatchIgnores(t, map[api.CommitID]bool{
			defaultBranches[rs[0].Name].commit: true,
		})

		searchMatches := []streamhttp.EventMatch{}

		want := []*RepoWorkspace{}
		wantIgnored := []api.RepoID{rs[0].ID}
		wantUnsupported := []api.RepoID{}
		resolveWorkspacesAndCompare(t, s, defaultOpts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)

		// with allowIgnored: true
		// Now we expect the repo to be returned.
		opts := defaultOpts
		opts.AllowIgnored = true

		want = []*RepoWorkspace{buildRepoWorkspace(rs[0], "", "", []string{})}
		wantIgnored = []api.RepoID{rs[0].ID}
		resolveWorkspacesAndCompare(t, s, opts, searchMatches, batchSpec, want, wantIgnored, wantUnsupported)
	})
}

func resolveWorkspacesAndCompare(t *testing.T, s *store.Store, opts ResolveWorkspacesForBatchSpecOpts, matches []streamhttp.EventMatch, spec *batcheslib.BatchSpec, want []*RepoWorkspace, wantIgnored, wantUnsupported []api.RepoID) {
	t.Helper()

	wr := &workspaceResolver{
		store:               s,
		frontendInternalURL: newStreamSearchTestServer(t, matches),
	}
	have, unsupported, ignored, err := wr.ResolveWorkspacesForBatchSpec(context.Background(), spec, opts)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	{
		repoIDs := make([]api.RepoID, 0, len(ignored))
		for r := range ignored {
			repoIDs = append(repoIDs, r.ID)
		}
		sort.Slice(repoIDs, func(i, j int) bool { return repoIDs[i] < repoIDs[j] })
		sort.Slice(wantIgnored, func(i, j int) bool { return wantIgnored[i] < wantIgnored[j] })
		if diff := cmp.Diff(repoIDs, wantIgnored); diff != "" {
			t.Fatalf("Invalid ignored repos returned: %s", diff)
		}
	}
	{
		repoIDs := make([]api.RepoID, 0, len(unsupported))
		for r := range unsupported {
			repoIDs = append(repoIDs, r.ID)
		}
		sort.Slice(repoIDs, func(i, j int) bool { return repoIDs[i] < repoIDs[j] })
		sort.Slice(wantUnsupported, func(i, j int) bool { return wantUnsupported[i] < wantUnsupported[j] })
		if diff := cmp.Diff(repoIDs, wantUnsupported); diff != "" {
			t.Fatalf("Invalid unsupported repos returned: %s", diff)
		}
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("returned workspaces wrong. (-want +got):\n%s", diff)
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

func mockResolveRevision(t *testing.T, branches map[string]api.CommitID) {
	git.Mocks.ResolveRevision = func(spec string, opt git.ResolveRevisionOptions) (api.CommitID, error) {
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

		wantSteps []batcheslib.Step
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
			wantSteps: []batcheslib.Step{},
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
				{Run: "echo 3"},
			},
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
			wantSteps: []batcheslib.Step{},
		},

		"one of many steps has if expression that can be evaluated to true": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: `${{ matches repository.name "horse" }}`},
				},
			},
			wantSteps: []batcheslib.Step{},
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
			haveSteps, err := stepsForRepo(tt.spec, "github.com/sourcegraph/src-cli", []string{})
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			opts := cmpopts.IgnoreUnexported(batcheslib.Step{})
			if diff := cmp.Diff(tt.wantSteps, haveSteps, opts); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
