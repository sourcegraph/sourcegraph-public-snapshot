package workers

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
)

func TestBatchSpecWorkspaceCreatorProcess(t *testing.T) {
	db := database.NewDB(dbtest.NewDB(t))

	repos, _ := ct.CreateTestRepos(t, context.Background(), db, 4)

	user := ct.CreateTestUser(t, db, true)

	s := store.New(db, &observation.TestContext, nil)

	batchSpec, err := btypes.NewBatchSpecFromRaw(ct.TestRawBatchSpecYAML)
	if err != nil {
		t.Fatal(err)
	}
	batchSpec.UserID = user.ID
	batchSpec.NamespaceUserID = user.ID
	if err := s.CreateBatchSpec(context.Background(), batchSpec); err != nil {
		t.Fatal(err)
	}

	job := &btypes.BatchSpecResolutionJob{BatchSpecID: batchSpec.ID}

	resolver := &dummyWorkspaceResolver{
		workspaces: []*service.RepoWorkspace{
			{
				RepoRevision: &service.RepoRevision{
					Repo:        repos[0],
					Branch:      "refs/heads/main",
					Commit:      "d34db33f",
					FileMatches: []string{},
				},
				Path:               "",
				OnlyFetchWorkspace: true,
			},
			{
				RepoRevision: &service.RepoRevision{
					Repo:        repos[0],
					Branch:      "refs/heads/main",
					Commit:      "d34db33f",
					FileMatches: []string{"a/b/c.go"},
				},
				Path:               "a/b",
				OnlyFetchWorkspace: false,
			},
			{
				RepoRevision: &service.RepoRevision{
					Repo:        repos[1],
					Branch:      "refs/heads/base-branch",
					Commit:      "c0ff33",
					FileMatches: []string{"d/e/f.go"},
				},
				Path:               "d/e",
				OnlyFetchWorkspace: true,
			},
			{
				// Unsupported
				RepoRevision: &service.RepoRevision{
					Repo:        repos[2],
					Branch:      "refs/heads/base-branch",
					Commit:      "h0rs3s",
					FileMatches: []string{"main.go"},
				},
				Path:        "",
				Unsupported: true,
			},
			{
				// Ignored
				RepoRevision: &service.RepoRevision{
					Repo:        repos[3],
					Branch:      "refs/heads/main-base-branch",
					Commit:      "f00b4r",
					FileMatches: []string{"lol.txt"},
				},
				Path:    "",
				Ignored: true,
			},
		},
	}

	creator := &batchSpecWorkspaceCreator{store: s, logger: logtest.Scoped(t)}
	if err := creator.process(context.Background(), s, resolver.DummyBuilder, job); err != nil {
		t.Fatalf("proces failed: %s", err)
	}

	have, _, err := s.ListBatchSpecWorkspaces(context.Background(), store.ListBatchSpecWorkspacesOpts{BatchSpecID: batchSpec.ID})
	if err != nil {
		t.Fatalf("listing workspaces failed: %s", err)
	}

	want := []*btypes.BatchSpecWorkspace{
		{
			RepoID:             repos[0].ID,
			BatchSpecID:        batchSpec.ID,
			ChangesetSpecIDs:   []int64{},
			Branch:             "refs/heads/main",
			Commit:             "d34db33f",
			FileMatches:        []string{},
			Path:               "",
			OnlyFetchWorkspace: true,
		},
		{
			RepoID:             repos[0].ID,
			BatchSpecID:        batchSpec.ID,
			ChangesetSpecIDs:   []int64{},
			Branch:             "refs/heads/main",
			Commit:             "d34db33f",
			FileMatches:        []string{"a/b/c.go"},
			Path:               "a/b",
			OnlyFetchWorkspace: false,
		},
		{
			RepoID:             repos[1].ID,
			BatchSpecID:        batchSpec.ID,
			ChangesetSpecIDs:   []int64{},
			Branch:             "refs/heads/base-branch",
			Commit:             "c0ff33",
			FileMatches:        []string{"d/e/f.go"},
			Path:               "d/e",
			OnlyFetchWorkspace: true,
		},
		{
			RepoID:           repos[2].ID,
			BatchSpecID:      batchSpec.ID,
			Branch:           "refs/heads/base-branch",
			Commit:           "h0rs3s",
			ChangesetSpecIDs: []int64{},
			FileMatches:      []string{"main.go"},
			Unsupported:      true,
		},
		{
			RepoID:           repos[3].ID,
			BatchSpecID:      batchSpec.ID,
			Branch:           "refs/heads/main-base-branch",
			Commit:           "f00b4r",
			ChangesetSpecIDs: []int64{},
			FileMatches:      []string{"lol.txt"},
			Ignored:          true,
		},
	}

	assertWorkspacesEqual(t, have, want)
}

