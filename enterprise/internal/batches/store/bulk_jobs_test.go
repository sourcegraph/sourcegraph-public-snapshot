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
)

func testStoreBulkJobs(t *testing.T, ctx context.Context, s *Store, clock ct.Clock) {
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

	changeset := ct.CreateChangeset(t, ctx, s, ct.TestChangesetOpts{Repo: repo.ID})
	changesetWithDeletedRepo := ct.CreateChangeset(t, ctx, s, ct.TestChangesetOpts{Repo: deletedRepo.ID})
	var batchChangeID int64 = 12345

	failureMessage := "bad error"
	jobs := make([]*btypes.ChangesetJob, 0, 3)
	bulkJobs := make([]*btypes.BulkJob, 0, 2)
	for i := 0; i < cap(jobs); i++ {
		groupID, err := RandomID()
		if err != nil {
			t.Fatal(err)
		}
		c := &btypes.ChangesetJob{
			BulkGroup:     groupID,
			UserID:        int32(i + 1234),
			BatchChangeID: batchChangeID,
			ChangesetID:   changeset.ID,
			State:         btypes.ChangesetJobStateQueued,
			JobType:       btypes.ChangesetJobTypeComment,
		}

		if i == cap(jobs)-1 {
			c.ChangesetID = changesetWithDeletedRepo.ID
		}
		if i == 0 {
			c.State = btypes.ChangesetJobStateFailed
			failureMessage := "bad error"
			c.FailureMessage = &failureMessage
		}
		jobs = append(jobs, c)
	}
	err := s.CreateChangesetJob(ctx, jobs...)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < cap(bulkJobs); i++ {
		j := &btypes.BulkJob{
			ID:        jobs[i].BulkGroup,
			DBID:      jobs[i].ID,
			State:     btypes.BulkJobStateProcessing,
			Type:      btypes.ChangesetJobTypeComment,
			CreatedAt: clock.Now(),
		}
		if i == 0 {
			j.Progress = 1
			j.State = btypes.BulkJobStateFailed
		}
		bulkJobs = append(bulkJobs, j)
	}

	t.Run("Get", func(t *testing.T) {
		for i, job := range jobs {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				have, err := s.GetBulkJob(ctx, GetBulkJobOpts{ID: job.BulkGroup})
				if i == cap(jobs)-1 {
					if err != ErrNoResults {
						t.Fatal("unexpected non-no-results error")
					}
					return
				} else if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(have, bulkJobs[i]); diff != "" {
					t.Fatal(diff)
				}
			})
		}

		t.Run("NoResults", func(t *testing.T) {
			opts := GetBulkJobOpts{ID: "deadbeef"}

			_, have := s.GetBulkJob(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("Count", func(t *testing.T) {
		t.Run("All", func(t *testing.T) {
			want := len(bulkJobs)
			have, err := s.CountBulkJobs(ctx, CountBulkJobsOpts{BatchChangeID: batchChangeID})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := CountBulkJobsOpts{BatchChangeID: -1}

			have, err := s.CountBulkJobs(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}
			want := 0

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("NoLimit", func(t *testing.T) {
			// Empty limit should return all entries.
			opts := ListBulkJobsOpts{BatchChangeID: batchChangeID}
			ts, next, err := s.ListBulkJobs(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := next, int64(0); have != want {
				t.Fatalf("opts: %+v: have next %v, want %v", opts, have, want)
			}

			have, want := ts, bulkJobs
			if len(have) != len(want) {
				t.Fatalf("listed %d bulk jobs, want: %d", len(have), len(want))
			}

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatalf("opts: %+v, diff: %s", opts, diff)
			}
		})

		t.Run("WithLimit", func(t *testing.T) {
			for i := 1; i <= len(bulkJobs); i++ {
				cs, next, err := s.ListBulkJobs(ctx, ListBulkJobsOpts{BatchChangeID: batchChangeID, LimitOpts: LimitOpts{Limit: i}})
				if err != nil {
					t.Fatal(err)
				}

				{
					have, want := next, int64(0)
					if i < len(bulkJobs) {
						want = bulkJobs[i].DBID
					}

					if have != want {
						t.Fatalf("limit: %v: have next %v, want %v", i, have, want)
					}
				}

				{
					have, want := cs, bulkJobs[:i]
					if len(have) != len(want) {
						t.Fatalf("listed %d bulkJobs, want: %d", len(have), len(want))
					}

					if diff := cmp.Diff(have, want); diff != "" {
						t.Fatal(diff)
					}
				}
			}
		})

		t.Run("WithLimitAndCursor", func(t *testing.T) {
			var cursor int64
			for i := 1; i <= len(bulkJobs); i++ {
				opts := ListBulkJobsOpts{BatchChangeID: batchChangeID, Cursor: cursor, LimitOpts: LimitOpts{Limit: 1}}
				have, next, err := s.ListBulkJobs(ctx, opts)
				if err != nil {
					t.Fatal(err)
				}

				want := bulkJobs[i-1 : i]
				if diff := cmp.Diff(have, want); diff != "" {
					t.Fatalf("opts: %+v, diff: %s", opts, diff)
				}

				cursor = next
			}
		})
	})

	t.Run("ListBulkJobErrors", func(t *testing.T) {
		for i, job := range jobs {
			errors, err := s.ListBulkJobErrors(ctx, ListBulkJobErrorsOpts{
				BulkJobID: job.BulkGroup,
			})
			if err != nil {
				t.Fatal(err)
			}
			if i != 0 {
				if have, want := len(errors), 0; have != want {
					t.Fatalf("invalid amount of errors returned, want=%d have=%d", want, have)
				}
				continue
			}
			have := errors
			want := []*btypes.BulkJobError{
				{
					ChangesetID: changeset.ID,
					Error:       failureMessage,
				},
			}
			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})
}
