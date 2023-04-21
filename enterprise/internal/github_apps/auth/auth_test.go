package auth

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Random valid private key generated for this test and nothing else
const testPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1LqMqnchtoTHiRfFds2RWWji43R5mHT65WZpZXpBuBwsgSWr
rN5VwTHZ4dxWk+XlyVDYsri7vlVWNX4EIt0jwxvh/OBXCFJXTL+byNHCimRIKvur
ofoT1eF3z+5WpH5ddHNPOkGZV0Chyd5kvUcNuFA7q203HRVEOloHEs4fqJrGPHIF
zc8Sug5qkOtZTS5xiHgTtmZkDLuLZ26H5Gfx3zZk5Gv2Jy/+fLsGninaiwTRsZf6
RgPgmdlkuM8OhSm4GtlpzoK0D3iZEhf4pITo1CK2U4Cs7vzkU0UkQ+J/z6dDmVBJ
gkalH1SHsboRqjNxkStEqGnbWwdtal01skbGOwIDAQABAoIBAQCls54oll17V5g5
0Htu3BdxBsNdG3gv6kcY85n7gqy4ZbHA83/zSsiPkW4/gasqzzQbiU8Sf9U2IDDj
wAImygy2SPzSRklk4QbBcKs/VSztMcoJOTprFGno+xShsexpe0j+kWdQYJK6JU0g
+ouL6FHmlRC1qn/4tn0L2t6Rpl+Aq4peDLqdwFHXj8GxGv0S10qMQ4/ER7onP6f0
99WDTvNQR5DugKqHxooOV5HfUP70scqhCcFhp2zc7/aYQFVt/k4hDOMu/w4HhkD3
r34y4EJoZsugGD1kPaJCw2rbSdoTtQHCqG5tfidY+XUIoC9mfmN8243jeRrhayeT
4ewiDuNhAoGBAPszeqN/+V8EVrlbBiBG+xVYGzWU0KgHu1TUiIrOSmKa6rTwaYMb
dKI8N4gYwFtb24AeDLfGnpaZAKSvNnrf8OapyLik7zXDylY0DBU7tRxQiUvNsVTs
7CYjxih5GWzUeP/xgpfVbHIIGdTHaZ6JWiDHWOolAw3hQyw6V/uQTDtxAoGBANjK
6vlctX55MjE2tuPk3ZCtCjgDFmWQjvFuiYYE/2cP4v4UBqgZn1vOaLRCnFm10ycl
peBLxPVpeeNBWc2ep2YNnJ+hm+SavhIDesLJTxuhC4wtcKMVAtq83VQmMQTU5wRO
KcUpviXLv2Z0UfbMWcohR4fJY1SkREwaxneHZc5rAoGBAIpT8c/BNBhPslYFutzh
WXiKeQlLdo9hGpZ/JuWQ7cNY3bBfxyqMXvDLyiSmxJ5KehgV9BjrRf9WJ9WIKq8F
TIooqsCLCrMHqw9HP/QdWgFKlCBrF6DVisEB6Cf3b7nPUwZV/v0PaNVugpL6cL39
kuUEAYGGeiUVi8D6K+L6tg/xAoGATlQqyAQ+Mz8Y6n0pYXfssfxDh+9dpT6w1vyo
RbsCiLtNuZ2EtjHjySjv3cl/ck5mx2sr3rmhpUYB2yFekBN1ykK6x1Z93AApEpsd
PMm9gm8SnAhC/Tl3OY8prODLr0I5Ye3X27v0TvWp5xu6DaDSBF032hDiic98Ob8m
3EMYfpcCgYBySPGnPmwqimiSyZRn+gJh+cZRF1aOKBqdqsfdcQrNpaZuZuQ4aYLo
cEoKFPr8HjXXUVCa3Q84tf9nGb4iUFslRSbS6RCP6Nm+JsfbCTtzyglYuPRKITGm
jSzka5UER3Dj1lSLMk9DkU+jrBxUsFeeiQOYhzQBaZxguvwYRIYHpg==
-----END RSA PRIVATE KEY-----`

func TestCreateEnterpriseFromConnection(t *testing.T) {
	t.Run("returns OAuthBearerToken when GitHubAppDetails is nil", func(t *testing.T) {
		ghAppsStore := store.NewMockGitHubAppsStore()
		fromConnection := CreateEnterpriseFromConnection(ghAppsStore)

		conn := &schema.GitHubConnection{
			Token: "abc123",
		}

		authenticator, err := fromConnection(context.Background(), conn)
		require.NoError(t, err)
		assert.IsType(t, &auth.OAuthBearerToken{}, authenticator)
		assert.Equal(t, "abc123", authenticator.(*auth.OAuthBearerToken).Token)
	})

	t.Run("returns InstallationAccessToken", func(t *testing.T) {
		installationID := 1
		appID := 2
		ghApp := &types.GitHubApp{
			AppID:      appID,
			PrivateKey: testPrivateKey,
		}
		ghAppsStore := store.NewMockGitHubAppsStore()
		ghAppsStore.GetByAppIDFunc.SetDefaultReturn(ghApp, nil)
		fromConnection := CreateEnterpriseFromConnection(ghAppsStore)

		conn := &schema.GitHubConnection{
			GitHubAppDetails: &schema.GitHubAppDetails{
				InstallationID: installationID,
				AppID:          appID,
			},
		}

		authenticator, err := fromConnection(context.Background(), conn)
		require.NoError(t, err)
		assert.IsType(t, &installationAccessToken{}, authenticator)
		assert.Equal(t, installationID,
			authenticator.(*installationAccessToken).installationID)
	})
}

func TestGitHubAppAuthenticator_Authenticate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		appID := 1234
		privateKey := []byte(testPrivateKey)
		authenticator, err := NewGitHubAppAuthenticator(appID, privateKey)
		require.NoError(t, err)

		req, err := http.NewRequest("GET", "https://api.github.com", nil)
		require.NoError(t, err)

		require.NoError(t, authenticator.Authenticate(req))

		assert.True(t, strings.HasPrefix(req.Header.Get("Authorization"), "Bearer "))
	})

	t.Run("invalid private key", func(t *testing.T) {
		appID := 1234
		privateKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
invalid key
-----END RSA PRIVATE KEY-----`)
		_, err := NewGitHubAppAuthenticator(appID, privateKey)
		require.Error(t, err)
	})
}

