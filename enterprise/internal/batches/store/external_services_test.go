package store

import (
	"context"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"

	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func testExternalServices(t *testing.T, ctx context.Context, store *Store, clock bt.Clock) {
	fixture := bt.NewExternalServiceWebhookFixture(t, ctx, store)

	// Awesome. Now let's run some actual test cases.
	t.Run("CountExternalServicesForBatchChange", func(t *testing.T) {
		for name, tc := range map[string]struct {
			batchChangeID int64
			user          *types.User
			wantCount     int64
		}{
			"invalid ID": {
				batchChangeID: fixture.OtherBatchChange.ID + 1,
				user:          fixture.AdminUser,
				wantCount:     0,
			},
			"empty": {
				batchChangeID: fixture.EmptyBatchChange.ID,
				user:          fixture.AdminUser,
				wantCount:     0,
			},
			"other": {
				batchChangeID: fixture.OtherBatchChange.ID,
				user:          fixture.AdminUser,
				wantCount:     1,
			},
			"primary as fixture.AdminUser": {
				batchChangeID: fixture.PrimaryBatchChange.ID,
				user:          fixture.AdminUser,
				wantCount:     3,
			},
			"primary as user": {
				batchChangeID: fixture.PrimaryBatchChange.ID,
				user:          fixture.RegularUser,
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

			_, err := store.CountExternalServicesForBatchChange(ctx, fixture.PrimaryBatchChange.ID)
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
				batchChangeID: fixture.OtherBatchChange.ID + 1,
				user:          fixture.AdminUser,
				want:          []*types.ExternalService{},
			},
			"empty": {
				batchChangeID: fixture.EmptyBatchChange.ID,
				user:          fixture.AdminUser,
				want:          []*types.ExternalService{},
			},
			"other": {
				batchChangeID: fixture.OtherBatchChange.ID,
				user:          fixture.AdminUser,
				want:          []*types.ExternalService{fixture.OtherSvc},
			},
			"primary as fixture.AdminUser": {
				batchChangeID: fixture.PrimaryBatchChange.ID,
				user:          fixture.AdminUser,
				want:          []*types.ExternalService{fixture.GitHubSvc, fixture.GitLabSvc, fixture.BitbucketServerSvc},
			},
			"primary as user": {
				batchChangeID: fixture.PrimaryBatchChange.ID,
				user:          fixture.RegularUser,
				want:          []*types.ExternalService{fixture.GitHubSvc, fixture.BitbucketServerSvc},
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
				BatchChangeID: fixture.PrimaryBatchChange.ID,
			})
			assert.True(t, errors.Is(err, wantErr))
		})
	})
}
