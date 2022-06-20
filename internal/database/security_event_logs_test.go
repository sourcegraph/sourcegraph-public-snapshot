package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSecurityEventLogs_ValidInfo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
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
			name:  "UserAndAnonymousMissing",
			event: &SecurityEvent{Name: "test_event", URL: "http://sourcegraph.com", Source: "WEB", UserID: 0, AnonymousUserID: ""},
			err:   `INSERT: ERROR: new row for relation "security_event_logs" violates check constraint "security_event_logs_check_has_user" (SQLSTATE 23514)`,
		},
		{
			name:  "JustUser",
			event: &SecurityEvent{Name: "test_event", URL: "http://sourcegraph.com", Source: "Web", UserID: 1, AnonymousUserID: ""},
			err:   "<nil>",
		},
		{
			name:  "JustAnonymous",
			event: &SecurityEvent{Name: "test_event", URL: "http://sourcegraph.com", Source: "Web", UserID: 0, AnonymousUserID: "blah"},
			err:   "<nil>",
		},
		{
			name:  "ValidInsert",
			event: &SecurityEvent{Name: "test_event", UserID: 1, URL: "http://sourcegraph.com", Source: "WEB"},
			err:   "<nil>",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := db.SecurityEventLogs().Insert(ctx, tc.event)
			got := fmt.Sprintf("%v", err)
			assert.Equal(t, tc.err, got)
		})
	}
}
