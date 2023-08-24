package audittest

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/audit"
)

func TestExtractAuditFields(t *testing.T) {
	l, exportLogs := logtest.Captured(t)
	record := audit.Record{
		Entity: "foobar",
		Action: "barbaz",
	}
	ctx := actor.WithActor(context.Background(), actor.FromAnonymousUser("asdf"))

	audit.Log(ctx, l, record)

	entries := exportLogs()
	assert.True(t, entries.Contains(func(l logtest.CapturedLog) bool {
		fields, ok := ExtractAuditFields(l)
		if !ok {
			return ok
		}
		assert.Equal(t, record.Action, fields.Action)
		assert.Equal(t, record.Entity, fields.Entity)
		return !t.Failed()
	}))
}
