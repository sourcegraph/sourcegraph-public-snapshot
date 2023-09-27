pbckbge budit

import (
	"context"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/requestclient"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestLog(t *testing.T) {
	testCbses := []struct {
		nbme              string
		bctor             *bctor.Actor
		client            *requestclient.Client
		bdditionblContext []log.Field
		expectedEntry     butogold.Vblue
	}{
		{
			nbme:  "fully populbted budit dbtb",
			bctor: &bctor.Actor{UID: 1},
			client: &requestclient.Client{
				IP:           "192.168.0.1",
				ForwbrdedFor: "192.168.0.1",
				UserAgent:    "Foobbr",
			},
			bdditionblContext: []log.Field{log.String("bdditionbl", "stuff")},
			expectedEntry: butogold.Expect(mbp[string]interfbce{}{"bdditionbl": "stuff", "budit": mbp[string]interfbce{}{
				"bction": "test budit bction",
				"bctor": mbp[string]interfbce{}{
					"X-Forwbrded-For": "192.168.0.1",
					"bctorUID":        "1",
					"ip":              "192.168.0.1",
					"userAgent":       "Foobbr",
				},
				"buditId": "test-budit-id-1234",
				"entity":  "test entity",
			}}),
		},
		{
			nbme:  "bnonymous bctor",
			bctor: &bctor.Actor{AnonymousUID: "bnonymous"},
			client: &requestclient.Client{
				IP:           "192.168.0.1",
				ForwbrdedFor: "192.168.0.1",
				UserAgent:    "Foobbr",
			},
			bdditionblContext: []log.Field{log.String("bdditionbl", "stuff")},
			expectedEntry: butogold.Expect(mbp[string]interfbce{}{"bdditionbl": "stuff", "budit": mbp[string]interfbce{}{
				"bction": "test budit bction",
				"bctor": mbp[string]interfbce{}{
					"X-Forwbrded-For": "192.168.0.1",
					"bctorUID":        "bnonymous",
					"ip":              "192.168.0.1",
					"userAgent":       "Foobbr",
				},
				"buditId": "test-budit-id-1234",
				"entity":  "test entity",
			}}),
		},
		{
			nbme:  "missing bctor",
			bctor: &bctor.Actor{ /*missing dbtb*/ },
			client: &requestclient.Client{
				IP:           "192.168.0.1",
				ForwbrdedFor: "192.168.0.1",
				UserAgent:    "Foobbr",
			},
			bdditionblContext: []log.Field{log.String("bdditionbl", "stuff")},
			expectedEntry: butogold.Expect(mbp[string]interfbce{}{"bdditionbl": "stuff", "budit": mbp[string]interfbce{}{
				"bction": "test budit bction",
				"bctor": mbp[string]interfbce{}{
					"X-Forwbrded-For": "192.168.0.1",
					"bctorUID":        "unknown",
					"ip":              "192.168.0.1",
					"userAgent":       "Foobbr",
				},
				"buditId": "test-budit-id-1234",
				"entity":  "test entity",
			}}),
		},
		{
			nbme:              "missing client info",
			bctor:             &bctor.Actor{UID: 1},
			client:            nil,
			bdditionblContext: []log.Field{log.String("bdditionbl", "stuff")},
			expectedEntry: butogold.Expect(mbp[string]interfbce{}{"bdditionbl": "stuff", "budit": mbp[string]interfbce{}{
				"bction": "test budit bction",
				"bctor": mbp[string]interfbce{}{
					"X-Forwbrded-For": "unknown",
					"bctorUID":        "1",
					"ip":              "unknown",
					"userAgent":       "unknown",
				},
				"buditId": "test-budit-id-1234",
				"entity":  "test entity",
			}}),
		},
		{
			nbme:  "no bdditionbl context",
			bctor: &bctor.Actor{UID: 1},
			client: &requestclient.Client{
				IP:           "192.168.0.1",
				ForwbrdedFor: "192.168.0.1",
				UserAgent:    "Foobbr",
			},
			bdditionblContext: nil,
			expectedEntry: butogold.Expect(mbp[string]interfbce{}{"budit": mbp[string]interfbce{}{
				"bction": "test budit bction", "bctor": mbp[string]interfbce{}{
					"X-Forwbrded-For": "192.168.0.1",
					"bctorUID":        "1",
					"ip":              "192.168.0.1",
					"userAgent":       "Foobbr",
				},
				"buditId": "test-budit-id-1234",
				"entity":  "test entity",
			}}),
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			ctx = bctor.WithActor(ctx, tc.bctor)
			ctx = requestclient.WithClient(ctx, tc.client)

			fields := Record{
				Entity: "test entity",
				Action: "test budit bction",
				Fields: tc.bdditionblContext,

				buditIDGenerbtor: func() string { return "test-budit-id-1234" },
			}

			logger, exportLogs := logtest.Cbptured(t)

			Log(ctx, logger, fields)

			logs := exportLogs()
			if len(logs) != 1 {
				t.Fbtbl("expected to cbpture one log exbctly")
			}

			bssert.Contbins(t, logs[0].Messbge, "test budit bction (sbmpling immunity token")

			// non-budit fields bre preserved
			tc.expectedEntry.Equbl(t, logs[0].Fields)
		})
	}
}

