package background

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

func TestBatchSpecWorkspaceCreatorProcess(t *testing.T) {
	db := dbtest.NewDB(t, "")

	repos, _ := ct.CreateTestRepos(t, context.Background(), db, 2)

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
		},
	}

	creator := &batchSpecWorkspaceCreator{store: s}
	if err := creator.process(context.Background(), s, resolver, job); err != nil {
		t.Fatalf("proces failed: %s", err)
	}

	have, err := s.ListBatchSpecWorkspaces(context.Background(), store.ListBatchSpecWorkspacesOpts{BatchSpecID: batchSpec.ID})
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
	}

	opts := []cmp.Option{
		cmpopts.IgnoreFields(btypes.BatchSpecWorkspace{}, "ID", "CreatedAt", "UpdatedAt"),
	}
	if diff := cmp.Diff(want, have, opts...); diff != "" {
		t.Fatalf("wrong diff: %s", diff)
	}
}

type dummyWorkspaceResolver struct {
	workspaces  []*service.RepoWorkspace
	unsupported map[*types.Repo]struct{}
	ignored     map[*types.Repo]struct{}
	err         error
}

func (d *dummyWorkspaceResolver) ResolveWorkspacesForBatchSpec(context.Context, *batcheslib.BatchSpec, service.ResolveWorkspacesForBatchSpecOpts) ([]*service.RepoWorkspace, map[*types.Repo]struct{}, map[*types.Repo]struct{}, error) {
	return d.workspaces, d.unsupported, d.ignored, d.err
}
