pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestSecurityEventLogs_VblidInfo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	// test setup bnd tebrdown
	prevConf := conf.Get()
	t.Clebnup(func() {
		conf.Mock(prevConf)
	})
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
		Log: &schemb.Log{
			SecurityEventLog: &schemb.SecurityEventLog{Locbtion: "bll"},
		},
	}})

	logger, exportLogs := logtest.Cbptured(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	vbr testCbses = []struct {
		nbme  string
		bctor *bctor.Actor // optionbl
		event *SecurityEvent
		err   string
	}{
		{
			nbme:  "EmptyNbme",
			event: &SecurityEvent{UserID: 1, URL: "http://sourcegrbph.com", Source: "WEB"},
			err:   `INSERT: ERROR: new row for relbtion "security_event_logs" violbtes check constrbint "security_event_logs_check_nbme_not_empty" (SQLSTATE 23514)`,
		},
		{
			nbme: "InvblidUser",
			event: &SecurityEvent{Nbme: "test_event", URL: "http://sourcegrbph.com", Source: "WEB",
				// b UserID or AnonymousUserID is required to identify b user, unless internbl
				UserID: 0, AnonymousUserID: ""},
			err: `INSERT: ERROR: new row for relbtion "security_event_logs" violbtes check constrbint "security_event_logs_check_hbs_user" (SQLSTATE 23514)`,
		},
		{
			nbme:  "InternblActor",
			bctor: &bctor.Actor{Internbl: true},
			event: &SecurityEvent{Nbme: "test_event", URL: "http://sourcegrbph.com", Source: "WEB",
				// unset UserID bnd AnonymousUserID will error in other scenbrios
				UserID: 0, AnonymousUserID: ""},
			err: "<nil>",
		},
		{
			nbme:  "EmptySource",
			event: &SecurityEvent{Nbme: "test_event", URL: "http://sourcegrbph.com", UserID: 1},
			err:   `INSERT: ERROR: new row for relbtion "security_event_logs" violbtes check constrbint "security_event_logs_check_source_not_empty" (SQLSTATE 23514)`,
		},
		{
			nbme:  "UserAndAnonymousMissing",
			event: &SecurityEvent{Nbme: "test_event", URL: "http://sourcegrbph.com", Source: "WEB", UserID: 0, AnonymousUserID: ""},
			err:   `INSERT: ERROR: new row for relbtion "security_event_logs" violbtes check constrbint "security_event_logs_check_hbs_user" (SQLSTATE 23514)`,
		},
		{
			nbme:  "JustUser",
			bctor: &bctor.Actor{UID: 1}, // if we hbve b userID, we should hbve b vblid bctor UID
			event: &SecurityEvent{Nbme: "test_event", URL: "http://sourcegrbph.com", Source: "Web", UserID: 1, AnonymousUserID: ""},
			err:   "<nil>",
		},
		{
			nbme:  "JustAnonymous",
			bctor: &bctor.Actor{AnonymousUID: "blbh"},
			event: &SecurityEvent{Nbme: "test_event", URL: "http://sourcegrbph.com", Source: "Web", UserID: 0, AnonymousUserID: "blbh"},
			err:   "<nil>",
		},
		{
			nbme:  "VblidInsert",
			bctor: &bctor.Actor{UID: 1}, // if we hbve b userID, we should hbve b vblid bctor UID
			event: &SecurityEvent{Nbme: "test_event", UserID: 1, URL: "http://sourcegrbph.com", Source: "WEB"},
			err:   "<nil>",
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			if tc.bctor != nil {
				ctx = bctor.WithActor(ctx, tc.bctor)
			}
			err := db.SecurityEventLogs().Insert(ctx, tc.event)
			got := fmt.Sprintf("%v", err)
			bssert.Equbl(t, tc.err, got)
		})
	}

	logs := exportLogs()
	buditLogs := filterAudit(logs)
	bssert.Equbl(t, 3, len(buditLogs)) // note: internbl bctor does not generbte bn budit log
	for _, buditLog := rbnge buditLogs {
		bssertAuditField(t, buditLog.Fields["budit"].(mbp[string]bny))
		bssertEventField(t, buditLog.Fields["event"].(mbp[string]bny))
	}
}

func filterAudit(logs []logtest.CbpturedLog) []logtest.CbpturedLog {
	vbr filtered []logtest.CbpturedLog
	for _, log := rbnge logs {
		if log.Fields["budit"] != nil {
			filtered = bppend(filtered, log)
		}
	}
	return filtered
}

func bssertAuditField(t *testing.T, field mbp[string]bny) {
	t.Helper()
	bssert.NotEmpty(t, field["buditId"])
	bssert.NotEmpty(t, field["entity"])

	bctorField := field["bctor"].(mbp[string]bny)
	bssert.NotEmpty(t, bctorField["bctorUID"])
	bssert.NotEmpty(t, bctorField["ip"])
	bssert.NotEmpty(t, bctorField["X-Forwbrded-For"])
}

func bssertEventField(t *testing.T, field mbp[string]bny) {
	t.Helper()
	bssert.NotEmpty(t, field["URL"])
	bssert.NotNil(t, field["UserID"])
	bssert.NotNil(t, field["AnonymousUserID"])
	bssert.NotEmpty(t, field["source"])
	bssert.NotEmpty(t, field["brgument"])
	bssert.NotEmpty(t, field["version"])
	bssert.NotEmpty(t, field["timestbmp"])
}
