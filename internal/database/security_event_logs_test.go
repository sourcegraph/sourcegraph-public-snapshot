package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSecurityEventLogs_ValidInfo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	// test setup and teardown
	prevConf := conf.Get()
	t.Cleanup(func() {
		conf.Mock(prevConf)
	})
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		Log: &schema.Log{
			SecurityEventLog: &schema.SecurityEventLog{Location: "all"},
		},
	}})

	logger, exportLogs := logtest.Captured(t)
	db := NewDB(logger, dbtest.NewDB(t))

	var testCases = []struct {
		name  string
		actor *actor.Actor // optional
		event *SecurityEvent
		err   string
	}{
		{
			name:  "EmptyName",
			event: &SecurityEvent{UserID: 1, URL: "http://sourcegraph.com", Source: "WEB"},
			err:   `INSERT: ERROR: new row for relation "security_event_logs" violates check constraint "security_event_logs_check_name_not_empty" (SQLSTATE 23514)`,
		},
		{
			name: "InvalidUser",
			event: &SecurityEvent{Name: "test_event", URL: "http://sourcegraph.com", Source: "WEB",
				// a UserID or AnonymousUserID is required to identify a user, unless internal
				UserID: 0, AnonymousUserID: ""},
			err: `INSERT: ERROR: new row for relation "security_event_logs" violates check constraint "security_event_logs_check_has_user" (SQLSTATE 23514)`,
		},
		{
			name:  "InternalActor",
			actor: &actor.Actor{Internal: true},
			event: &SecurityEvent{Name: "test_event", URL: "http://sourcegraph.com", Source: "WEB",
				// unset UserID and AnonymousUserID will error in other scenarios
				UserID: 0, AnonymousUserID: ""},
			err: "<nil>",
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
			actor: &actor.Actor{UID: 1}, // if we have a userID, we should have a valid actor UID
			event: &SecurityEvent{Name: "test_event", URL: "http://sourcegraph.com", Source: "Web", UserID: 1, AnonymousUserID: ""},
			err:   "<nil>",
		},
		{
			name:  "JustAnonymous",
			actor: &actor.Actor{AnonymousUID: "blah"},
			event: &SecurityEvent{Name: "test_event", URL: "http://sourcegraph.com", Source: "Web", UserID: 0, AnonymousUserID: "blah"},
			err:   "<nil>",
		},
		{
			name:  "ValidInsert",
			actor: &actor.Actor{UID: 1}, // if we have a userID, we should have a valid actor UID
			event: &SecurityEvent{Name: "test_event", UserID: 1, URL: "http://sourcegraph.com", Source: "WEB"},
			err:   "<nil>",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.actor != nil {
				ctx = actor.WithActor(ctx, tc.actor)
			}
			err := db.SecurityEventLogs().Insert(ctx, tc.event)
			got := fmt.Sprintf("%v", err)
			assert.Equal(t, tc.err, got)
		})
	}

	logs := exportLogs()
	auditLogs := filterAudit(logs)
	assert.Equal(t, 3, len(auditLogs)) // note: internal actor does not generate an audit log
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
	assert.NotNil(t, field["UserID"])
	assert.NotNil(t, field["AnonymousUserID"])
	assert.NotEmpty(t, field["source"])
	assert.NotEmpty(t, field["argument"])
	assert.NotEmpty(t, field["version"])
	assert.NotEmpty(t, field["timestamp"])
}

func TestLogSecurityEvent1(t *testing.T) {
	ctx := context.Background()
	logger, exportLogs := logtest.Captured(t)

	db := NewDB(logger, dbtest.NewDB(t))

	t.Run("valid event", func(t *testing.T) {
		err := db.SecurityEventLogs().LogSecurityEvent(ctx, SecurityEventAccessTokenCreated, "http://sourcegraph.com", 123, "AnonymousUserID", "source", nil)
		require.NoError(t, err)
	})

	t.Run("invalid arguments", func(t *testing.T) {
		err := db.SecurityEventLogs().LogSecurityEvent(ctx, SecurityEventAccessTokenCreated, "http://sourcegraph.com", 123, "AnonymousUserID", "source", make(chan int))
		require.Error(t, err)
	})

	t.Run("sourcegraph operator", func(t *testing.T) {
		ctx = actor.WithActor(context.Background(), &actor.Actor{UID: 123, SourcegraphOperator: true})
		err := db.SecurityEventLogs().LogSecurityEvent(ctx, SecurityEventAccessTokenCreated, "http://sourcegraph.com", 123, "AnonymousUserID", "source", nil)
		require.NoError(t, err)

		logs := exportLogs()
		for _, log := range logs {
			require.NotEqual(t, log.Level, sglog.LevelError, "Should not return error: %v", log.Fields["error"])
		}
	})
}
