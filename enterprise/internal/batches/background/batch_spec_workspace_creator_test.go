package background

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
)

func TestBatchSpecWorkspaceCreatorProcess(t *testing.T) {
	db := dbtest.NewDB(t)

	repos, _ := ct.CreateTestRepos(t, context.Background(), db, 4)

	user := ct.CreateTestUser(t, db, true)

	s := store.New(db, &observation.TestContext, nil)

	batchSpec := &btypes.BatchSpec{UserID: user.ID, NamespaceUserID: user.ID, RawSpec: ct.TestRawBatchSpecYAML}
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
				Steps:              []batcheslib.Step{},
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
				Steps:              []batcheslib.Step{},
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
				Steps:              []batcheslib.Step{},
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
				Steps:       []batcheslib.Step{},
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
				Steps:   []batcheslib.Step{},
				Ignored: true,
			},
		},
	}

	creator := &batchSpecWorkspaceCreator{store: s}
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
			Steps:              []batcheslib.Step{},
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
			Steps:              []batcheslib.Step{},
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
			Steps:              []batcheslib.Step{},
			OnlyFetchWorkspace: true,
		},
		{
			RepoID:           repos[2].ID,
			BatchSpecID:      batchSpec.ID,
			Branch:           "refs/heads/base-branch",
			Commit:           "h0rs3s",
			ChangesetSpecIDs: []int64{},
			FileMatches:      []string{"main.go"},
			Steps:            []batcheslib.Step{},
			Unsupported:      true,
		},
		{
			RepoID:           repos[3].ID,
			BatchSpecID:      batchSpec.ID,
			Branch:           "refs/heads/main-base-branch",
			Commit:           "f00b4r",
			ChangesetSpecIDs: []int64{},
			FileMatches:      []string{"lol.txt"},
			Steps:            []batcheslib.Step{},
			Ignored:          true,
		},
	}

	opts := []cmp.Option{
		cmpopts.IgnoreFields(btypes.BatchSpecWorkspace{}, "ID", "CreatedAt", "UpdatedAt"),
	}
	if diff := cmp.Diff(want, have, opts...); diff != "" {
		t.Fatalf("wrong diff: %s", diff)
	}
}

func TestBatchSpecWorkspaceCreatorProcess_Caching(t *testing.T) {
	db := dbtest.NewDB(t)

	repos, _ := ct.CreateTestRepos(t, context.Background(), db, 1)

	user := ct.CreateTestUser(t, db, true)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	s := store.NewWithClock(db, &observation.TestContext, nil, clock)

	batchSpec, err := btypes.NewBatchSpecFromRaw(ct.TestRawBatchSpecYAML, false)
	if err != nil {
		t.Fatal(err)
	}
	batchSpec.UserID = user.ID
	batchSpec.NamespaceUserID = user.ID
	if err := s.CreateBatchSpec(context.Background(), batchSpec); err != nil {
		t.Fatal(err)
	}

	job := &btypes.BatchSpecResolutionJob{BatchSpecID: batchSpec.ID}

	workspace := &service.RepoWorkspace{
		RepoRevision: &service.RepoRevision{
			Repo:        repos[0],
			Branch:      "refs/heads/main",
			Commit:      "d34db33f",
			FileMatches: []string{},
		},
		Path:               "",
		Steps:              []batcheslib.Step{},
		OnlyFetchWorkspace: true,
	}

	key, err := cacheKeyForWorkspace(batchSpec, workspace)
	if err != nil {
		t.Fatal(err)
	}
	executionResult := &execution.Result{
		Diff:         testDiff,
		ChangedFiles: &git.Changes{Modified: []string{"README.md", "urls.txt"}},
		Outputs:      map[string]interface{}{},
	}

	entry, err := btypes.NewCacheEntryFromResult(key, executionResult)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CreateBatchSpecExecutionCacheEntry(context.Background(), entry); err != nil {
		t.Fatal(err)
	}

	resolver := &dummyWorkspaceResolver{workspaces: []*service.RepoWorkspace{workspace}}

	creator := &batchSpecWorkspaceCreator{store: s}
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
			ChangesetSpecIDs:   have[0].ChangesetSpecIDs,
			Branch:             "refs/heads/main",
			Commit:             "d34db33f",
			FileMatches:        []string{},
			Path:               "",
			Steps:              []batcheslib.Step{},
			OnlyFetchWorkspace: true,
			CachedResultFound:  true,
		},
	}

	opts := []cmp.Option{
		cmpopts.IgnoreFields(btypes.BatchSpecWorkspace{}, "ID", "CreatedAt", "UpdatedAt"),
	}
	if diff := cmp.Diff(want, have, opts...); diff != "" {
		t.Fatalf("wrong diff: %s", diff)
	}

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

	reloadedEntry, err := s.GetBatchSpecExecutionCacheEntry(context.Background(), store.GetBatchSpecExecutionCacheEntryOpts{Key: entry.Key})
	if err != nil {
		t.Fatal(err)
	}
	if !reloadedEntry.LastUsedAt.Equal(now) {
		t.Fatalf("cache entry LastUsedAt not updated. want=%s, have=%s", now, reloadedEntry.LastUsedAt)
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
