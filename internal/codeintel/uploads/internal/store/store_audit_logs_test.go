package store

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestUploadAuditLogs(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(db, &observation.TestContext)

	insertUploads(t, db, shared.Upload{ID: 1})
	updateUploads(t, db, shared.Upload{ID: 1, State: "deleting"})

	logs, err := store.GetAuditLogsForUpload(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error fetching audit logs: %s", err)
	}
	if len(logs) != 2 {
		t.Fatalf("unexpected number of logs. want=%v have=%v", 2, len(logs))
	}

	stateTransition := transitionForColumn(t, "state", logs[1].TransitionColumns)
	if *stateTransition["new"] != "deleting" {
		t.Fatalf("unexpected state column transition values. want=%v got=%v", "deleting", *stateTransition["new"])
	}
}

func transitionForColumn(t *testing.T, key string, transitions []map[string]*string) map[string]*string {
	for _, transition := range transitions {
		if val := transition["column"]; val != nil && *val == key {
			return transition
		}
	}

	t.Fatalf("no transition for key found. key=%s, transitions=%v", key, transitions)
	return nil
}

func TestDeleteOldAuditLogs(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(db, &observation.TestContext)

	// Sanity check for syntax only
	if _, err := store.DeleteOldAuditLogs(context.Background(), time.Second, time.Now()); err != nil {
		t.Fatalf("unexpected error deleting old audit logs: %s", err)
	}
}