func TestGitHubAppInstallationAuthenticator_Authenticate(t *testing.T) {
	installationID := 1
	appAuthenticator := &mockAuthenticator{}
	u, err := url.Parse("https://github.com")
	require.NoError(t, err)
	token := NewInstallationAccessToken(
		u,
		installationID,
		appAuthenticator,
	)
	token.token = "installation-token"

	req, err := http.NewRequest("GET", "https://api.github.com", nil)
	require.NoError(t, err)

	require.NoError(t, token.Authenticate(req))

	assert.Equal(t, "Bearer installation-token", req.Header.Get("Authorization"))
}

func TestGitHubAppInstallationAuthenticator_Refresh(t *testing.T) {
	appAuthenticator := &mockAuthenticator{}
	u, err := url.Parse("https://github.com")
	require.NoError(t, err)
	token := NewInstallationAccessToken(
		u,
		1,
		appAuthenticator,
	)

	mockClient := &mockHTTPClient{}
	require.NoError(t, token.Refresh(context.Background(), mockClient))

	require.True(t, mockClient.DoCalled)

	require.True(t, appAuthenticator.AuthenticateCalled)

	require.Equal(t, token.token, "new-token")
	wantTime, err := time.Parse(time.RFC3339, "2016-07-11T22:14:10Z")
	require.NoError(t, err)
	require.True(t, token.expiresAt.Equal(wantTime))
}

func TestInstallationAccessToken_NeedsRefresh(t *testing.T) {
	testCases := map[string]struct {
		token        installationAccessToken
		needsRefresh bool
	}{
		"empty token":   {installationAccessToken{}, true},
		"valid token":   {installationAccessToken{token: "abc123"}, false},
		"not expired":   {installationAccessToken{token: "abc123", expiresAt: time.Now().Add(10 * time.Minute)}, false},
		"expired":       {installationAccessToken{token: "abc123", expiresAt: time.Now().Add(-10 * time.Minute)}, true},
		"expiring soon": {installationAccessToken{token: "abc123", expiresAt: time.Now().Add(3 * time.Minute)}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.needsRefresh, tc.token.NeedsRefresh())
		})
	}
}

func TestInstallationAccessToken_SetURLUser(t *testing.T) {
	token := "abc123"
	u, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	it := installationAccessToken{token: token}
	it.SetURLUser(u)

	want := "x-access-token:abc123"
	if got := u.User.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

type mockAuthenticator struct {
	AuthenticateCalled bool
}

func (m *mockAuthenticator) Authenticate(r *http.Request) error {
	m.AuthenticateCalled = true
	return nil
}

func (m *mockAuthenticator) Hash() string {
	return ""
}

type mockHTTPClient struct {
	DoCalled bool
}

func (c *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.DoCalled = true

	return &http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(strings.NewReader(`{"token": "new-token", "expires_at": "2016-07-11T22:14:10Z"}`)),
	}, nil
}
