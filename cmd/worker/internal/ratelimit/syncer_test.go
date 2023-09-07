package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSyncRateLimiters2(t *testing.T) {
	t.Parallel()
	store := getTestRepoStore(t)

	clock := timeutil.NewFakeClock(time.Now(), 0)
	now := clock.Now()
	ctx := context.Background()
	transact(ctx, store, func(t testing.TB, tx repos.Store) {
		toCreate := 501 // Larger than default page size in order to test pagination
		services := make([]*types.ExternalService, 0, toCreate)
		for i := 0; i < toCreate; i++ {
			svc := &types.ExternalService{
				ID:          int64(i) + 1,
				Kind:        "GITLAB",
				DisplayName: "GitLab",
				CreatedAt:   now,
				UpdatedAt:   now,
				DeletedAt:   time.Time{},
				Config:      extsvc.NewEmptyConfig(),
			}
			config := schema.GitLabConnection{
				Token: "abc",
				Url:   fmt.Sprintf("http://example%d.com/", i),
				RateLimit: &schema.GitLabRateLimit{
					RequestsPerHour: 3600,
					Enabled:         true,
				},
				ProjectQuery: []string{
					"None",
				},
			}
			data, err := json.Marshal(config)
			if err != nil {
				t.Fatal(err)
			}
			svc.Config.Set(string(data))
			services = append(services, svc)
		}

		if err := tx.ExternalServiceStore().Upsert(ctx, services...); err != nil {
			t.Fatalf("failed to setup store: %v", err)
		}

		registry := ratelimit.NewRegistry()
		syncer := NewRateLimitSyncer(registry, tx.ExternalServiceStore(), RateLimitSyncerOpts{})
		err := syncer.SyncRateLimiters(ctx)
		if err != nil {
			t.Fatal(err)
		}
		have := registry.Count()
		if have != toCreate {
			t.Fatalf("Want %d, got %d", toCreate, have)
		}
	})(t)
}

func TestSyncRateLimiters(t *testing.T) {
	ctx := context.Background()
	svcs := []*types.ExternalService{
		{
			ID:     1,
			Kind:   extsvc.KindGitHub,
			Config: extsvc.NewEmptyConfig(), // Use default
		},
		{
			ID:     2,
			Kind:   extsvc.KindGitLab,
			Config: extsvc.NewUnencryptedConfig(`{ "rateLimit": {"enabled": true, "requestsPerHour": 10} }`),
		},
	}

	t.Run("sync for all external services", func(t *testing.T) {
		listCalled := 0

		reg := ratelimit.NewRegistry()
		r := &RateLimitSyncer{
			registry: reg,
			serviceLister: &MockExternalServicesLister{
				list: func(ctx context.Context, args database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
					assert.Empty(t, args.IDs)

					listCalled++
					if listCalled > 1 {
						return nil, nil
					}
					return svcs, nil
				},
			},
		}

		err := r.SyncRateLimiters(ctx)
		require.NoError(t, err)

		gh := reg.Get(svcs[0].URN())
		assert.Equal(t, rate.Inf, gh.Limit())

		gl := reg.Get(svcs[1].URN())
		assert.Equal(t, rate.Limit(10.0/3600.0), gl.Limit())
	})

	t.Run("sync for selected external services", func(t *testing.T) {
		listCalled := 0

		reg := ratelimit.NewRegistry()
		r := &RateLimitSyncer{
			registry: reg,
			serviceLister: &MockExternalServicesLister{
				list: func(ctx context.Context, args database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
					assert.Len(t, args.IDs, 1)

					listCalled++
					if listCalled > 1 {
						return nil, nil
					}
					return svcs[:1], nil
				},
			},
		}

		err := r.SyncRateLimiters(ctx, 1)
		require.NoError(t, err)

		gh := reg.Get(svcs[0].URN())
		assert.Equal(t, rate.Inf, gh.Limit())

		// GitLab should have the infinite
		gl := reg.Get(svcs[1].URN())
		assert.Equal(t, rate.Inf, gl.Limit())
	})

	t.Run("limit offset", func(t *testing.T) {
		listCalled := 0

		reg := ratelimit.NewRegistry()
		r := &RateLimitSyncer{
			registry: reg,
			serviceLister: &MockExternalServicesLister{
				list: func(ctx context.Context, args database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
					assert.Equal(t, 1, args.Limit)
					assert.Equal(t, listCalled, args.Offset)

					listCalled++
					if listCalled > 2 {
						return nil, nil
					}
					return svcs[listCalled-1 : listCalled], nil
				},
			},
			pageSize: 1,
		}

		err := r.SyncRateLimiters(ctx)
		require.NoError(t, err)
		assert.Equal(t, 3, listCalled)
	})
}

type MockExternalServicesLister struct {
	list func(context.Context, database.ExternalServicesListOptions) ([]*types.ExternalService, error)
}

func (m MockExternalServicesLister) List(ctx context.Context, args database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
	return m.list(ctx, args)
}
