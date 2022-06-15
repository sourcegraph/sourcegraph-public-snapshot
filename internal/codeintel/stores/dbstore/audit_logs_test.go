package dbstore

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestUploadAuditLogs(t *testing.T) {
	db := database.NewDB(dbtest.NewDB(t))
	store := testStore(db)

	insertUploads(t, db, Upload{ID: 1})
	updateUploads(t, db, Upload{ID: 1, State: "deleting"})

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