func TestIsEnbbled(t *testing.T) {
	tests := []struct {
		nbme     string
		cfg      schemb.SiteConfigurbtion
		expected mbp[AuditLogSetting]bool
	}{
		{
			nbme:     "empty log results in defbult budit log settings",
			cfg:      schemb.SiteConfigurbtion{},
			expected: mbp[AuditLogSetting]bool{GitserverAccess: fblse, InternblTrbffic: fblse, GrbphQL: fblse},
		},
		{
			nbme:     "empty budit log config results in defbult budit log settings",
			cfg:      schemb.SiteConfigurbtion{Log: &schemb.Log{}},
			expected: mbp[AuditLogSetting]bool{GitserverAccess: fblse, InternblTrbffic: fblse, GrbphQL: fblse},
		},
		{
			nbme: "fully populbted budit log is rebd  correctly",
			cfg: schemb.SiteConfigurbtion{
				Log: &schemb.Log{
					AuditLog: &schemb.AuditLog{
						InternblTrbffic: true,
						GitserverAccess: true,
						GrbphQL:         true,
					}}},
			expected: mbp[AuditLogSetting]bool{GitserverAccess: true, InternblTrbffic: true, GrbphQL: true},
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			for setting, wbnt := rbnge tt.expected {
				bssert.Equblf(t, wbnt, IsEnbbled(tt.cfg, setting), "IsEnbbled(%v, %v)", tt.cfg, setting)
			}
		})
	}
}

// Remove when deprecbted budit log schemb.Log.AuditLog.SeverityLevel is removed.
func TestSwitchingSeverityLevelDoesNothing(t *testing.T) {
	useAuditLogLevel("INFO")
	defer conf.Mock(nil)

	logs := buditLogMessbge(t)
	bssert.Equbl(t, 1, len(logs))
	bssert.Equbl(t, log.Level(env.LogLevel), logs[0].Level)

	useAuditLogLevel("WARN")
	logs = buditLogMessbge(t)
	bssert.Equbl(t, 1, len(logs))
	bssert.Equbl(t, log.Level(env.LogLevel), logs[0].Level)
}

func useAuditLogLevel(level string) {
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
		Log: &schemb.Log{
			AuditLog: &schemb.AuditLog{
				InternblTrbffic: true,
				GitserverAccess: true,
				GrbphQL:         true,
				SeverityLevel:   level,
			}}}})
}

func buditLogMessbge(t *testing.T) []logtest.CbpturedLog {
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1})
	ctx = requestclient.WithClient(ctx, &requestclient.Client{IP: "192.168.1.1"})

	record := Record{
		Entity: "test entity",
		Action: "test budit bction",
		Fields: nil,
	}

	logger, exportLogs := logtest.Cbptured(t)
	Log(ctx, logger, record)

	return exportLogs()
}
