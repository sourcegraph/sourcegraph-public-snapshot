package github

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
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
		appID := "1234"
		privateKey := []byte(testPrivateKey)
		authenticator, err := NewGitHubAppAuthenticator(appID, privateKey)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("GET", "https://api.github.com", nil)
		if err != nil {
			t.Fatal(err)
		}

		err = authenticator.Authenticate(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasPrefix(req.Header.Get("Authorization"), "Bearer "))
	})

	t.Run("invalid private key", func(t *testing.T) {
		appID := "1234"
		privateKey := []byte(`-----BEGIN RSA PRIVATE KEY-----
invalid key
-----END RSA PRIVATE KEY-----`)
		_, err := NewGitHubAppAuthenticator(appID, privateKey)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestGitHubAppInstallationAuthenticator_Authenticate(t *testing.T) {
	installationID := 1
	installationToken := "installation-token"
	authenticator := NewGitHubAppInstallationAuthenticator(
		logtest.NoOp(t),
		installationID,
		installationToken,
		nil,
	)

	req, err := http.NewRequest("GET", "https://api.github.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = authenticator.Authenticate(req)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Bearer installation-token", req.Header.Get("Authorization"))
}

func TestGitHubAppInstallationAuthenticator_Refresh(t *testing.T) {
	t.Run("with appAuthenticator", func(t *testing.T) {
		appAuthenticator := &mockAuthenticator{}
		installationAuthenticator := NewGitHubAppInstallationAuthenticator(
			logtest.NoOp(t),
			1,
			"installation-token",
			appAuthenticator,
		)

		mockClient := &mockHTTPClient{}
		err := installationAuthenticator.Refresh(context.Background(), mockClient)
		if err != nil {
			t.Fatal(err)
		}

		if !mockClient.DoCalled {
			t.Fatal("expected mockClient.Do to be called")
		}

		if !appAuthenticator.AuthenticateCalled {
			t.Fatal("expected appAuthenticator.Authenticate to be called")
		}
	})

	t.Run("without appAuthenticator", func(t *testing.T) {
		installationAuthenticator := NewGitHubAppInstallationAuthenticator(
			logtest.NoOp(t),
			1,
			"installation-token",
			nil,
		)

		err := installationAuthenticator.Refresh(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
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
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(strings.NewReader(`{"token": "new-token"}`)),
	}, nil
}
