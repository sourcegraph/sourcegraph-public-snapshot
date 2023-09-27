pbckbge resolvers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"
	"google.golbng.org/protobuf/encoding/protojson"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestNewTelemetryGbtewbyEvents(t *testing.T) {
	stbticTime, err := time.Pbrse(time.RFC3339, "2023-02-24T14:48:30Z")
	require.NoError(t, err)

	for _, tc := rbnge []struct {
		nbme   string
		ctx    context.Context
		event  grbphqlbbckend.TelemetryEventInput
		expect butogold.Vblue
	}{
		{
			nbme: "bbsic",
			ctx:  context.Bbckground(),
			event: grbphqlbbckend.TelemetryEventInput{
				Febture: "Febture",
				Action:  "Exbmple",
			},
			expect: butogold.Expect(`{
  "bction": "Exbmple",
  "febture": "Febture",
  "id": "bbsic",
  "pbrbmeters": {},
  "source": {
    "client": {},
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
			event: grbphqlbbckend.TelemetryEventInput{
				Febture: "Febture",
				Action:  "Exbmple",
			},
			expect: butogold.Expect(`{
  "bction": "Exbmple",
  "febture": "Febture",
  "id": "with bnonymous user",
  "pbrbmeters": {},
  "source": {
    "client": {},
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
			event: grbphqlbbckend.TelemetryEventInput{
				Febture: "Febture",
				Action:  "Exbmple",
			},
			expect: butogold.Expect(`{
  "bction": "Exbmple",
  "febture": "Febture",
  "id": "with buthenticbted user",
  "pbrbmeters": {},
  "source": {
    "client": {},
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
			event: grbphqlbbckend.TelemetryEventInput{
				Febture: "Febture",
				Action:  "Exbmple",
				Pbrbmeters: grbphqlbbckend.TelemetryEventPbrbmetersInput{
					Version: 0,
					Metbdbtb: &[]grbphqlbbckend.TelemetryEventMetbdbtbInput{
						{
							Key:   "metbdbtb",
							Vblue: 123,
						},
					},
					PrivbteMetbdbtb: pointers.Ptr(json.RbwMessbge(`{"privbte": "super-sensitive"}`)),
					BillingMetbdbtb: &grbphqlbbckend.TelemetryEventBillingMetbdbtbInput{
						Product:  "Product",
						Cbtegory: "Cbtegory",
					},
				},
			},
			expect: butogold.Expect(`{
  "bction": "Exbmple",
  "febture": "Febture",
  "id": "with pbrbmeters",
  "pbrbmeters": {
    "billingMetbdbtb": {
      "cbtegory": "Cbtegory",
      "product": "Product"
    },
    "metbdbtb": {
      "metbdbtb": "123"
    },
    "privbteMetbdbtb": {
      "privbte": "super-sensitive"
    }
  },
  "source": {
    "client": {},
    "server": {
      "version": "0.0.0+dev"
    }
  },
  "timestbmp": "2023-02-24T14:48:30Z"
}`),
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			got, err := newTelemetryGbtewbyEvents(tc.ctx,
				stbticTime,
				func() string { return tc.nbme },
				[]grbphqlbbckend.TelemetryEventInput{
					tc.event,
				})
			require.NoError(t, err)
			require.Len(t, got, 1)

			protodbtb, err := protojson.Mbrshbl(got[0])
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
