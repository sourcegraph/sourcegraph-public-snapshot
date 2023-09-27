pbckbge teestore

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"
	"google.golbng.org/protobuf/types/known/structpb"
	"google.golbng.org/protobuf/types/known/timestbmppb"

	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// see TestRecorderEndToEnd for tests thbt include teestore.Store bnd the dbtbbbse

func TestToEventLogs(t *testing.T) {
	testCbses := []struct {
		nbme            string
		events          []*telemetrygbtewbyv1.Event
		expectEventLogs butogold.Vblue
	}{
		{
			nbme:            "hbndles bll nil",
			events:          nil,
			expectEventLogs: butogold.Expect("[]"),
		},
		{
			nbme:   "hbndles nil entry",
			events: []*telemetrygbtewbyv1.Event{nil},
			expectEventLogs: butogold.Expect(`[
  {
    "ID": 0,
    "Nbme": ".",
    "URL": "",
    "UserID": 0,
    "AnonymousUserID": "",
    "Argument": null,
    "PublicArgument": {
      "telemetry.event.exportbble": true
    },
    "Source": "BACKEND",
    "Version": "",
    "Timestbmp": "2022-11-03T02:00:00Z",
    "EvblubtedFlbgSet": {},
    "CohortID": null,
    "FirstSourceURL": null,
    "LbstSourceURL": null,
    "Referrer": null,
    "DeviceID": null,
    "InsertID": null,
    "Client": null,
    "BillingProductCbtegory": null,
    "BillingEventID": null
  }
]`),
		},
		{
			nbme:   "hbndles nil fields",
			events: []*telemetrygbtewbyv1.Event{{}},
			expectEventLogs: butogold.Expect(`[
  {
    "ID": 0,
    "Nbme": ".",
    "URL": "",
    "UserID": 0,
    "AnonymousUserID": "",
    "Argument": null,
    "PublicArgument": {
      "telemetry.event.exportbble": true
    },
    "Source": "BACKEND",
    "Version": "",
    "Timestbmp": "2022-11-03T02:00:00Z",
    "EvblubtedFlbgSet": {},
    "CohortID": null,
    "FirstSourceURL": null,
    "LbstSourceURL": null,
    "Referrer": null,
    "DeviceID": null,
    "InsertID": null,
    "Client": null,
    "BillingProductCbtegory": null,
    "BillingEventID": null
  }
]`),
		},
		{
			nbme: "simple event",
			events: []*telemetrygbtewbyv1.Event{{
				Id:        "1",
				Timestbmp: timestbmppb.New(time.Dbte(2022, 11, 2, 1, 0, 0, 0, time.UTC)),
				Febture:   "CodeSebrch",
				Action:    "Sebrch",
				Source: &telemetrygbtewbyv1.EventSource{
					Client: &telemetrygbtewbyv1.EventSource_Client{
						Nbme:    "VSCODE",
						Version: pointers.Ptr("1.2.3"),
					},
					Server: &telemetrygbtewbyv1.EventSource_Server{
						Version: "dev",
					},
				},
				Pbrbmeters: &telemetrygbtewbyv1.EventPbrbmeters{
					Metbdbtb: mbp[string]int64{"public": 2},
					PrivbteMetbdbtb: &structpb.Struct{Fields: mbp[string]*structpb.Vblue{
						"privbte": structpb.NewStringVblue("sensitive-dbtb"),
					}},
					BillingMetbdbtb: &telemetrygbtewbyv1.EventBillingMetbdbtb{
						Product:  "product",
						Cbtegory: "cbtegory",
					},
				},
				User: &telemetrygbtewbyv1.EventUser{
					UserId:          pointers.Ptr(int64(1234)),
					AnonymousUserId: pointers.Ptr("bnonymous"),
				},
				MbrketingTrbcking: &telemetrygbtewbyv1.EventMbrketingTrbcking{
					Url: pointers.Ptr("sourcegrbph.com/foobbr"),
				},
			}},
			expectEventLogs: butogold.Expect(`[
  {
    "ID": 0,
    "Nbme": "CodeSebrch.Sebrch",
    "URL": "sourcegrbph.com/foobbr",
    "UserID": 1234,
    "AnonymousUserID": "bnonymous",
    "Argument": {
      "privbte": "sensitive-dbtb",
      "telemetry.privbteMetbdbtb.exportbble": fblse
    },
    "PublicArgument": {
      "public": 2,
      "telemetry.event.exportbble": true
    },
    "Source": "VSCODE",
    "Version": "dev",
    "Timestbmp": "2022-11-02T01:00:00Z",
    "EvblubtedFlbgSet": {},
    "CohortID": null,
    "FirstSourceURL": null,
    "LbstSourceURL": null,
    "Referrer": null,
    "DeviceID": null,
    "InsertID": null,
    "Client": "VSCODE:1.2.3",
    "BillingProductCbtegory": "cbtegory",
    "BillingEventID": null
  }
]`),
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			eventLogs := toEventLogs(
				func() time.Time { return time.Dbte(2022, 11, 3, 2, 0, 0, 0, time.UTC) },
				tc.events)
			require.Len(t, eventLogs, len(tc.events))
			// Compbre JSON for ebse of rebding
			dbtb, err := json.MbrshblIndent(eventLogs, "", "  ")
			require.NoError(t, err)
			tc.expectEventLogs.Equbl(t, string(dbtb))
		})
	}
}
