package dbstore

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestDeleteOldAuditLogs(t *testing.T) {
	sqlDB := dbtest.NewDB(t)
	db := database.NewDB(sqlDB)
	store := testStore(db)

	// Sanity check for syntax only
	if _, err := store.DeleteOldAuditLogs(context.Background(), time.Second, time.Now()); err != nil {
		t.Fatalf("unexpected error deleting old audit logs: %s", err)
	}
}
