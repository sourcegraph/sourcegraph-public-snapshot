package store

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

func testStoreBatchSpecWorkspaces(t *testing.T, ctx context.Context, s *Store, clock ct.Clock) {
	repoStore := database.ReposWith(s)
	esStore := database.ExternalServicesWith(s)

	repo := ct.TestRepo(t, esStore, extsvc.KindGitHub)
	deletedRepo := ct.TestRepo(t, esStore, extsvc.KindGitHub).With(types.Opt.RepoDeletedAt(clock.Now()))

	if err := repoStore.Create(ctx, repo, deletedRepo); err != nil {
		t.Fatal(err)
	}
	if err := repoStore.Delete(ctx, deletedRepo.ID); err != nil {
		t.Fatal(err)
	}

	workspaces := make([]*btypes.BatchSpecWorkspace, 0, 3)
	for i := 0; i < cap(workspaces); i++ {
		job := &btypes.BatchSpecWorkspace{
			BatchSpecID:      int64(i + 567),
			ChangesetSpecIDs: []int64{int64(i + 456), int64(i + 678)},
			RepoID:           repo.ID,
			Branch:           "master",
			Commit:           "d34db33f",
			Path:             "sub/dir/ectory",
			FileMatches: []string{
				"a.go",
				"a/b/horse.go",
				"a/b/c.go",
			},
			Steps: []batcheslib.Step{
				{
					Run:       "complex command that changes code",
					Container: "alpine:3",
					Files: map[string]string{
						"/tmp/foobar.go": "package main",
					},
					Outputs: map[string]batcheslib.Output{
						"myOutput": {Value: `${{ step.stdout }}`},
					},
					If: `${{ eq repository.name "github.com/sourcegraph/sourcegraph" }}`,
				},
			},
			OnlyFetchWorkspace: true,
		}

		if i == cap(workspaces)-1 {
			job.RepoID = deletedRepo.ID
		}

		workspaces = append(workspaces, job)
	}

	t.Run("Create", func(t *testing.T) {
		for _, job := range workspaces {
			if err := s.CreateBatchSpecWorkspace(ctx, job); err != nil {
				t.Fatal(err)
			}

			have := job
			if have.ID == 0 {
				t.Fatal("ID should not be zero")
			}

			want := have
			want.CreatedAt = clock.Now()
			want.UpdatedAt = clock.Now()

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("GetByID", func(t *testing.T) {
			for i, job := range workspaces {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					have, err := s.GetBatchSpecWorkspace(ctx, GetBatchSpecWorkspaceOpts{ID: job.ID})

					if job.RepoID == deletedRepo.ID {
						if err != ErrNoResults {
							t.Fatalf("wrong error: %s", err)
						}
						return
					}

					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, job); diff != "" {
						t.Fatal(diff)
					}
				})
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetBatchSpecWorkspaceOpts{ID: 0xdeadbeef}

			_, have := s.GetBatchSpecWorkspace(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("All", func(t *testing.T) {
			have, err := s.ListBatchSpecWorkspaces(ctx, ListBatchSpecWorkspacesOpts{})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(have, workspaces[:len(workspaces)-1]); diff != "" {
				t.Fatalf("invalid jobs returned: %s", diff)
			}
		})
	})
}
