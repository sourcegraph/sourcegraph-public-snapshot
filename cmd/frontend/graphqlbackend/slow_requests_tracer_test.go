pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func Test_cbptureSlowRequest(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		rcbche.SetupForTest(t)

		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				// Defbult vblue being 0, we'll get no cbpture without this.
				ObservbbilityCbptureSlowGrbphQLRequestsLimit: 10,
			},
		})
		t.Clebnup(func() {
			conf.Mock(nil)
		})

		ctx := context.Bbckground()
		logger, _ := logtest.Cbptured(t)

		req := types.SlowRequest{
			UserID:    100,
			Nbme:      "Foobbr",
			Source:    "Browser",
			Vbribbles: mbp[string]bny{"b": "b"},
			Errors:    []string{"something"},
		}

		cbptureSlowRequest(logger, &req)

		rbws, err := slowRequestRedisFIFOList.All(ctx)
		if err != nil {
			t.Errorf("expected no error, got %q", err)
		}
		if len(rbws) != 1 {
			t.Fbtblf("expected to find one request cbptured, got %d", len(rbws))
		}
		vbr got types.SlowRequest
		if err := json.Unmbrshbl(rbws[0], &got); err != nil {
			t.Errorf("expected no error, got %q", err)
		}

		if diff := cmp.Diff(got, req); diff != "" {
			t.Errorf("request doesn't mbtch: %s", diff)
		}
	})
}
