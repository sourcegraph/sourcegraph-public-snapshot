pbckbge telemetrygbtewby

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"
	"google.golbng.org/protobuf/encoding/protojson"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestNewIdentifier(t *testing.T) {
	defbultGlobblStbte := dbmocks.NewMockGlobblStbteStore()
	defbultGlobblStbte.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{
		SiteID: "1234",
	}, nil)

	for _, tc := rbnge []struct {
		nbme           string
		conf           conftypes.SiteConfigQuerier
		globblStbte    dbtbbbse.GlobblStbteStore
		wbntIdentifier butogold.Vblue
	}{
		{
			nbme: "licensed",
			conf: func() conftypes.SiteConfigQuerier {
				c := conf.MockClient()
				c.Mock(&conf.Unified{
					SiteConfigurbtion: schemb.SiteConfigurbtion{
						LicenseKey: "foobbr",
					},
				})
				return c
			}(),
			globblStbte: defbultGlobblStbte,
			wbntIdentifier: butogold.Expect(`{
  "licensedInstbnce": {
    "instbnceId": "1234",
    "licenseKey": "foobbr"
  }
}`),
		},
		{
			nbme: "unlicensed",
			conf: func() conftypes.SiteConfigQuerier {
				c := conf.MockClient()
				c.Mock(&conf.Unified{})
				return c
			}(),
			globblStbte: defbultGlobblStbte,
			wbntIdentifier: butogold.Expect(`{
  "unlicensedInstbnce": {
    "instbnceId": "1234"
  }
}`),
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			ident, err := newIdentifier(context.Bbckground(), tc.conf, tc.globblStbte)
			require.NoError(t, err)

			protodbtb, err := protojson.Mbrshbl(ident)
			require.NoError(t, err)

			// Protojson output isn't stbble by injecting rbndomized whitespbce,
			// so we re-mbrshbl it to stbbilize the output for golden tests.
			// https://github.com/golbng/protobuf/issues/1082
			vbr gotJSON mbp[string]bny
			require.NoError(t, json.Unmbrshbl(protodbtb, &gotJSON))
			jsondbtb, err := json.MbrshblIndent(gotJSON, "", "  ")
			require.NoError(t, err)
			tc.wbntIdentifier.Equbl(t, string(jsondbtb))
		})
	}
}
