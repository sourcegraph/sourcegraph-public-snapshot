package store

import (
	"context"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"

	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func testExternalServices(t *testing.T, ctx context.Context, store *Store, clock bt.Clock) {
	// OK, let's set up a useful scenario. We're going to set up four external
	// services. Then we'll set up three batch changes:
	//
	// 1. A batch change that has changesets on three of them, including some
	// duplicates, and that will allow us to test that the right external
	// services are returned, and that pagination works as expected.
	//
	// 2. We'll set up another batch change on the other external service to
	// ensure that we filter correctly to the batch change we're interested in.
	//
	// 3. We'll set up an empty batch change to make sure that we don't error
	// unexpectedly if there are no changesets.
	//
	// We'll then also set up some permissions, which will basically make the
	// GitLab repo invisible to the normal user.
	//
	// TODO: probably reuse this logic for the resolver tests.
	ghRepos, ghExtSvc := bt.CreateGitHubSSHTestRepos(t, ctx, store.DB(), 2)
	glRepos, glExtSvc := bt.CreateGitlabTestRepos(t, ctx, store.DB(), 1)
	bbsRepos, bbsExtSvc := bt.CreateBbsSSHTestRepos(t, ctx, store.DB(), 1)
	unusedRepos, unusedExtSvc := bt.CreateGitHubSSHTestRepos(t, ctx, store.DB(), 1)

	user := bt.CreateTestUser(t, store.DB(), false)
	admin := bt.CreateTestUser(t, store.DB(), true)

	batchSpec := bt.CreateBatchSpec(t, ctx, store, "external-services", user.ID)
	batchChange := bt.CreateBatchChange(t, ctx, store, "external-services", user.ID, batchSpec.ID)

	repos := []*types.Repo{ghRepos[0], ghRepos[1], glRepos[0], bbsRepos[0]}
	changesets := make([]*btypes.Changeset, len(repos))
	for i, repo := range repos {
		spec := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
			User:      user.ID,
			Repo:      repo.ID,
			BatchSpec: batchSpec.ID,
			HeadRef:   "main",
			Published: true,
		})
		changesets[i] = bt.CreateChangeset(t, ctx, store, bt.TestChangesetOpts{
			Repo:             repo.ID,
			BatchChange:      batchChange.ID,
			CurrentSpec:      spec.ID,
			PublicationState: btypes.ChangesetPublicationStatePublished,
		})
	}

	otherBatchSpec := bt.CreateBatchSpec(t, ctx, store, "other", user.ID)
	otherBatchChange := bt.CreateBatchChange(t, ctx, store, "other", user.ID, otherBatchSpec.ID)
	otherChangesetSpec := bt.CreateChangesetSpec(t, ctx, store, bt.TestSpecOpts{
		User:      user.ID,
		Repo:      unusedRepos[0].ID,
		BatchSpec: otherBatchSpec.ID,
		HeadRef:   "main",
		Published: true,
	})
	bt.CreateChangeset(t, ctx, store, bt.TestChangesetOpts{
		Repo:             unusedRepos[0].ID,
		BatchChange:      otherBatchChange.ID,
		CurrentSpec:      otherChangesetSpec.ID,
		PublicationState: btypes.ChangesetPublicationStatePublished,
	})

	emptyBatchSpec := bt.CreateBatchSpec(t, ctx, store, "empty", user.ID)
	emptyBatchChange := bt.CreateBatchChange(t, ctx, store, "empty", user.ID, emptyBatchSpec.ID)

	// Set the GitLab and GitHub repos to be private, and their associated
	// external services to be restricted. We need to do these with raw queries
	// because the external service store will recalculate Unrestricted based on
	// the configuration, and because the repo store doesn't provide an update
	// or upsert method.
	store.Exec(ctx, sqlf.Sprintf("UPDATE external_services SET unrestricted = FALSE WHERE id IN (%d, %d)", ghExtSvc.ID, glExtSvc.ID))
	store.Exec(ctx, sqlf.Sprintf("UPDATE repo SET private = TRUE where id IN (%d, %d)", ghRepos[0].ID, glRepos[0].ID))

	// For comparison purposes later, though, we need the restricted external
	// services to match what we just did in the database.
	ghExtSvc.Unrestricted = false
	glExtSvc.Unrestricted = false

	// Set up the repo permissions to allow the regular user access to GitHub,
	// but not GitLab.
	bt.MockRepoPermissions(t, store.DB(), user.ID, ghRepos[0].ID)

	// Awesome. Now let's run some actual test cases.
	t.Run("CountExternalServicesForBatchChange", func(t *testing.T) {
		for name, tc := range map[string]struct {
			batchChangeID int64
			user          *types.User
			wantCount     int64
		}{
			"invalid ID": {
				batchChangeID: otherBatchChange.ID + 1,
				user:          admin,
				wantCount:     0,
			},
			"empty": {
				batchChangeID: emptyBatchChange.ID,
				user:          admin,
				wantCount:     0,
			},
			"other": {
				batchChangeID: otherBatchChange.ID,
				user:          admin,
				wantCount:     1,
			},
			"primary as admin": {
				batchChangeID: batchChange.ID,
				user:          admin,
				wantCount:     3,
			},
			"primary as user": {
				batchChangeID: batchChange.ID,
				user:          user,
				wantCount:     2,
			},
		} {
			t.Run(name, func(t *testing.T) {
				ctx := actor.WithActor(ctx, actor.FromUser(tc.user.ID))

				count, err := store.CountExternalServicesForBatchChange(ctx, tc.batchChangeID)
				assert.Nil(t, err)
				assert.EqualValues(t, tc.wantCount, count)
			})
		}

		t.Run("DB error", func(t *testing.T) {
			store := NewWithClock(dbtesting.NewErrorDB(nil), nil, nil, store.Clock())

			_, err := store.CountExternalServicesForBatchChange(ctx, batchChange.ID)
			assert.True(t, errors.Is(err, ErrNoResults))
		})
	})

	t.Run("ListExternalServicesForBatchChange", func(t *testing.T) {
		for name, tc := range map[string]struct {
			batchChangeID int64
			user          *types.User
			want          []*types.ExternalService
		}{
			"invalid ID": {
				batchChangeID: otherBatchChange.ID + 1,
				user:          admin,
				want:          []*types.ExternalService{},
			},
			"empty": {
				batchChangeID: emptyBatchChange.ID,
				user:          admin,
				want:          []*types.ExternalService{},
			},
			"other": {
				batchChangeID: otherBatchChange.ID,
				user:          admin,
				want:          []*types.ExternalService{unusedExtSvc},
			},
			"primary as admin": {
				batchChangeID: batchChange.ID,
				user:          admin,
				want:          []*types.ExternalService{ghExtSvc, glExtSvc, bbsExtSvc},
			},
			"primary as user": {
				batchChangeID: batchChange.ID,
				user:          user,
				want:          []*types.ExternalService{ghExtSvc, bbsExtSvc},
			}} {
			t.Run(name, func(t *testing.T) {
				ctx := actor.WithActor(ctx, actor.FromUser(tc.user.ID))

				have := []*types.ExternalService{}
				for cursor := int64(0); ; {
					page, next, err := store.ListExternalServicesForBatchChange(ctx, ListExternalServicesForBatchChangeOpts{
						LimitOpts:     LimitOpts{Limit: 1},
						Cursor:        cursor,
						BatchChangeID: tc.batchChangeID,
					})
					assert.Nil(t, err)
					assert.LessOrEqual(t, len(page), 1)
					have = append(have, page...)

					if next == 0 {
						break
					}
					cursor = next
				}

				assert.Equal(t, tc.want, have)
			})
		}

		t.Run("DB error", func(t *testing.T) {
			wantErr := errors.New("test error")
			store := NewWithClock(dbtesting.NewErrorDB(wantErr), nil, nil, store.Clock())

			_, _, err := store.ListExternalServicesForBatchChange(ctx, ListExternalServicesForBatchChangeOpts{
				BatchChangeID: batchChange.ID,
			})
			assert.True(t, errors.Is(err, wantErr))
		})
	})
}
