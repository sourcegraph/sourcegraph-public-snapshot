package ratelimit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

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
		err := SyncServices(ctx, svcs)
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