func TestBatchSpecWorkspaceCreatorProcess_Caching(t *testing.T) {
	db := database.NewDB(dbtest.NewDB(t))

	repos, _ := ct.CreateTestRepos(t, context.Background(), db, 1)

	user := ct.CreateTestUser(t, db, true)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	s := store.NewWithClock(db, &observation.TestContext, nil, clock)

	creator := &batchSpecWorkspaceCreator{store: s, logger: logtest.Scoped(t)}

	buildWorkspace := func(commit string) *service.RepoWorkspace {
		return &service.RepoWorkspace{
			RepoRevision: &service.RepoRevision{
				Repo:   repos[0],
				Branch: "refs/heads/main",
				// We use a different commit so we get different cache keys and
				// don't overwrite the cache keys in the tests.
				Commit:      api.CommitID(commit),
				FileMatches: []string{},
			},
			Path:               "",
			OnlyFetchWorkspace: true,
		}
	}

	executionResult := &execution.AfterStepResult{
		Diff:      testDiff,
		StepIndex: 0,
		// todo: required? step 0 should not have this.
		PreviousStepResult: execution.StepResult{
			Files:  &git.Changes{Modified: []string{"README.md", "urls.txt"}},
			Stdout: bytes.NewBufferString("asdf2"),
			Stderr: bytes.NewBufferString("asdf"),
		},
		Outputs: map[string]any{},
	}

	createBatchSpec := func(t *testing.T, noCache bool, spec string) *btypes.BatchSpec {
		batchSpec, err := btypes.NewBatchSpecFromRaw(spec)
		if err != nil {
			t.Fatal(err)
		}
		batchSpec.UserID = user.ID
		batchSpec.NamespaceUserID = user.ID
		batchSpec.NoCache = noCache
		if err := s.CreateBatchSpec(context.Background(), batchSpec); err != nil {
			t.Fatal(err)
		}
		return batchSpec
	}

	createCacheEntry := func(t *testing.T, batchSpec *btypes.BatchSpec, workspace *service.RepoWorkspace, result *execution.AfterStepResult) *btypes.BatchSpecExecutionCacheEntry {
		t.Helper()

		execKey := cache.KeyForWorkspace(
			&template.BatchChangeAttributes{
				Name:        batchSpec.Spec.Name,
				Description: batchSpec.Spec.Description,
			},
			batcheslib.Repository{
				ID:          string(graphqlbackend.MarshalRepositoryID(workspace.Repo.ID)),
				Name:        string(workspace.Repo.Name),
				BaseRef:     workspace.Branch,
				BaseRev:     string(workspace.Commit),
				FileMatches: workspace.FileMatches,
			},
			workspace.Path,
			workspace.OnlyFetchWorkspace,
			batchSpec.Spec.Steps,
		)
		key := cache.StepsCacheKey{
			ExecutionKey: &execKey,
			StepIndex:    0,
		}
		rawKey, err := key.Key()
		if err != nil {
			t.Fatal(err)
		}
		entry, err := btypes.NewCacheEntryFromResult(rawKey, result)
		if err != nil {
			t.Fatal(err)
		}
		entry.UserID = batchSpec.UserID
		if err := s.CreateBatchSpecExecutionCacheEntry(context.Background(), entry); err != nil {
			t.Fatal(err)
		}
		return entry
	}

	t.Run("caching enabled", func(t *testing.T) {
		workspace := buildWorkspace("caching-enabled")

		batchSpec := createBatchSpec(t, false, ct.TestRawBatchSpecYAML)
		entry := createCacheEntry(t, batchSpec, workspace, executionResult)

		resolver := &dummyWorkspaceResolver{workspaces: []*service.RepoWorkspace{workspace}}
		job := &btypes.BatchSpecResolutionJob{BatchSpecID: batchSpec.ID}
		if err := creator.process(context.Background(), s, resolver.DummyBuilder, job); err != nil {
			t.Fatalf("proces failed: %s", err)
		}

		have, _, err := s.ListBatchSpecWorkspaces(context.Background(), store.ListBatchSpecWorkspacesOpts{BatchSpecID: batchSpec.ID})
		if err != nil {
			t.Fatalf("listing workspaces failed: %s", err)
		}

		assertWorkspacesEqual(t, have, []*btypes.BatchSpecWorkspace{
			{
				RepoID:             repos[0].ID,
				BatchSpecID:        batchSpec.ID,
				ChangesetSpecIDs:   have[0].ChangesetSpecIDs,
				Branch:             "refs/heads/main",
				Commit:             "caching-enabled",
				FileMatches:        []string{},
				Path:               "",
				OnlyFetchWorkspace: true,
				CachedResultFound:  true,
				StepCacheResults: map[int]btypes.StepCacheResult{
					1: {
						Key:   entry.Key,
						Value: executionResult,
					},
				},
			},
		})

		changesetSpecIDs := have[0].ChangesetSpecIDs
		if len(changesetSpecIDs) == 0 {
			t.Fatal("BatchSpecWorkspace has no changeset specs")
		}

		changesetSpec, err := s.GetChangesetSpec(context.Background(), store.GetChangesetSpecOpts{ID: have[0].ChangesetSpecIDs[0]})
		if err != nil {
			t.Fatal(err)
		}

		haveDiff, err := changesetSpec.Spec.Diff()
		if err != nil {
			t.Fatal(err)
		}
		if haveDiff != testDiff {
			t.Fatalf("changeset spec built from cache has wrong diff: %s", haveDiff)
		}

		reloadedEntries, err := s.ListBatchSpecExecutionCacheEntries(context.Background(), store.ListBatchSpecExecutionCacheEntriesOpts{
			UserID: batchSpec.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(reloadedEntries) != 1 {
			t.Fatal("cache entry not found")
		}
		reloadedEntry := reloadedEntries[0]
		if !reloadedEntry.LastUsedAt.Equal(now) {
			t.Fatalf("cache entry LastUsedAt not updated. want=%s, have=%s", now, reloadedEntry.LastUsedAt)
		}
	})

	t.Run("only step is statically skipped", func(t *testing.T) {
		workspace := buildWorkspace("no-step-after-eval")

		spec := `
name: my-unique-name
description: My description
on:
- repository: github.com/sourcegraph/src-cli
steps:
- run: echo 'foobar'
  container: alpine
  if: ${{ eq repository.name "not the repo" }}
changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
`
		batchSpec := createBatchSpec(t, false, spec)

		resolver := &dummyWorkspaceResolver{workspaces: []*service.RepoWorkspace{workspace}}
		job := &btypes.BatchSpecResolutionJob{BatchSpecID: batchSpec.ID}
		if err := creator.process(context.Background(), s, resolver.DummyBuilder, job); err != nil {
			t.Fatalf("proces failed: %s", err)
		}

		have, _, err := s.ListBatchSpecWorkspaces(context.Background(), store.ListBatchSpecWorkspacesOpts{BatchSpecID: batchSpec.ID})
		if err != nil {
			t.Fatalf("listing workspaces failed: %s", err)
		}

		assertWorkspacesEqual(t, have, []*btypes.BatchSpecWorkspace{
			{
				RepoID:             repos[0].ID,
				BatchSpecID:        batchSpec.ID,
				ChangesetSpecIDs:   []int64{},
				Branch:             "refs/heads/main",
				Commit:             "no-step-after-eval",
				FileMatches:        []string{},
				Path:               "",
				OnlyFetchWorkspace: true,
				CachedResultFound:  true,
			},
		})

		changesetSpecIDs := have[0].ChangesetSpecIDs
		if len(changesetSpecIDs) != 0 {
			t.Fatal("BatchSpecWorkspace has changeset specs, even though nothing ran")
		}
	})

	t.Run("all steps are statically skipped", func(t *testing.T) {
		workspace := buildWorkspace("no-steps-after-eval")

		spec := `
name: my-unique-name
description: My description
on:
- repository: github.com/sourcegraph/src-cli
steps:
- run: echo 'foobar'
  container: alpine
  if: ${{ eq repository.name "not the repo" }}
- run: echo 'foobar'
  container: alpine
  if: ${{ eq repository.name "not the repo" }}
changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
`
		batchSpec := createBatchSpec(t, false, spec)

		resolver := &dummyWorkspaceResolver{workspaces: []*service.RepoWorkspace{workspace}}
		job := &btypes.BatchSpecResolutionJob{BatchSpecID: batchSpec.ID}
		if err := creator.process(context.Background(), s, resolver.DummyBuilder, job); err != nil {
			t.Fatalf("proces failed: %s", err)
		}

		have, _, err := s.ListBatchSpecWorkspaces(context.Background(), store.ListBatchSpecWorkspacesOpts{BatchSpecID: batchSpec.ID})
		if err != nil {
			t.Fatalf("listing workspaces failed: %s", err)
		}

		assertWorkspacesEqual(t, have, []*btypes.BatchSpecWorkspace{
			{
				RepoID:             repos[0].ID,
				BatchSpecID:        batchSpec.ID,
				ChangesetSpecIDs:   []int64{},
				Branch:             "refs/heads/main",
				Commit:             "no-steps-after-eval",
				FileMatches:        []string{},
				Path:               "",
				OnlyFetchWorkspace: true,
				CachedResultFound:  true,
			},
		})

		changesetSpecIDs := have[0].ChangesetSpecIDs
		if len(changesetSpecIDs) != 0 {
			t.Fatal("BatchSpecWorkspace has changeset specs, even though nothing ran")
		}
	})

	t.Run("caching enabled but no diff in cache entry", func(t *testing.T) {
		workspace := buildWorkspace("caching-enabled-no-diff")

		batchSpec := createBatchSpec(t, false, ct.TestRawBatchSpecYAML)

		resultWithoutDiff := *executionResult
		resultWithoutDiff.Diff = ""

		entry := createCacheEntry(t, batchSpec, workspace, &resultWithoutDiff)

		resolver := &dummyWorkspaceResolver{workspaces: []*service.RepoWorkspace{workspace}}
		job := &btypes.BatchSpecResolutionJob{BatchSpecID: batchSpec.ID}
		if err := creator.process(context.Background(), s, resolver.DummyBuilder, job); err != nil {
			t.Fatalf("proces failed: %s", err)
		}

		have, _, err := s.ListBatchSpecWorkspaces(context.Background(), store.ListBatchSpecWorkspacesOpts{BatchSpecID: batchSpec.ID})
		if err != nil {
			t.Fatalf("listing workspaces failed: %s", err)
		}

		changesetSpecIDs := have[0].ChangesetSpecIDs
		if len(changesetSpecIDs) != 0 {
			t.Fatal("BatchSpecWorkspace has changeset specs, even though diff was empty")
		}

		reloadedEntries, err := s.ListBatchSpecExecutionCacheEntries(context.Background(), store.ListBatchSpecExecutionCacheEntriesOpts{
			UserID: batchSpec.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(reloadedEntries) != 1 {
			t.Fatal("cache entry not found")
		}
		reloadedEntry := reloadedEntries[0]
		if !reloadedEntry.LastUsedAt.Equal(now) {
			t.Fatalf("cache entry LastUsedAt not updated. want=%s, have=%s", now, reloadedEntry.LastUsedAt)
		}
	})

	t.Run("caching disabled", func(t *testing.T) {
		workspace := buildWorkspace("caching-disabled")

		batchSpec := createBatchSpec(t, true, ct.TestRawBatchSpecYAML)
		entry := createCacheEntry(t, batchSpec, workspace, executionResult)

		resolver := &dummyWorkspaceResolver{workspaces: []*service.RepoWorkspace{workspace}}
		job := &btypes.BatchSpecResolutionJob{BatchSpecID: batchSpec.ID}
		if err := creator.process(context.Background(), s, resolver.DummyBuilder, job); err != nil {
			t.Fatalf("proces failed: %s", err)
		}

		have, _, err := s.ListBatchSpecWorkspaces(context.Background(), store.ListBatchSpecWorkspacesOpts{BatchSpecID: batchSpec.ID})
		if err != nil {
			t.Fatalf("listing workspaces failed: %s", err)
		}

		assertWorkspacesEqual(t, have, []*btypes.BatchSpecWorkspace{
			{
				RepoID:             repos[0].ID,
				BatchSpecID:        batchSpec.ID,
				ChangesetSpecIDs:   []int64{},
				Branch:             "refs/heads/main",
				Commit:             "caching-disabled",
				FileMatches:        []string{},
				Path:               "",
				OnlyFetchWorkspace: true,
				CachedResultFound:  false,
			},
		})

		reloadedEntries, err := s.ListBatchSpecExecutionCacheEntries(context.Background(), store.ListBatchSpecExecutionCacheEntriesOpts{
			UserID: batchSpec.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(reloadedEntries) != 1 {
			t.Fatal("cache entry not found")
		}
		reloadedEntry := reloadedEntries[0]
		if !reloadedEntry.LastUsedAt.IsZero() {
			t.Fatalf("cache entry LastUsedAt updated, but should not be used: %s", reloadedEntry.LastUsedAt)
		}
	})

	t.Run("caching enabled but workspace is ignored", func(t *testing.T) {
		workspace := buildWorkspace("caching-enabled-ignored")
		workspace.Ignored = true

		batchSpec := createBatchSpec(t, false, ct.TestRawBatchSpecYAML)

		entry := createCacheEntry(t, batchSpec, workspace, executionResult)

		resolver := &dummyWorkspaceResolver{workspaces: []*service.RepoWorkspace{workspace}}
		job := &btypes.BatchSpecResolutionJob{BatchSpecID: batchSpec.ID}
		if err := creator.process(context.Background(), s, resolver.DummyBuilder, job); err != nil {
			t.Fatalf("proces failed: %s", err)
		}

		reloadedEntries, err := s.ListBatchSpecExecutionCacheEntries(context.Background(), store.ListBatchSpecExecutionCacheEntriesOpts{
			UserID: batchSpec.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(reloadedEntries) != 1 {
			t.Fatal("cache entry not found")
		}
		reloadedEntry := reloadedEntries[0]
		if !reloadedEntry.LastUsedAt.IsZero() {
			t.Fatalf("cache entry LastUsedAt updated, but should not be used: %s", reloadedEntry.LastUsedAt)
		}
	})

	t.Run("caching enabled but workspace is unsupported", func(t *testing.T) {
		workspace := buildWorkspace("caching-enabled-ignored")
		workspace.Unsupported = true

		batchSpec := createBatchSpec(t, false, ct.TestRawBatchSpecYAML)

		entry := createCacheEntry(t, batchSpec, workspace, executionResult)

		resolver := &dummyWorkspaceResolver{workspaces: []*service.RepoWorkspace{workspace}}
		job := &btypes.BatchSpecResolutionJob{BatchSpecID: batchSpec.ID}
		if err := creator.process(context.Background(), s, resolver.DummyBuilder, job); err != nil {
			t.Fatalf("proces failed: %s", err)
		}

		reloadedEntries, err := s.ListBatchSpecExecutionCacheEntries(context.Background(), store.ListBatchSpecExecutionCacheEntriesOpts{
			UserID: batchSpec.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(reloadedEntries) != 1 {
			t.Fatal("cache entry not found")
		}
		reloadedEntry := reloadedEntries[0]
		if !reloadedEntry.LastUsedAt.IsZero() {
			t.Fatalf("cache entry LastUsedAt updated, but should not be used: %s", reloadedEntry.LastUsedAt)
		}
	})
}

func TestBatchSpecWorkspaceCreatorProcess_Importing(t *testing.T) {
	db := database.NewDB(dbtest.NewDB(t))

	repos, _ := ct.CreateTestRepos(t, context.Background(), db, 1)

	user := ct.CreateTestUser(t, db, true)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	s := store.NewWithClock(db, &observation.TestContext, nil, clock)

	var testSpecYAML = `
name: my-unique-name
importChangesets:
  - repository: ` + string(repos[0].Name) + `
    externalIDs:
      - 123
`

	batchSpec := &btypes.BatchSpec{UserID: user.ID, NamespaceUserID: user.ID, RawSpec: testSpecYAML}
	if err := s.CreateBatchSpec(context.Background(), batchSpec); err != nil {
		t.Fatal(err)
	}

	job := &btypes.BatchSpecResolutionJob{BatchSpecID: batchSpec.ID}

	resolver := &dummyWorkspaceResolver{}

	creator := &batchSpecWorkspaceCreator{store: s, logger: logtest.Scoped(t)}
	if err := creator.process(context.Background(), s, resolver.DummyBuilder, job); err != nil {
		t.Fatalf("proces failed: %s", err)
	}

	have, _, err := s.ListChangesetSpecs(context.Background(), store.ListChangesetSpecsOpts{BatchSpecID: batchSpec.ID})
	if err != nil {
		t.Fatalf("listing specs failed: %s", err)
	}

	want := btypes.ChangesetSpecs{
		{
			ID:          have[0].ID,
			RandID:      have[0].RandID,
			UserID:      user.ID,
			RepoID:      repos[0].ID,
			BatchSpecID: batchSpec.ID,
			Spec: &batcheslib.ChangesetSpec{
				BaseRepository: string(graphqlbackend.MarshalRepositoryID(repos[0].ID)),
				ExternalID:     "123",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}

func TestBatchSpecWorkspaceCreatorProcess_NoDiff(t *testing.T) {
	db := database.NewDB(dbtest.NewDB(t))

	repos, _ := ct.CreateTestRepos(t, context.Background(), db, 1)

	user := ct.CreateTestUser(t, db, true)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	s := store.NewWithClock(db, &observation.TestContext, nil, clock)

	var testSpecYAML = `
name: my-unique-name
importChangesets:
  - repository: ` + string(repos[0].Name) + `
    externalIDs:
      - 123
`

	batchSpec := &btypes.BatchSpec{UserID: user.ID, NamespaceUserID: user.ID, RawSpec: testSpecYAML}
	if err := s.CreateBatchSpec(context.Background(), batchSpec); err != nil {
		t.Fatal(err)
	}

	job := &btypes.BatchSpecResolutionJob{BatchSpecID: batchSpec.ID}

	resolver := &dummyWorkspaceResolver{}

	creator := &batchSpecWorkspaceCreator{store: s, logger: logtest.Scoped(t)}
	if err := creator.process(context.Background(), s, resolver.DummyBuilder, job); err != nil {
		t.Fatalf("proces failed: %s", err)
	}

	have, _, err := s.ListChangesetSpecs(context.Background(), store.ListChangesetSpecsOpts{BatchSpecID: batchSpec.ID})
	if err != nil {
		t.Fatalf("listing specs failed: %s", err)
	}

	want := btypes.ChangesetSpecs{
		{
			ID:          have[0].ID,
			RandID:      have[0].RandID,
			UserID:      user.ID,
			RepoID:      repos[0].ID,
			BatchSpecID: batchSpec.ID,
			Spec: &batcheslib.ChangesetSpec{
				BaseRepository: string(graphqlbackend.MarshalRepositoryID(repos[0].ID)),
				ExternalID:     "123",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}

type dummyWorkspaceResolver struct {
	workspaces []*service.RepoWorkspace
	err        error
}

// DummyBuilder is a simple implementation of the service.WorkspaceResolverBuilder
func (d *dummyWorkspaceResolver) DummyBuilder(s *store.Store) service.WorkspaceResolver {
	return d
}

func (d *dummyWorkspaceResolver) ResolveWorkspacesForBatchSpec(context.Context, *batcheslib.BatchSpec) ([]*service.RepoWorkspace, error) {
	return d.workspaces, d.err
}

const testDiff = `diff README.md README.md
index 671e50a..851b23a 100644
--- README.md
+++ README.md
@@ -1,2 +1,2 @@
 # README
-This file is hosted at example.com and is a test file.
+This file is hosted at sourcegraph.com and is a test file.
diff --git urls.txt urls.txt
index 6f8b5d9..17400bc 100644
--- urls.txt
+++ urls.txt
@@ -1,3 +1,3 @@
 another-url.com
-example.com
+sourcegraph.com
 never-touch-the-mouse.com
`

func assertWorkspacesEqual(t *testing.T, have, want []*btypes.BatchSpecWorkspace) {
	t.Helper()

	opts := []cmp.Option{
		cmpopts.IgnoreFields(btypes.BatchSpecWorkspace{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreUnexported(bytes.Buffer{}),
	}
	if diff := cmp.Diff(want, have, opts...); diff != "" {
		t.Fatalf("wrong diff: %s", diff)
	}
}
