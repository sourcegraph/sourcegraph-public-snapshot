package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSecurityEventLogs_ValidInfo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger, exportLogs := logtest.Captured(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
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

	logs := exportLogs()
	auditLogs := filterAudit(logs)
	assert.Equal(t, 3, len(auditLogs))
	for _, auditLog := range auditLogs {
		assertAuditField(t, auditLog.Fields["audit"].(map[string]any))
		assertEventField(t, auditLog.Fields["event"].(map[string]any))
	}
}

func filterAudit(logs []logtest.CapturedLog) []logtest.CapturedLog {
	var filtered []logtest.CapturedLog
	for _, log := range logs {
		if log.Fields["audit"] != nil {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

func assertAuditField(t *testing.T, field map[string]any) {
	t.Helper()
	assert.NotEmpty(t, field["auditId"])
	assert.NotEmpty(t, field["entity"])

	actorField := field["actor"].(map[string]any)
	assert.NotEmpty(t, actorField["actorUID"])
	assert.NotEmpty(t, actorField["ip"])
	assert.NotEmpty(t, actorField["X-Forwarded-For"])
}

func assertEventField(t *testing.T, field map[string]any) {
	t.Helper()
	assert.NotEmpty(t, field["URL"])
	assert.NotEmpty(t, field["source"])
	assert.NotEmpty(t, field["argument"])
	assert.NotEmpty(t, field["version"])
	assert.NotEmpty(t, field["timestamp"])
}
