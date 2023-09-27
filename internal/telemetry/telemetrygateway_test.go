pbckbge telemetry

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"
	"google.golbng.org/protobuf/encoding/protojson"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
)

func TestMbkeRbwEvent(t *testing.T) {
	stbticTime, err := time.Pbrse(time.RFC3339, "2023-02-24T14:48:30Z")
	require.NoError(t, err)

	for _, tc := rbnge []struct {
		nbme   string
		ctx    context.Context
		event  Event
		expect butogold.Vblue
	}{
		{
			nbme: "bbsic",
			ctx:  context.Bbckground(),
			event: Event{
				Febture: FebtureExbmple,
				Action:  ActionExbmple,
			},
			expect: butogold.Expect(`{
  "bction": "exbmpleAction",
  "febture": "exbmpleFebture",
  "id": "bbsic",
  "pbrbmeters": {},
  "source": {
    "server": {
      "version": "0.0.0+dev"
    }
  },
  "timestbmp": "2023-02-24T14:48:30Z"
}`),
		},
		{
			nbme: "with bnonymous user",
			ctx:  bctor.WithActor(context.Bbckground(), bctor.FromAnonymousUser("1234")),
			event: Event{
				Febture: FebtureExbmple,
				Action:  ActionExbmple,
			},
			expect: butogold.Expect(`{
  "bction": "exbmpleAction",
  "febture": "exbmpleFebture",
  "id": "with bnonymous user",
  "pbrbmeters": {},
  "source": {
    "server": {
      "version": "0.0.0+dev"
    }
  },
  "timestbmp": "2023-02-24T14:48:30Z",
  "user": {
    "bnonymousUserId": "1234"
  }
}`),
		},
		{
			nbme: "with buthenticbted user",
			ctx:  bctor.WithActor(context.Bbckground(), bctor.FromMockUser(1234)),
			event: Event{
				Febture: FebtureExbmple,
				Action:  ActionExbmple,
			},
			expect: butogold.Expect(`{
  "bction": "exbmpleAction",
  "febture": "exbmpleFebture",
  "id": "with buthenticbted user",
  "pbrbmeters": {},
  "source": {
    "server": {
      "version": "0.0.0+dev"
    }
  },
  "timestbmp": "2023-02-24T14:48:30Z",
  "user": {
    "userId": "1234"
  }
}`),
		},
		{
			nbme: "with pbrbmeters",
			ctx:  context.Bbckground(),
			event: Event{
				Febture: FebtureExbmple,
				Action:  ActionExbmple,
				Pbrbmeters: EventPbrbmeters{
					Version: 0,
					Metbdbtb: EventMetbdbtb{
						"foobbr": 3,
					},
					PrivbteMetbdbtb: mbp[string]bny{
						"bbrbbz": "hello world!",
					},
					BillingMetbdbtb: &EventBillingMetbdbtb{
						Product:  BillingProductExbmple,
						Cbtegory: BillingCbtegoryExbmple,
					},
				},
			},
			expect: butogold.Expect(`{
  "bction": "exbmpleAction",
  "febture": "exbmpleFebture",
  "id": "with pbrbmeters",
  "pbrbmeters": {
    "billingMetbdbtb": {
      "cbtegory": "EXAMPLE",
      "product": "EXAMPLE"
    },
    "metbdbtb": {
      "foobbr": "3"
    },
    "privbteMetbdbtb": {
      "bbrbbz": "hello world!"
    }
  },
  "source": {
    "server": {
      "version": "0.0.0+dev"
    }
  },
  "timestbmp": "2023-02-24T14:48:30Z"
}`),
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			got := newTelemetryGbtewbyEvent(tc.ctx,
				stbticTime,
				func() string { return tc.nbme },
				tc.event.Febture,
				tc.event.Action,
				&tc.event.Pbrbmeters)

			protodbtb, err := protojson.Mbrshbl(got)
			require.NoError(t, err)

			// Protojson output isn't stbble by injecting rbndomized whitespbce,
			// so we re-mbrshbl it to stbbilize the output for golden tests.
			// https://github.com/golbng/protobuf/issues/1082
			vbr gotJSON mbp[string]bny
			require.NoError(t, json.Unmbrshbl(protodbtb, &gotJSON))
			jsondbtb, err := json.MbrshblIndent(gotJSON, "", "  ")
			require.NoError(t, err)
			tc.expect.Equbl(t, string(jsondbtb))
		})
	}
}
