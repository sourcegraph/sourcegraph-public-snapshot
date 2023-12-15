package telemetrygateway

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNewIdentifier(t *testing.T) {
	defaultGlobalState := dbmocks.NewMockGlobalStateStore()
	defaultGlobalState.GetFunc.SetDefaultReturn(database.GlobalState{
		SiteID: "1234",
	}, nil)

	for _, tc := range []struct {
		name           string
		conf           conftypes.SiteConfigQuerier
		globalState    database.GlobalStateStore
		wantIdentifier autogold.Value
	}{
		{
			name: "licensed",
			conf: func() conftypes.SiteConfigQuerier {
				c := conf.MockClient()
				c.Mock(&conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						LicenseKey:  "foobar",
						ExternalURL: "sourcegraph.com",
					},
				})
				return c
			}(),
			globalState: defaultGlobalState,
			wantIdentifier: autogold.Expect(`{
  "licensedInstance": {
    "externalUrl": "sourcegraph.com",
    "instanceId": "1234",
    "licenseKey": "foobar"
  }
}`),
		},
		{
			name: "unlicensed",
			conf: func() conftypes.SiteConfigQuerier {
				c := conf.MockClient()
				c.Mock(&conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						ExternalURL: "sourcegraph.com",
					},
				})
				return c
			}(),
			globalState: defaultGlobalState,
			wantIdentifier: autogold.Expect(`{
  "unlicensedInstance": {
    "externalUrl": "sourcegraph.com",
    "instanceId": "1234"
  }
}`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ident, err := newIdentifier(context.Background(), tc.conf, tc.globalState)
			require.NoError(t, err)

			protodata, err := protojson.Marshal(ident)
			require.NoError(t, err)

			// Protojson output isn't stable by injecting randomized whitespace,
			// so we re-marshal it to stabilize the output for golden tests.
			// https://github.com/golang/protobuf/issues/1082
			var gotJSON map[string]any
			require.NoError(t, json.Unmarshal(protodata, &gotJSON))
			jsondata, err := json.MarshalIndent(gotJSON, "", "  ")
			require.NoError(t, err)
			tc.wantIdentifier.Equal(t, string(jsondata))
		})
	}
}
