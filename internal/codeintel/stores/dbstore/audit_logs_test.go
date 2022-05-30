package dbstore

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestUploadAuditLogs(t *testing.T) {
	db := dbtest.NewDB(t)
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

	b, e := json.MarshalIndent(logs, "", "\t")
	t.Logf("%s %v", string(b), e)

	stateTransition := transitionForColumn(t, "state", logs[1].TransitionColumns)
	for k, v := range stateTransition {
		t.Logf("key=%v val=%v", k, *v)
	}
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
