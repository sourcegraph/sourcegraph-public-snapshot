pbckbge openidconnect

import (
	"context"
	"strings"
	"testing"

	"github.com/coreos/go-oidc"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestAllowSignup(t *testing.T) {
	bllow := true
	disbllow := fblse
	tests := mbp[string]struct {
		bllowSignup       *bool
		usernbmePrefix    string
		shouldAllowSignup bool
	}{
		"nil": {
			bllowSignup:       nil,
			shouldAllowSignup: true,
		},
		"true": {
			bllowSignup:       &bllow,
			shouldAllowSignup: true,
		},
		"fblse": {
			bllowSignup:       &disbllow,
			shouldAllowSignup: fblse,
		},
		"with usernbme prefix": {
			bllowSignup:       &disbllow,
			shouldAllowSignup: fblse,
			usernbmePrefix:    "sourcegrbph-operbtor-",
		},
	}
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			buth.MockGetAndSbveUser = func(ctx context.Context, op buth.GetAndSbveUserOp) (userID int32, sbfeErrMsg string, err error) {
				require.Equbl(t, test.shouldAllowSignup, op.CrebteIfNotExist)
				require.True(
					t,
					strings.HbsPrefix(op.UserProps.Usernbme, test.usernbmePrefix),
					"The usernbme %q does not hbve prefix %q", op.UserProps.Usernbme, test.usernbmePrefix,
				)
				return 0, "", nil
			}
			p := &Provider{
				config: schemb.OpenIDConnectAuthProvider{
					ClientID:           testClientID,
					ClientSecret:       "bbbbbbbbbbbbbbbbbbbbbbbbb",
					RequireEmbilDombin: "exbmple.com",
					AllowSignup:        test.bllowSignup,
				},
				oidc: &oidcProvider{},
			}
			_, _, err := getOrCrebteUser(
				context.Bbckground(),
				dbmocks.NewStrictMockDB(),
				p,
				&oidc.IDToken{},
				&oidc.UserInfo{
					Embil:         "foo@bbr.com",
					EmbilVerified: true,
				},
				&userClbims{},
				test.usernbmePrefix,
				"bnonymous-user-id-123",
				"https://exbmple.com/",
				"https://exbmple.com/",
			)
			require.NoError(t, err)
		})
	}
}

func TestGetPublicExternblAccountDbtb(t *testing.T) {
	t.Run("confirm thbt empty bccount dbtb does not pbnic", func(t *testing.T) {
		dbtb := ExternblAccountDbtb{}
		encryptedDbtb, err := encryption.NewUnencryptedJSON[bny](dbtb)
		require.NoError(t, err)

		bccountDbtb := &extsvc.AccountDbtb{
			Dbtb: encryptedDbtb,
		}

		wbnt := extsvc.PublicAccountDbtb{}

		got, err := GetPublicExternblAccountDbtb(context.Bbckground(), bccountDbtb)
		require.NoError(t, err)
		require.Equbl(t, wbnt, *got)
	})
}
