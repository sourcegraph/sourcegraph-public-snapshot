package store

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/batches"
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
		for idx, job := range jobs {
			if err := ct.CreateBatchSpecWorkspaceExecutionJob(ctx, s, ScanBatchSpecWorkspaceExecutionJob, job); err != nil {
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

			job.PlaceInQueue = int64(idx + 1)
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
				job.Cancel = true
				job.PlaceInQueue = 1
			case 1:
				job.State = btypes.BatchSpecWorkspaceExecutionJobStateProcessing
				job.PlaceInQueue = 0
			case 2:
				job.State = btypes.BatchSpecWorkspaceExecutionJobStateFailed
				job.PlaceInQueue = 0
			}

			ct.UpdateJobState(t, ctx, s, job)
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

		t.Run("IDs", func(t *testing.T) {
			for _, job := range jobs {
				have, err := s.ListBatchSpecWorkspaceExecutionJobs(ctx, ListBatchSpecWorkspaceExecutionJobsOpts{
					IDs: []int64{job.ID},
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

	t.Run("CancelBatchSpecWorkspaceExecutionJobs", func(t *testing.T) {
		t.Run("single job by ID", func(t *testing.T) {
			opts := CancelBatchSpecWorkspaceExecutionJobsOpts{IDs: []int64{jobs[0].ID}}

			t.Run("Queued", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_workspace_execution_jobs SET state = 'queued', started_at = NULL, finished_at = NULL WHERE id = %s", jobs[0].ID)); err != nil {
					t.Fatal(err)
				}
				records, err := s.CancelBatchSpecWorkspaceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}
				record := records[0]
				if have, want := record.State, btypes.BatchSpecWorkspaceExecutionJobStateFailed; have != want {
					t.Errorf("invalid state: have=%q want=%q", have, want)
				}
				if have, want := record.Cancel, true; have != want {
					t.Errorf("invalid cancel value: have=%t want=%t", have, want)
				}
				if record.FinishedAt.IsZero() {
					t.Error("finished_at not set")
				} else if have, want := record.FinishedAt, s.now(); !have.Equal(want) {
					t.Errorf("invalid finished_at: have=%s want=%s", have, want)
				}
				if have, want := record.UpdatedAt, s.now(); !have.Equal(want) {
					t.Errorf("invalid updated_at: have=%s want=%s", have, want)
				}
			})

			t.Run("Processing", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_workspace_execution_jobs SET state = 'processing', started_at = now(), finished_at = NULL WHERE id = %s", jobs[0].ID)); err != nil {
					t.Fatal(err)
				}
				records, err := s.CancelBatchSpecWorkspaceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}
				record := records[0]
				if have, want := record.State, btypes.BatchSpecWorkspaceExecutionJobStateProcessing; have != want {
					t.Errorf("invalid state: have=%q want=%q", have, want)
				}
				if have, want := record.Cancel, true; have != want {
					t.Errorf("invalid cancel value: have=%t want=%t", have, want)
				}
				if !record.FinishedAt.IsZero() {
					t.Error("finished_at set")
				}
				if have, want := record.UpdatedAt, s.now(); !have.Equal(want) {
					t.Errorf("invalid updated_at: have=%s want=%s", have, want)
				}
			})

			t.Run("already completed", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_workspace_execution_jobs SET state = 'completed' WHERE id = %s", jobs[0].ID)); err != nil {
					t.Fatal(err)
				}
				records, err := s.CancelBatchSpecWorkspaceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}
				if len(records) != 0 {
					t.Fatalf("unexpected records returned: %d", len(records))
				}
			})

			t.Run("still queued", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_workspace_execution_jobs SET state = 'queued' WHERE id = %s", jobs[0].ID)); err != nil {
					t.Fatal(err)
				}
				records, err := s.CancelBatchSpecWorkspaceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}
				if len(records) != 1 {
					t.Fatalf("unexpected records returned: %d", len(records))
				}
			})
		})

		t.Run("multiple jobs by BatchSpecID", func(t *testing.T) {
			spec := &btypes.BatchSpec{UserID: 1234, NamespaceUserID: 4567}
			if err := s.CreateBatchSpec(ctx, spec); err != nil {
				t.Fatal(err)
			}

			var specJobIDs []int64
			for i := 0; i < 3; i++ {
				ws := &btypes.BatchSpecWorkspace{BatchSpecID: spec.ID, RepoID: api.RepoID(i)}
				if err := s.CreateBatchSpecWorkspace(ctx, ws); err != nil {
					t.Fatal(err)
				}

				job := &btypes.BatchSpecWorkspaceExecutionJob{BatchSpecWorkspaceID: ws.ID}
				if err := ct.CreateBatchSpecWorkspaceExecutionJob(ctx, s, ScanBatchSpecWorkspaceExecutionJob, job); err != nil {
					t.Fatal(err)
				}
				specJobIDs = append(specJobIDs, job.ID)
			}

			opts := CancelBatchSpecWorkspaceExecutionJobsOpts{BatchSpecID: spec.ID}

			t.Run("Queued", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_workspace_execution_jobs SET state = 'queued', started_at = NULL, finished_at = NULL WHERE id = ANY(%s)", pq.Array(specJobIDs))); err != nil {
					t.Fatal(err)
				}
				records, err := s.CancelBatchSpecWorkspaceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}
				record := records[0]
				if have, want := record.State, btypes.BatchSpecWorkspaceExecutionJobStateFailed; have != want {
					t.Errorf("invalid state: have=%q want=%q", have, want)
				}
				if have, want := record.Cancel, true; have != want {
					t.Errorf("invalid cancel value: have=%t want=%t", have, want)
				}
				if record.FinishedAt.IsZero() {
					t.Error("finished_at not set")
				} else if have, want := record.FinishedAt, s.now(); !have.Equal(want) {
					t.Errorf("invalid finished_at: have=%s want=%s", have, want)
				}
				if have, want := record.UpdatedAt, s.now(); !have.Equal(want) {
					t.Errorf("invalid updated_at: have=%s want=%s", have, want)
				}
			})

			t.Run("Processing", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_workspace_execution_jobs SET state = 'processing', started_at = now(), finished_at = NULL WHERE id = ANY(%s)", pq.Array(specJobIDs))); err != nil {
					t.Fatal(err)
				}
				records, err := s.CancelBatchSpecWorkspaceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}
				record := records[0]
				if have, want := record.State, btypes.BatchSpecWorkspaceExecutionJobStateProcessing; have != want {
					t.Errorf("invalid state: have=%q want=%q", have, want)
				}
				if have, want := record.Cancel, true; have != want {
					t.Errorf("invalid cancel value: have=%t want=%t", have, want)
				}
				if !record.FinishedAt.IsZero() {
					t.Error("finished_at set")
				}
				if have, want := record.UpdatedAt, s.now(); !have.Equal(want) {
					t.Errorf("invalid updated_at: have=%s want=%s", have, want)
				}
			})

			t.Run("Already completed", func(t *testing.T) {
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_workspace_execution_jobs SET state = 'completed' WHERE id = ANY(%s)", pq.Array(specJobIDs))); err != nil {
					t.Fatal(err)
				}
				records, err := s.CancelBatchSpecWorkspaceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}
				if len(records) != 0 {
					t.Fatalf("unexpected records returned: %d", len(records))
				}
			})

			t.Run("subset processing, subset completed", func(t *testing.T) {
				completed := specJobIDs[1:]
				processing := specJobIDs[0:1]
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_workspace_execution_jobs SET state = 'processing', started_at = now(), finished_at = NULL WHERE id = ANY(%s)", pq.Array(processing))); err != nil {
					t.Fatal(err)
				}
				if err := s.Exec(ctx, sqlf.Sprintf("UPDATE batch_spec_workspace_execution_jobs SET state = 'completed' WHERE id = ANY(%s)", pq.Array(completed))); err != nil {
					t.Fatal(err)
				}
				records, err := s.CancelBatchSpecWorkspaceExecutionJobs(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}
				if have, want := len(records), len(processing); have != want {
					t.Fatalf("wrong number of canceled records. have=%d, want=%d", have, want)
				}
			})
		})
	})

	t.Run("SetBatchSpecWorkspaceExecutionJobAccessToken", func(t *testing.T) {
		err := s.SetBatchSpecWorkspaceExecutionJobAccessToken(ctx, jobs[0].ID, 12345)
		if err != nil {
			t.Fatal(err)
		}

		reloadedJob, err := s.GetBatchSpecWorkspaceExecutionJob(ctx, GetBatchSpecWorkspaceExecutionJobOpts{ID: jobs[0].ID})
		if err != nil {
			t.Fatal(err)
		}

		if reloadedJob.AccessTokenID != 12345 {
			t.Fatalf("wrong access token ID: %d", reloadedJob.AccessTokenID)
		}
	})

	t.Run("CreateBatchSpecWorkspaceExecutionJobs", func(t *testing.T) {
		singleStep := []batches.Step{{Run: "echo lol", Container: "alpine:3"}}
		createWorkspaces := func(t *testing.T, batchSpec *btypes.BatchSpec, workspaces ...*btypes.BatchSpecWorkspace) {
			t.Helper()

			batchSpec.NamespaceUserID = 1
			if err := s.CreateBatchSpec(ctx, batchSpec); err != nil {
				t.Fatal(err)
			}

			for i, workspace := range workspaces {
				workspace.BatchSpecID = batchSpec.ID
				workspace.RepoID = 1
				workspace.Branch = fmt.Sprintf("refs/heads/main-%d", i)
				workspace.Commit = fmt.Sprintf("commit-%d", i)
			}

			if err := s.CreateBatchSpecWorkspace(ctx, workspaces...); err != nil {
				t.Fatal(err)
			}
		}

		createJobsAndAssert := func(t *testing.T, batchSpec *btypes.BatchSpec, wantJobsForWorkspaces []int64) {
			t.Helper()

			err := s.CreateBatchSpecWorkspaceExecutionJobs(ctx, batchSpec.ID)
			if err != nil {
				t.Fatal(err)
			}

			jobs, err := s.ListBatchSpecWorkspaceExecutionJobs(ctx, ListBatchSpecWorkspaceExecutionJobsOpts{
				BatchSpecWorkspaceIDs: wantJobsForWorkspaces,
			})
			if err != nil {
				t.Fatal(err)
			}

			if have, want := len(jobs), len(wantJobsForWorkspaces); have != want {
				t.Fatalf("wrong number of execution jobs created. want=%d, have=%d", want, have)
			}
		}

		t.Run("success", func(t *testing.T) {
			normalWorkspace := &btypes.BatchSpecWorkspace{Steps: singleStep}
			ignoredWorkspace := &btypes.BatchSpecWorkspace{Steps: singleStep, Ignored: true}
			unsupportedWorkspace := &btypes.BatchSpecWorkspace{Steps: singleStep, Unsupported: true}
			noStepsWorkspace := &btypes.BatchSpecWorkspace{Steps: []batches.Step{}}

			batchSpec := &btypes.BatchSpec{}

			createWorkspaces(t, batchSpec, normalWorkspace, ignoredWorkspace, unsupportedWorkspace, noStepsWorkspace)
			createJobsAndAssert(t, batchSpec, []int64{normalWorkspace.ID})
		})

		t.Run("allowIgnored", func(t *testing.T) {
			normalWorkspace := &btypes.BatchSpecWorkspace{Steps: singleStep}
			ignoredWorkspace := &btypes.BatchSpecWorkspace{Steps: singleStep, Ignored: true}

			batchSpec := &btypes.BatchSpec{AllowIgnored: true}

			createWorkspaces(t, batchSpec, normalWorkspace, ignoredWorkspace)
			createJobsAndAssert(t, batchSpec, []int64{normalWorkspace.ID, ignoredWorkspace.ID})
		})

		t.Run("allowUnsupported", func(t *testing.T) {
			normalWorkspace := &btypes.BatchSpecWorkspace{Steps: singleStep}
			unsupportedWorkspace := &btypes.BatchSpecWorkspace{Steps: singleStep, Unsupported: true}

			batchSpec := &btypes.BatchSpec{AllowUnsupported: true}

			createWorkspaces(t, batchSpec, normalWorkspace, unsupportedWorkspace)
			createJobsAndAssert(t, batchSpec, []int64{normalWorkspace.ID, unsupportedWorkspace.ID})
		})

		t.Run("allowUnsupported and allowIgnored", func(t *testing.T) {
			normalWorkspace := &btypes.BatchSpecWorkspace{Steps: singleStep}
			ignoredWorkspace := &btypes.BatchSpecWorkspace{Steps: singleStep, Ignored: true}
			unsupportedWorkspace := &btypes.BatchSpecWorkspace{Steps: singleStep, Unsupported: true}
			noStepsWorkspace := &btypes.BatchSpecWorkspace{Steps: []batches.Step{}}

			batchSpec := &btypes.BatchSpec{AllowUnsupported: true, AllowIgnored: true}

			createWorkspaces(t, batchSpec, normalWorkspace, ignoredWorkspace, unsupportedWorkspace, noStepsWorkspace)
			createJobsAndAssert(t, batchSpec, []int64{normalWorkspace.ID, ignoredWorkspace.ID, unsupportedWorkspace.ID})
		})
	})
}
