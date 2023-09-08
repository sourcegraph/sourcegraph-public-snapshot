package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
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
	rcache.SetupForTest(t)
	installationID := 1
	appAuthenticator := &mockAuthenticator{}
	u, err := url.Parse("https://github.com")
	require.NoError(t, err)
	token := NewInstallationAccessToken(
		u,
		installationID,
		appAuthenticator,
		keyring.Default().GitHubAppKey,
	)
	token.installationAccessToken.Token = "installation-token"

	req, err := http.NewRequest("GET", "https://api.github.com", nil)
	require.NoError(t, err)

	require.NoError(t, token.Authenticate(req))

	assert.Equal(t, "Bearer installation-token", req.Header.Get("Authorization"))
}

func TestGitHubAppInstallationAuthenticator_Refresh(t *testing.T) {
	appAuthenticator := &mockAuthenticator{}
	u, err := url.Parse("https://github.com")
	require.NoError(t, err)

	t.Run("token refreshes", func(t *testing.T) {
		rcache.SetupForTest(t)

		token := NewInstallationAccessToken(
			u,
			1,
			appAuthenticator,
			keyring.Default().GitHubAppKey,
		)

		mockClient := &mockHTTPClient{}
		wantToken := mockClient.generateToken()
		require.NoError(t, token.Refresh(context.Background(), mockClient))

		require.True(t, mockClient.DoCalled)

		require.True(t, appAuthenticator.AuthenticateCalled)

		require.Equal(t, token.installationAccessToken.Token, wantToken.Token)
	})

	t.Run("uses token cache", func(t *testing.T) {
		rcache.SetupForTest(t)

		// We create 2 tokens for the same installation ID
		token1 := NewInstallationAccessToken(
			u,
			1,
			appAuthenticator,
			keyring.Default().GitHubAppKey,
		)
		token2 := NewInstallationAccessToken(
			u,
			1,
			appAuthenticator,
			keyring.Default().GitHubAppKey,
		)

		// We create a token for token1
		mockClient := &mockHTTPClient{}
		wantToken := mockClient.generateToken()
		require.NoError(t, token1.Refresh(context.Background(), mockClient))

		require.True(t, mockClient.DoCalled)
		require.True(t, appAuthenticator.AuthenticateCalled)

		require.Equal(t, token1.installationAccessToken.Token, wantToken.Token)

		// First we generate a new token for the mockClient so that we're sure
		// we're not returning the same one.
		mockClient.generateToken()
		require.NotEqual(t, mockClient.installationAccessToken, wantToken)
		// Now we refresh token2 and assert that we get the same token from the cache
		token2.Refresh(context.Background(), mockClient)
		require.Equal(t, token1.installationAccessToken.Token, token2.installationAccessToken.Token)
	})

	t.Run("refreshes cache if stale", func(t *testing.T) {
		rcache.SetupForTest(t)

		// We create 2 tokens for the same installation ID
		token1 := NewInstallationAccessToken(
			u,
			1,
			appAuthenticator,
			keyring.Default().GitHubAppKey,
		)
		token2 := NewInstallationAccessToken(
			u,
			1,
			appAuthenticator,
			keyring.Default().GitHubAppKey,
		)

		// We create a token for token1
		mockClient := &mockHTTPClient{}
		mockClient.generateToken()
		mockClient.installationAccessToken.ExpiresAt = time.Now().Add(-1 * time.Hour)
		wantToken := mockClient.installationAccessToken
		require.NoError(t, token1.Refresh(context.Background(), mockClient))

		require.True(t, mockClient.DoCalled)
		require.True(t, appAuthenticator.AuthenticateCalled)

		require.Equal(t, token1.installationAccessToken.Token, wantToken.Token)

		// First we generate a new token for the mockClient so that we're sure
		// we're not returning the same one.
		mockClient.generateToken()
		require.NotEqual(t, mockClient.installationAccessToken, wantToken)
		// Now we refresh token2 and assert that we get A DIFFERENT token from the cache
		token2.Refresh(context.Background(), mockClient)
		require.NotEqual(t, token1.installationAccessToken.Token, token2.installationAccessToken.Token)
		// For good measure, assert that token2 is not expired
		require.False(t, token2.NeedsRefresh())
	})
}

func TestInstallationAccessToken_NeedsRefresh(t *testing.T) {
	testCases := map[string]struct {
		token        InstallationAuthenticator
		needsRefresh bool
	}{
		"empty token":   {InstallationAuthenticator{}, true},
		"valid token":   {InstallationAuthenticator{installationAccessToken: installationAccessToken{Token: "abc123"}}, false},
		"not expired":   {InstallationAuthenticator{installationAccessToken: installationAccessToken{Token: "abc123", ExpiresAt: time.Now().Add(10 * time.Minute)}}, false},
		"expired":       {InstallationAuthenticator{installationAccessToken: installationAccessToken{Token: "abc123", ExpiresAt: time.Now().Add(-10 * time.Minute)}}, true},
		"expiring soon": {InstallationAuthenticator{installationAccessToken: installationAccessToken{Token: "abc123", ExpiresAt: time.Now().Add(3 * time.Minute)}}, true},
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

	it := InstallationAuthenticator{installationAccessToken: installationAccessToken{Token: token}}
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
	DoCalled                bool
	installationAccessToken installationAccessToken
}

func (c *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.DoCalled = true
	marshal, err := json.Marshal(c.installationAccessToken)
	if err != nil {
		return nil, err
	}

	return &http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(bytes.NewReader(marshal)),
	}, nil
}

func (c *mockHTTPClient) generateToken() installationAccessToken {
	c.installationAccessToken.Token = uuid.New().String()
	c.installationAccessToken.ExpiresAt = time.Now().Add(1 * time.Hour)
	return c.installationAccessToken
}
