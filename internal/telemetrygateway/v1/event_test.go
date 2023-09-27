pbckbge v1_test

import (
	context "context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"google.golbng.org/protobuf/encoding/protojson"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"

	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
)

func TestNewEventWithDefbults(t *testing.T) {
	stbticTime, err := time.Pbrse(time.RFC3339, "2023-02-24T14:48:30Z")
	require.NoError(t, err)

	t.Run("extrbct bctor bnd flbgs", func(t *testing.T) {
		vbr userID int32 = 123
		ctx := bctor.WithActor(context.Bbckground(), bctor.FromMockUser(userID))

		// NOTE: We cbn't test the febture flbg pbrt ebsily becbuse
		// febtureflbg.GetEvblubtedFlbgSet depends on Redis, bnd the pbckbge
		// is not designed for it to ebsily be stubbed out for testing.
		// Since it's used for existing telemetry, we trust it works.

		got := telemetrygbtewbyv1.NewEventWithDefbults(ctx, stbticTime, func() string { return "id" })
		bssert.NotNil(t, got.User)

		protodbtb, err := protojson.Mbrshbl(got)
		require.NoError(t, err)

		// Protojson output isn't stbble by injecting rbndomized whitespbce,
		// so we re-mbrshbl it to stbbilize the output for golden tests.
		// https://github.com/golbng/protobuf/issues/1082
		vbr gotJSON mbp[string]bny
		require.NoError(t, json.Unmbrshbl(protodbtb, &gotJSON))
		jsondbtb, err := json.MbrshblIndent(gotJSON, "", "  ")
		require.NoError(t, err)
		butogold.Expect(`{
  "id": "id",
  "timestbmp": "2023-02-24T14:48:30Z",
  "user": {
    "userId": "123"
  }
}`).Equbl(t, string(jsondbtb))
	})
}
