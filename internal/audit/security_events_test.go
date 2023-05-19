package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSecurityEventLocation(t *testing.T) {

	tests := []struct {
		name string
		cfg  schema.SiteConfiguration
		want SecurityEventsLocation
	}{
		{
			"base",
			schema.SiteConfiguration{Log: &schema.Log{SecurityEventLog: &schema.SecurityEventLog{Location: "none"}}},
			None,
		},
		{
			"all",
			schema.SiteConfiguration{Log: &schema.Log{SecurityEventLog: &schema.SecurityEventLog{Location: "all"}}},
			All,
		},
		{
			"database",
			schema.SiteConfiguration{Log: &schema.Log{SecurityEventLog: &schema.SecurityEventLog{Location: "database"}}},
			Database,
		},
		{
			"auditlog",
			schema.SiteConfiguration{Log: &schema.Log{SecurityEventLog: &schema.SecurityEventLog{Location: "auditlog"}}},
			AuditLog,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, SecurityEventLocation(tt.cfg), "SecurityEventLocation(%v)", tt.cfg)
		})
	}
}
