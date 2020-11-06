package repos_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/internal"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/schema"
)

func testSyncRateLimiters(t *testing.T, store *internal.Store) func(*testing.T) {
	clock := dbtesting.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	return func(t *testing.T) {
		ctx := context.Background()
		internal.Transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			toCreate := 501 // Larger than default page size in order to test pagination
			services := make([]*types.ExternalService, 0, toCreate)
			for i := 0; i < toCreate; i++ {
				svc := &types.ExternalService{
					ID:          int64(i) + 1,
					Kind:        "GitHub",
					DisplayName: "GitHub",
					CreatedAt:   now,
					UpdatedAt:   now,
					DeletedAt:   time.Time{},
				}
				config := schema.GitLabConnection{
					Url: fmt.Sprintf("http://example%d.com/", i),
					RateLimit: &schema.GitLabRateLimit{
						RequestsPerHour: 3600,
						Enabled:         true,
					},
				}
				data, err := json.Marshal(config)
				if err != nil {
					t.Fatal(err)
				}
				svc.Config = string(data)
				services = append(services, svc)
			}

			for _, svc := range services {
				if err := tx.ExternalServiceStore().Upsert(ctx, svc); err != nil {
					t.Fatalf("failed to setup store: %v", err)
				}
			}

			registry := ratelimit.NewRegistry()
			syncer := repos.NewRateLimitSyncer(registry, tx.ExternalServiceStore())
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
}
