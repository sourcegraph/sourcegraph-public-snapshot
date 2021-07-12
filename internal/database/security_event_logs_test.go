package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSecurityEventLogs_ValidInfo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()

	var testCases = []struct {
		name  string
		event *SecurityEvent
		err   string
	}{
		{
			name:  "EmptyName",
			event: &SecurityEvent{UserID: 1, URL: "http://sourcegraph.com", Source: "WEB"},
			err:   `INSERT: ERROR: new row for relation "security_event_logs" violates check constraint "security_event_logs_check_name_not_empty" (SQLSTATE 23514)`,
		},
		{
			name:  "InvalidUser",
			event: &SecurityEvent{Name: "test_event", URL: "http://sourcegraph.com", Source: "WEB"},
			err:   `INSERT: ERROR: new row for relation "security_event_logs" violates check constraint "security_event_logs_check_has_user" (SQLSTATE 23514)`,
		},
		{
			name:  "EmptySource",
			event: &SecurityEvent{Name: "test_event", URL: "http://sourcegraph.com", UserID: 1},
			err:   `INSERT: ERROR: new row for relation "security_event_logs" violates check constraint "security_event_logs_check_source_not_empty" (SQLSTATE 23514)`,
		},
		{
			name:  "ValidInsert",
			event: &SecurityEvent{Name: "test_event", UserID: 1, URL: "http://sourcegraph.com", Source: "WEB"},
			err:   "<nil>",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SecurityEventLogs(db).Insert(ctx, tc.event)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("have %+v, want %+v", have, want)
			}
		})
	}
}
