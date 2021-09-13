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
)

func testStoreBatchSpecWorkspaceExecutionJobs(t *testing.T, ctx context.Context, s *Store, clock ct.Clock) {
	jobs := make([]*btypes.BatchSpecWorkspaceExecutionJob, 0, 2)
	for i := 0; i < cap(jobs); i++ {
		job := &btypes.BatchSpecWorkspaceExecutionJob{
			BatchSpecWorkspaceID: int64(i + 456),
		}

		jobs = append(jobs, job)
	}

	t.Run("Create", func(t *testing.T) {
		for _, job := range jobs {
			if err := s.CreateBatchSpecWorkspaceExecutionJob(ctx, job); err != nil {
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
					have, err := s.GetBatchSpecWorkspaceExecutionJob(ctx, GetBatchSpecWorkspaceExecutionJobOpts{ID: job.ID})
					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, job); diff != "" {
						t.Fatal(diff)
					}
				})
			}
		})

		t.Run("GetByBatchSpecWorkspaceID", func(t *testing.T) {
			for i, job := range jobs {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					have, err := s.GetBatchSpecWorkspaceExecutionJob(ctx, GetBatchSpecWorkspaceExecutionJobOpts{BatchSpecWorkspaceID: job.BatchSpecWorkspaceID})
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
			opts := GetBatchSpecWorkspaceExecutionJobOpts{ID: 0xdeadbeef}

			_, have := s.GetBatchSpecWorkspaceExecutionJob(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		for i, job := range jobs {
			job.WorkerHostname = fmt.Sprintf("worker-hostname-%d", i)
			switch i {
			case 0:
				job.State = btypes.BatchSpecWorkspaceExecutionJobStateQueued
			case 1:
				job.State = btypes.BatchSpecWorkspaceExecutionJobStateProcessing
			case 2:
				job.State = btypes.BatchSpecWorkspaceExecutionJobStateFailed
			}

			if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_workspace_execution_jobs SET worker_hostname = %s, state = %s WHERE id = %s", job.WorkerHostname, job.State, job.ID)); err != nil {
				t.Fatal(err)
			}
		}

		t.Run("All", func(t *testing.T) {
			have, err := s.ListBatchSpecWorkspaceExecutionJobs(ctx, ListBatchSpecWorkspaceExecutionJobsOpts{})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(have, jobs); diff != "" {
				t.Fatalf("invalid jobs returned: %s", diff)
			}
		})

		t.Run("WorkerHostname", func(t *testing.T) {
			for _, job := range jobs {
				have, err := s.ListBatchSpecWorkspaceExecutionJobs(ctx, ListBatchSpecWorkspaceExecutionJobsOpts{
					WorkerHostname: job.WorkerHostname,
				})
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(have, []*btypes.BatchSpecWorkspaceExecutionJob{job}); diff != "" {
					t.Fatalf("invalid batch spec workspace jobs returned: %s", diff)
				}
			}
		})

		t.Run("State", func(t *testing.T) {
			for _, job := range jobs {
				have, err := s.ListBatchSpecWorkspaceExecutionJobs(ctx, ListBatchSpecWorkspaceExecutionJobsOpts{
					State: job.State,
				})
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(have, []*btypes.BatchSpecWorkspaceExecutionJob{job}); diff != "" {
					t.Fatalf("invalid batch spec workspace jobs returned: %s", diff)
				}
			}
		})
	})
}
