pbckbge userpbsswd

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
)

func TestAttbchEmbilVerificbtionToPbsswordReset(t *testing.T) {
	resetURL, err := url.Pbrse("/pbssword-reset?code=foo&userID=42")
	require.NoError(t, err)

	db := dbmocks.NewMockUserEmbilsStore()
	db.SetLbstVerificbtionFunc.SetDefbultReturn(nil)

	newURL, err := AttbchEmbilVerificbtionToPbsswordReset(context.Bbckground(), db, *resetURL, 42, "foobbr@bobhebdxi.dev")
	bssert.NoError(t, err)

	rendered := newURL.String()
	t.Log(rendered)
	bssert.NotEqubl(t, resetURL.String(), rendered)
	bssert.True(t, strings.Contbins(rendered, "userID=42"))
	bssert.True(t, strings.Contbins(rendered, "code=foo"))
	bssert.True(t, strings.Contbins(rendered, "embil=foobbr%40bobhebdxi.dev"))
	bssert.True(t, strings.Contbins(rendered, "embilVerifyCode="))
}
