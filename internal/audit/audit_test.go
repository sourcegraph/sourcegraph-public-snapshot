package audit

import (
	"context"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestLog(t *testing.T) {
	testCases := []struct {
		name              string
		actor             *actor.Actor
		client            *requestclient.Client
		additionalContext []log.Field
		expectedEntry     map[string]interface{}
	}{
		{
			name:  "fully populated audit data",
			actor: &actor.Actor{UID: 1},
			client: &requestclient.Client{
				IP:           "192.168.0.1",
				ForwardedFor: "192.168.0.1",
			},
			additionalContext: []log.Field{log.String("additional", "stuff")},
			expectedEntry: map[string]interface{}{
				"audit": map[string]interface{}{
					"entity": "test entity",
					"actor": map[string]interface{}{
						"actorUID":        "1",
						"ip":              "192.168.0.1",
						"X-Forwarded-For": "192.168.0.1",
					},
				},
				"additional": "stuff",
			},
		},
		{
			name:  "anonymous actor",
			actor: &actor.Actor{AnonymousUID: "anonymous"},
			client: &requestclient.Client{
				IP:           "192.168.0.1",
				ForwardedFor: "192.168.0.1",
			},
			additionalContext: []log.Field{log.String("additional", "stuff")},
			expectedEntry: map[string]interface{}{
				"audit": map[string]interface{}{
					"entity": "test entity",
					"actor": map[string]interface{}{
						"actorUID":        "anonymous",
						"ip":              "192.168.0.1",
						"X-Forwarded-For": "192.168.0.1",
					},
				},
				"additional": "stuff",
			},
		},
		{
			name:  "missing actor",
			actor: &actor.Actor{ /*missing data*/ },
			client: &requestclient.Client{
				IP:           "192.168.0.1",
				ForwardedFor: "192.168.0.1",
			},
			additionalContext: []log.Field{log.String("additional", "stuff")},
			expectedEntry: map[string]interface{}{
				"audit": map[string]interface{}{
					"entity": "test entity",
					"actor": map[string]interface{}{
						"actorUID":        "unknown",
						"ip":              "192.168.0.1",
						"X-Forwarded-For": "192.168.0.1",
					},
				},
				"additional": "stuff",
			},
		},
		{
			name:              "missing client info",
			actor:             &actor.Actor{UID: 1},
			client:            nil,
			additionalContext: []log.Field{log.String("additional", "stuff")},
			expectedEntry: map[string]interface{}{
				"audit": map[string]interface{}{
					"entity": "test entity",
					"actor": map[string]interface{}{
						"actorUID":        "1",
						"ip":              "unknown",
						"X-Forwarded-For": "unknown",
					},
				},
				"additional": "stuff",
			},
		},
		{
			name:  "no additional context",
			actor: &actor.Actor{UID: 1},
			client: &requestclient.Client{
				IP:           "192.168.0.1",
				ForwardedFor: "192.168.0.1",
			},
			additionalContext: nil,
			expectedEntry: map[string]interface{}{
				"audit": map[string]interface{}{
					"entity": "test entity",
					"actor": map[string]interface{}{
						"actorUID":        "1",
						"ip":              "192.168.0.1",
						"X-Forwarded-For": "192.168.0.1",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = actor.WithActor(ctx, tc.actor)
			ctx = requestclient.WithClient(ctx, tc.client)

			fields := Record{
				Entity: "test entity",
				Action: "test audit action",
				Fields: tc.additionalContext,
			}

			logger, exportLogs := logtest.Captured(t)

			Log(ctx, logger, fields)

			logs := exportLogs()
			if len(logs) != 1 {
				t.Fatal("expected to capture one log exactly")
			}

			assert.Contains(t, logs[0].Message, "test audit action (sampling immunity token")

			// non-audit fields are preserved
			assert.Equal(t, tc.expectedEntry["additional"], logs[0].Fields["additional"])

			expectedAudit := tc.expectedEntry["audit"].(map[string]interface{})
			actualAudit := logs[0].Fields["audit"].(map[string]interface{})
			assert.Equal(t, expectedAudit["entity"], actualAudit["entity"])
			assert.NotEmpty(t, actualAudit["auditId"])

			expectedActor := expectedAudit["actor"].(map[string]interface{})
			actualActor := actualAudit["actor"].(map[string]interface{})
			assert.Equal(t, expectedActor["actorUID"], actualActor["actorUID"])
			assert.Equal(t, expectedActor["ip"], actualActor["ip"])
			assert.Equal(t, expectedActor["X-Forwarded-For"], actualActor["X-Forwarded-For"])
		})
	}
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		cfg      schema.SiteConfiguration
		expected map[AuditLogSetting]bool
	}{
		{
			name:     "empty log results in default audit log settings",
			cfg:      schema.SiteConfiguration{},
			expected: map[AuditLogSetting]bool{GitserverAccess: false, InternalTraffic: false, GraphQL: false},
		},
		{
			name:     "empty audit log config results in default audit log settings",
			cfg:      schema.SiteConfiguration{Log: &schema.Log{}},
			expected: map[AuditLogSetting]bool{GitserverAccess: false, InternalTraffic: false, GraphQL: false},
		},
		{
			name: "fully populated audit log is read  correctly",
			cfg: schema.SiteConfiguration{
				Log: &schema.Log{
					AuditLog: &schema.AuditLog{
						InternalTraffic: true,
						GitserverAccess: true,
						GraphQL:         true,
					}}},
			expected: map[AuditLogSetting]bool{GitserverAccess: true, InternalTraffic: true, GraphQL: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for setting, want := range tt.expected {
				assert.Equalf(t, want, IsEnabled(tt.cfg, setting), "IsEnabled(%v, %v)", tt.cfg, setting)
			}
		})
	}
}

func TestSwitchingSeverityLevel(t *testing.T) {
	useAuditLogLevel("INFO")
	defer conf.Mock(nil)

	logs := auditLogMessage(t)
	assert.Equal(t, 1, len(logs))
	assert.Equal(t, log.LevelInfo, logs[0].Level)

	useAuditLogLevel("WARN")
	logs = auditLogMessage(t)
	assert.Equal(t, 1, len(logs))
	assert.Equal(t, log.LevelWarn, logs[0].Level)
}

func useAuditLogLevel(level string) {
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		Log: &schema.Log{
			AuditLog: &schema.AuditLog{
				InternalTraffic: true,
				GitserverAccess: true,
				GraphQL:         true,
				SeverityLevel:   level,
			}}}})
}

func auditLogMessage(t *testing.T) []logtest.CapturedLog {
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
	ctx = requestclient.WithClient(ctx, &requestclient.Client{IP: "192.168.1.1"})

	record := Record{
		Entity: "test entity",
		Action: "test audit action",
		Fields: nil,
	}

	logger, exportLogs := logtest.Captured(t)
	Log(ctx, logger, record)

	return exportLogs()
}
