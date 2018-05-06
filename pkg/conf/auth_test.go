package conf

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAuthProvider(t *testing.T) {
	tests := map[string]struct {
		input schema.SiteConfiguration
		want  schema.AuthProviders
	}{
		"no auth.provider": {
			input: schema.SiteConfiguration{},
			want:  schema.AuthProviders{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
		},

		// builtin
		"auth.provider builtin": {
			input: schema.SiteConfiguration{AuthProvider: "builtin"},
			want:  schema.AuthProviders{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}},
		},
		"auth.provider builtin with allowSignup": {
			input: schema.SiteConfiguration{AuthProvider: "builtin", AuthAllowSignup: true},
			want:  schema.AuthProviders{Builtin: &schema.BuiltinAuthProvider{Type: "builtin", AllowSignup: true}},
		},

		// openidconnect
		"auth.provider openidconnect": {
			input: schema.SiteConfiguration{AuthProvider: "openidconnect", AuthOpenIDConnect: &schema.OpenIDConnectAuthProvider{Type: "openidconnect", ClientID: "v"}},
			want:  schema.AuthProviders{Openidconnect: &schema.OpenIDConnectAuthProvider{Type: "openidconnect", ClientID: "v"}},
		},

		// saml
		"auth.provider saml": {
			input: schema.SiteConfiguration{AuthProvider: "saml", AuthSaml: &schema.SAMLAuthProvider{Type: "saml", IdentityProviderMetadata: "v"}},
			want:  schema.AuthProviders{Saml: &schema.SAMLAuthProvider{Type: "saml", IdentityProviderMetadata: "v"}},
		},

		// http-header
		"auth.provider http-header": {
			input: schema.SiteConfiguration{AuthProvider: "http-header", AuthUserIdentityHTTPHeader: "v"},
			want:  schema.AuthProviders{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: "v"}},
		},

		// 🚨 SECURITY: Test that our backcompat helpers still apply. This is also tested elsewhere,
		// but we want to be extra sure to not regress because a mistake could cause upgraded
		// servers to lose their authentication config (and, e.g., expose private data).
		"auth.provider saml old": {
			input: schema.SiteConfiguration{
				SamlIDProviderMetadataURL: "a",
				SamlSPCert:                "b",
				SamlSPKey:                 "c",
			},
			want: schema.AuthProviders{Saml: &schema.SAMLAuthProvider{
				Type: "saml",
				IdentityProviderMetadataURL: "a",
				ServiceProviderCertificate:  "b",
				ServiceProviderPrivateKey:   "c",
			}},
		},
		"auth.provider openidconnect old": {
			input: schema.SiteConfiguration{
				OidcProvider:     "a",
				OidcClientID:     "b",
				OidcClientSecret: "c",
				OidcEmailDomain:  "d",
			},
			want: schema.AuthProviders{Openidconnect: &schema.OpenIDConnectAuthProvider{
				Type:               "openidconnect",
				Issuer:             "a",
				ClientID:           "b",
				ClientSecret:       "c",
				RequireEmailDomain: "d",
			}},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			got := authProvider(&test.input)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}
