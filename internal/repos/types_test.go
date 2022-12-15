package repos

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestReposNamesSummary(t *testing.T) {
	var rps types.Repos

	eid := func(id int) api.ExternalRepoSpec {
		return api.ExternalRepoSpec{
			ID:          strconv.Itoa(id),
			ServiceType: "fake",
			ServiceID:   "https://fake.com",
		}
	}

	for i := 0; i < 5; i++ {
		rps = append(rps, &types.Repo{Name: "bar", ExternalRepo: eid(i)})
	}

	expected := "bar bar bar bar bar"
	ns := rps.NamesSummary()
	if ns != expected {
		t.Errorf("expected %s, got %s", expected, ns)
	}

	rps = nil

	for i := 0; i < 22; i++ {
		rps = append(rps, &types.Repo{Name: "b", ExternalRepo: eid(i)})
	}

	expected = "b b b b b b b b b b b b b b b b b b b b..."
	ns = rps.NamesSummary()
	if ns != expected {
		t.Errorf("expected %s, got %s", expected, ns)
	}
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
		assert.Equal(t, rate.Limit(5000.0/3600.0), gh.Limit())

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
		assert.Equal(t, rate.Limit(5000.0/3600.0), gh.Limit())

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
