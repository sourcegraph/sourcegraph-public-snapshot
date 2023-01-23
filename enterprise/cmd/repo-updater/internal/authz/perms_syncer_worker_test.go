package authz

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

func TestPermsSyncerWorker_Handle(t *testing.T) {
	ctx := context.Background()

	dummySyncer := &dummyPermsSyncer{}

	worker := MakePermsSyncerWorker(ctx, &observation.TestContext, dummySyncer)

	t.Run("user sync request", func(t *testing.T) {
		worker.Handle(ctx, logtest.Scoped(t), &database.PermissionSyncJob{
			ID:               99,
			UserID:           1234,
			InvalidateCaches: true,
			HighPriority:     true,
		})

		wantMeta := &requestMeta{
			ID:       1234,
			Priority: priorityHigh,
			Type:     requestTypeUser,
			Options: authz.FetchPermsOptions{
				InvalidateCaches: true,
			},
		}

		if diff := cmp.Diff(dummySyncer.request.requestMeta, wantMeta); diff != "" {
			t.Fatalf("wrong sync request: %s", diff)
		}
	})

	t.Run("repo sync request", func(t *testing.T) {
		worker.Handle(ctx, logtest.Scoped(t), &database.PermissionSyncJob{
			ID:               777,
			RepositoryID:     4567,
			InvalidateCaches: false,
			HighPriority:     false,
		})

		wantMeta := &requestMeta{
			ID:       4567,
			Priority: priorityLow,
			Type:     requestTypeRepo,
			Options: authz.FetchPermsOptions{
				InvalidateCaches: false,
			},
		}

		if diff := cmp.Diff(dummySyncer.request.requestMeta, wantMeta); diff != "" {
			t.Fatalf("wrong sync request: %s", diff)
		}
	})
}

type dummyPermsSyncer struct {
	request *syncRequest
}

func (d *dummyPermsSyncer) syncPerms(ctx context.Context, syncGroups map[requestType]group.ContextGroup, request *syncRequest) {
	d.request = request
}
