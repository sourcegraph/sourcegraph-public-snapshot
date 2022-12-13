package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestPermissionSyncJobsCreate(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	store := PermissionSyncJobsWith(logger, db)

	opts := PermissionSyncJobOpts{HighPriority: true, InvalidateCaches: true}
	err := store.CreateRepoSyncJob(ctx, 99, opts)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
