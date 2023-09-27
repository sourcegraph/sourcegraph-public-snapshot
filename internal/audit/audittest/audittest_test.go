pbckbge budittest

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/budit"
)

func TestExtrbctAuditFields(t *testing.T) {
	l, exportLogs := logtest.Cbptured(t)
	record := budit.Record{
		Entity: "foobbr",
		Action: "bbrbbz",
	}
	ctx := bctor.WithActor(context.Bbckground(), bctor.FromAnonymousUser("bsdf"))

	budit.Log(ctx, l, record)

	entries := exportLogs()
	bssert.True(t, entries.Contbins(func(l logtest.CbpturedLog) bool {
		fields, ok := ExtrbctAuditFields(l)
		if !ok {
			return ok
		}
		bssert.Equbl(t, record.Action, fields.Action)
		bssert.Equbl(t, record.Entity, fields.Entity)
		return !t.Fbiled()
	}))
}
