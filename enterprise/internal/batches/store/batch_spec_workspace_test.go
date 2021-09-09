package store

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

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

	jobs := make([]*btypes.BatchSpecWorkspace, 0, 3)
	for i := 0; i < cap(jobs); i++ {
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

		switch i {
		case 0:
			job.State = btypes.BatchSpecWorkspaceStatePending
		case 1:
			job.State = btypes.BatchSpecWorkspaceStateProcessing
		case 2:
			job.State = btypes.BatchSpecWorkspaceStateFailed
		}
		if i == cap(jobs)-1 {
			job.RepoID = deletedRepo.ID
		}

		jobs = append(jobs, job)
	}

	t.Run("Create", func(t *testing.T) {
		for _, job := range jobs {
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
			for i, job := range jobs {
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
		for i, job := range jobs {
			job.WorkerHostname = fmt.Sprintf("worker-hostname-%d", i)
			if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_workspaces SET worker_hostname = %s, state = %s WHERE id = %s", job.WorkerHostname, job.State, job.ID)); err != nil {
				t.Fatal(err)
			}
		}

		t.Run("All", func(t *testing.T) {
			have, err := s.ListBatchSpecWorkspaces(ctx, ListBatchSpecWorkspacesOpts{})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(have, jobs[:len(jobs)-1]); diff != "" {
				t.Fatalf("invalid jobs returned: %s", diff)
			}
		})

		t.Run("WorkerHostname", func(t *testing.T) {
			for _, job := range jobs {
				have, err := s.ListBatchSpecWorkspaces(ctx, ListBatchSpecWorkspacesOpts{
					WorkerHostname: job.WorkerHostname,
				})

				if job.RepoID == deletedRepo.ID {
					if len(have) != 0 {
						t.Fatalf("expected no jobs to be returned, but got %d", len(have))
					}
					return
				}

				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(have, []*btypes.BatchSpecWorkspace{job}); diff != "" {
					t.Fatalf("invalid batch spec workspace jobs returned: %s", diff)
				}
			}
		})

		t.Run("State", func(t *testing.T) {
			for _, job := range jobs {
				have, err := s.ListBatchSpecWorkspaces(ctx, ListBatchSpecWorkspacesOpts{
					State: job.State,
				})

				if job.RepoID == deletedRepo.ID {
					if len(have) != 0 {
						t.Fatalf("expected no jobs to be returned, but got %d", len(have))
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(have, []*btypes.BatchSpecWorkspace{job}); diff != "" {
					t.Fatalf("invalid batch spec workspace jobs returned: %s", diff)
				}
			}
		})
	})
}
