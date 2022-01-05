package conf

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

var secret = "secret-to-redact"

func TestRedactSecrets(t *testing.T) {
	t.Run("successfully redacts secrets", func(t *testing.T) {
		store := newStore()
		server := Server{
			Source: mockConfigSource{},
			store:  store,
		}
		server.Start()
		redacted, err := server.RedactSecrets()
		if err != nil {
			t.Errorf("unexpected error redacting secrets from config: %s", err)
		}
		if strings.Contains(redacted.Site, secret) {
			t.Errorf("secret should've been redacted from config")
		}
		if !strings.Contains(redacted.Site, RedactedSecret) {
			t.Errorf("expected to see REDACTED in place of secret in config")
		}
	})
}

type mockConfigSource struct{}

func (m mockConfigSource) Write(ctx context.Context, data conftypes.RawUnified) error {
	return nil
}

func (m mockConfigSource) Read(ctx context.Context) (conftypes.RawUnified, error) {
	unified := Unified{SiteConfiguration: schema.SiteConfiguration{
		ExecutorsAccessToken: secret,
		AuthProviders: []schema.AuthProviders{
			{
				Openidconnect: &schema.OpenIDConnectAuthProvider{
					ClientSecret: secret,
					Type:         "openidconnect",
				},
			},
			{
				Gitlab: &schema.GitLabAuthProvider{
					ClientSecret: secret,
					Type:         "gitlab",
				},
			},
			{
				Github: &schema.GitHubAuthProvider{
					ClientSecret: secret,
					Type:         "github",
				},
			},
		},
	}}
	site, err := jsonc.Edit("", unified)
	return conftypes.RawUnified{Site: site}, err
}
