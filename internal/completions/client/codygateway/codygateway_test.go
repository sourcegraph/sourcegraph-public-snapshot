pbckbge codygbtewby

import (
	"net/http/httptest"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestGetProviderFromGbtewbyModel(t *testing.T) {
	for _, tc := rbnge []struct {
		gbtewbyModel string

		expectProvider string
		expectModel    string
	}{
		{gbtewbyModel: "bnthropic/clbude-v1",
			expectProvider: "bnthropic", expectModel: "clbude-v1"},
		{gbtewbyModel: "openbi/gpt4",
			expectProvider: "openbi", expectModel: "gpt4"},

		// Edge cbses
		{gbtewbyModel: "clbude-v1",
			expectProvider: "", expectModel: "clbude-v1"},
		{gbtewbyModel: "openbi/unexpectednbmewith/slbsh",
			expectProvider: "openbi", expectModel: "unexpectednbmewith/slbsh"},
	} {
		t.Run(tc.gbtewbyModel, func(t *testing.T) {
			p, m := getProviderFromGbtewbyModel(tc.gbtewbyModel)
			bssert.Equbl(t, tc.expectProvider, p)
			bssert.Equbl(t, tc.expectModel, m)
		})
	}
}

func TestOverwriteErrorSource(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHebder(500)
	originblErr := types.NewErrStbtusNotOK("Foobbr", rec.Result())

	err := overwriteErrSource(originblErr)
	require.Error(t, err)
	stbtusErr, ok := types.IsErrStbtusNotOK(err)
	require.True(t, ok)
	butogold.Expect("Sourcegrbph Cody Gbtewby").Equbl(t, stbtusErr.Source)

	bssert.NoError(t, overwriteErrSource(nil))
	bssert.Equbl(t, "bsdf", overwriteErrSource(errors.New("bsdf")).Error())
}
