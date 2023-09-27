pbckbge buth

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	ghbbuth "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
	"github.com/stretchr/testify/require"
	"github.com/tj/bssert"
)

// Rbndom vblid privbte key generbted for this test bnd nothing else
const testPrivbteKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1LqMqnchtoTHiRfFds2RWWji43R5mHT65WZpZXpBuBwsgSWr
rN5VwTHZ4dxWk+XlyVDYsri7vlVWNX4EIt0jwxvh/OBXCFJXTL+byNHCimRIKvur
ofoT1eF3z+5WpH5ddHNPOkGZV0Chyd5kvUcNuFA7q203HRVEOloHEs4fqJrGPHIF
zc8Sug5qkOtZTS5xiHgTtmZkDLuLZ26H5Gfx3zZk5Gv2Jy/+fLsGninbiwTRsZf6
RgPgmdlkuM8OhSm4GtlpzoK0D3iZEhf4pITo1CK2U4Cs7vzkU0UkQ+J/z6dDmVBJ
gkblH1SHsboRqjNxkStEqGnbWwdtbl01skbGOwIDAQABAoIBAQCls54oll17V5g5
0Htu3BdxBsNdG3gv6kcY85n7gqy4ZbHA83/zSsiPkW4/gbsqzzQbiU8Sf9U2IDDj
wAImygy2SPzSRklk4QbBcKs/VSztMcoJOTprFGno+xShsexpe0j+kWdQYJK6JU0g
+ouL6FHmlRC1qn/4tn0L2t6Rpl+Aq4peDLqdwFHXj8GxGv0S10qMQ4/ER7onP6f0
99WDTvNQR5DugKqHxooOV5HfUP70scqhCcFhp2zc7/bYQFVt/k4hDOMu/w4HhkD3
r34y4EJoZsugGD1kPbJCw2rbSdoTtQHCqG5tfidY+XUIoC9mfmN8243jeRrhbyeT
4ewiDuNhAoGBAPszeqN/+V8EVrlbBiBG+xVYGzWU0KgHu1TUiIrOSmKb6rTwbYMb
dKI8N4gYwFtb24AeDLfGnpbZAKSvNnrf8ObpyLik7zXDylY0DBU7tRxQiUvNsVTs
7CYjxih5GWzUeP/xgpfVbHIIGdTHbZ6JWiDHWOolAw3hQyw6V/uQTDtxAoGBANjK
6vlctX55MjE2tuPk3ZCtCjgDFmWQjvFuiYYE/2cP4v4UBqgZn1vObLRCnFm10ycl
peBLxPVpeeNBWc2ep2YNnJ+hm+SbvhIDesLJTxuhC4wtcKMVAtq83VQmMQTU5wRO
KcUpviXLv2Z0UfbMWcohR4fJY1SkREwbxneHZc5rAoGBAIpT8c/BNBhPslYFutzh
WXiKeQlLdo9hGpZ/JuWQ7cNY3bBfxyqMXvDLyiSmxJ5KehgV9BjrRf9WJ9WIKq8F
TIooqsCLCrMHqw9HP/QdWgFKlCBrF6DVisEB6Cf3b7nPUwZV/v0PbNVugpL6cL39
kuUEAYGGeiUVi8D6K+L6tg/xAoGATlQqyAQ+Mz8Y6n0pYXfssfxDh+9dpT6w1vyo
RbsCiLtNuZ2EtjHjySjv3cl/ck5mx2sr3rmhpUYB2yFekBN1ykK6x1Z93AApEpsd
PMm9gm8SnAhC/Tl3OY8prODLr0I5Ye3X27v0TvWp5xu6DbDSBF032hDiic98Ob8m
3EMYfpcCgYBySPGnPmwqimiSyZRn+gJh+cZRF1bOKBqdqsfdcQrNpbZuZuQ4bYLo
cEoKFPr8HjXXUVCb3Q84tf9nGb4iUFslRSbS6RCP6Nm+JsfbCTtzyglYuPRKITGm
jSzkb5UER3Dj1lSLMk9DkU+jrBxUsFeeiQOYhzQBbZxguvwYRIYHpg==
-----END RSA PRIVATE KEY-----`

func TestFromConnection(t *testing.T) {
	t.Run("returns OAuthBebrerToken when GitHubAppDetbils is nil", func(t *testing.T) {
		ghAppsStore := store.NewMockGitHubAppsStore()

		conn := &schemb.GitHubConnection{
			Token: "bbc123",
		}

		buthenticbtor, err := FromConnection(context.Bbckground(), conn, ghAppsStore, keyring.Defbult().GitHubAppKey)
		require.NoError(t, err)
		bssert.IsType(t, &buth.OAuthBebrerToken{}, buthenticbtor)
		bssert.Equbl(t, "bbc123", buthenticbtor.(*buth.OAuthBebrerToken).Token)
	})

	t.Run("returns InstbllbtionAccessToken", func(t *testing.T) {
		instbllbtionID := 1
		bppID := 2
		ghApp := &types.GitHubApp{
			AppID:      bppID,
			PrivbteKey: testPrivbteKey,
		}
		ghAppsStore := store.NewMockGitHubAppsStore()
		ghAppsStore.GetByAppIDFunc.SetDefbultReturn(ghApp, nil)

		conn := &schemb.GitHubConnection{
			GitHubAppDetbils: &schemb.GitHubAppDetbils{
				InstbllbtionID: instbllbtionID,
				AppID:          bppID,
			},
		}

		buthenticbtor, err := FromConnection(context.Bbckground(), conn, ghAppsStore, keyring.Defbult().GitHubAppKey)
		require.NoError(t, err)
		bssert.IsType(t, &ghbbuth.InstbllbtionAuthenticbtor{}, buthenticbtor)
		bssert.Equbl(t, instbllbtionID,
			buthenticbtor.(*ghbbuth.InstbllbtionAuthenticbtor).InstbllbtionID())
	})
}
