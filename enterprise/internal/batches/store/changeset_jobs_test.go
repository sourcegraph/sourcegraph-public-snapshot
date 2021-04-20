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

func testStoreChangesetJobs(t *testing.T, ctx context.Context, s *Store, clock ct.Clock) {
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

	jobs := make([]*btypes.ChangesetJob, 0, 3)
	for i := 0; i < cap(jobs); i++ {
		c := &btypes.ChangesetJob{
			UserID:        int32(i + 1234),
			BatchChangeID: int64(i + 910),
			ChangesetID:   changeset.ID,
			JobType:       btypes.ChangesetJobTypeComment,
		}

		if i == cap(jobs)-1 {
			c.ChangesetID = changesetWithDeletedRepo.ID
		}
		jobs = append(jobs, c)
	}

	t.Run("Create", func(t *testing.T) {

		for _, c := range jobs {
			want := *c
			have := c

			err := s.CreateChangesetJob(ctx, have)
			if err != nil {
				t.Fatal(err)
			}

			if have.ID == 0 {
				t.Fatal("ID should not be zero")
			}

			want.ID = have.ID
			want.Payload = &btypes.ChangesetJobCommentPayload{}
			want.CreatedAt = clock.Now()
			want.UpdatedAt = clock.Now()

			if diff := cmp.Diff(have, &want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		for i, job := range jobs {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				have, err := s.GetChangesetJob(ctx, GetChangesetJobOpts{ID: job.ID})
				if i == cap(jobs)-1 {
					if err != ErrNoResults {
						t.Fatal("unexpected non-no-results error")
					}
					return
				} else if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(have, job); diff != "" {
					t.Fatal(diff)
				}
			})
		}

		t.Run("NoResults", func(t *testing.T) {
			opts := GetChangesetJobOpts{ID: 0xdeadbeef}

			_, have := s.GetChangesetJob(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})
}
