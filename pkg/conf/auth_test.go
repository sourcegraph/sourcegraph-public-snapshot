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
			want:  schema.AuthProviders{},
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
		"openidconnect: provider not set": {
			input: schema.SiteConfiguration{
				AuthOpenIDConnect: &schema.OpenIDConnectAuthProvider{
					Issuer:             "a",
					ClientID:           "b",
					ClientSecret:       "c",
					RequireEmailDomain: "d",
					OverrideToken:      "e",
				},
			},
			want: schema.AuthProviders{},
		},
		"openidconnect: auth.provider none": {
			input: schema.SiteConfiguration{AuthProvider: "openidconnect"},
			want: schema.AuthProviders{Openidconnect: &schema.OpenIDConnectAuthProvider{
				Type: "openidconnect",
			}},
		},
		"openidconnect: old": {
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
		"openidconnect: auth.provider": {
			input: schema.SiteConfiguration{
				AuthProvider: "openidconnect",
				AuthOpenIDConnect: &schema.OpenIDConnectAuthProvider{
					Issuer:             "a",
					ClientID:           "b",
					ClientSecret:       "c",
					RequireEmailDomain: "d",
					OverrideToken:      "e",
				},
			},
			want: schema.AuthProviders{Openidconnect: &schema.OpenIDConnectAuthProvider{
				Type:               "openidconnect",
				Issuer:             "a",
				ClientID:           "b",
				ClientSecret:       "c",
				RequireEmailDomain: "d",
				OverrideToken:      "e",
			}},
		},
		"openidconnect: auth.provider and old": {
			input: schema.SiteConfiguration{
				AuthProvider:     "openidconnect",
				OidcProvider:     "a",
				OidcClientID:     "b",
				OidcClientSecret: "c",
				OidcEmailDomain:  "d",
				AuthOpenIDConnect: &schema.OpenIDConnectAuthProvider{
					Issuer:             "a2",
					ClientID:           "b2",
					ClientSecret:       "c2",
					RequireEmailDomain: "d2",
					OverrideToken:      "e2",
				},
			},
			want: schema.AuthProviders{Openidconnect: &schema.OpenIDConnectAuthProvider{
				Type:               "openidconnect",
				Issuer:             "a2",
				ClientID:           "b2",
				ClientSecret:       "c2",
				RequireEmailDomain: "d2",
				OverrideToken:      "e2",
			}},
		},
		"openidconnect: auth.providers": {
			input: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{{
					Openidconnect: &schema.OpenIDConnectAuthProvider{
						Type:               "openidconnect",
						Issuer:             "a",
						ClientID:           "b",
						ClientSecret:       "c",
						RequireEmailDomain: "d",
						OverrideToken:      "e",
					},
				}},
			},
			want: schema.AuthProviders{Openidconnect: &schema.OpenIDConnectAuthProvider{
				Type:               "openidconnect",
				Issuer:             "a",
				ClientID:           "b",
				ClientSecret:       "c",
				RequireEmailDomain: "d",
				OverrideToken:      "e",
			}},
		},
		"openidconnect: auth.provider and auth.providers": {
			input: schema.SiteConfiguration{
				AuthProvider: "openidconnect",
				AuthOpenIDConnect: &schema.OpenIDConnectAuthProvider{
					Issuer:             "a2",
					ClientID:           "b2",
					ClientSecret:       "c2",
					RequireEmailDomain: "d2",
					OverrideToken:      "e2",
				},
				AuthProviders: []schema.AuthProviders{{
					Openidconnect: &schema.OpenIDConnectAuthProvider{
						Type:               "openidconnect",
						Issuer:             "a",
						ClientID:           "b",
						ClientSecret:       "c",
						RequireEmailDomain: "d",
						OverrideToken:      "e",
					},
				}},
			},
			want: schema.AuthProviders{Openidconnect: &schema.OpenIDConnectAuthProvider{
				Type:               "openidconnect",
				Issuer:             "a",
				ClientID:           "b",
				ClientSecret:       "c",
				RequireEmailDomain: "d",
				OverrideToken:      "e",
			}},
		},
		"openidconnect: all": {
			input: schema.SiteConfiguration{
				OidcProvider:     "a3",
				OidcClientID:     "b3",
				OidcClientSecret: "c3",
				OidcEmailDomain:  "d3",
				AuthProvider:     "openidconnect",
				AuthOpenIDConnect: &schema.OpenIDConnectAuthProvider{
					Issuer:             "a2",
					ClientID:           "b2",
					ClientSecret:       "c2",
					RequireEmailDomain: "d2",
					OverrideToken:      "e2",
				},
				AuthProviders: []schema.AuthProviders{{
					Openidconnect: &schema.OpenIDConnectAuthProvider{
						Type:               "openidconnect",
						Issuer:             "a",
						ClientID:           "b",
						ClientSecret:       "c",
						RequireEmailDomain: "d",
						OverrideToken:      "e",
					},
				}},
			},
			want: schema.AuthProviders{Openidconnect: &schema.OpenIDConnectAuthProvider{
				Type:               "openidconnect",
				Issuer:             "a",
				ClientID:           "b",
				ClientSecret:       "c",
				RequireEmailDomain: "d",
				OverrideToken:      "e",
			}},
		},

		// saml
		"saml: provider not set": {
			input: schema.SiteConfiguration{
				AuthSaml: &schema.SAMLAuthProvider{
					IdentityProviderMetadataURL: "a",
					ServiceProviderCertificate:  "b",
					ServiceProviderPrivateKey:   "c",
				},
			},
			want: schema.AuthProviders{},
		},
		"saml: auth.provider none": {
			input: schema.SiteConfiguration{AuthProvider: "saml"},
			want: schema.AuthProviders{Saml: &schema.SAMLAuthProvider{
				Type: "saml",
			}},
		},
		"saml: old": {
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
		"saml: auth.provider": {
			input: schema.SiteConfiguration{
				AuthProvider: "saml",
				AuthSaml: &schema.SAMLAuthProvider{
					IdentityProviderMetadataURL: "a",
					ServiceProviderCertificate:  "b",
					ServiceProviderPrivateKey:   "c",
				},
			},
			want: schema.AuthProviders{Saml: &schema.SAMLAuthProvider{
				Type: "saml",
				IdentityProviderMetadataURL: "a",
				ServiceProviderCertificate:  "b",
				ServiceProviderPrivateKey:   "c",
			}},
		},
		"saml: auth.provider and old": {
			input: schema.SiteConfiguration{
				AuthProvider:              "saml",
				SamlIDProviderMetadataURL: "a",
				SamlSPCert:                "b",
				SamlSPKey:                 "c",
				AuthSaml: &schema.SAMLAuthProvider{
					IdentityProviderMetadataURL: "a2",
					ServiceProviderCertificate:  "b2",
					ServiceProviderPrivateKey:   "c2",
				},
			},
			want: schema.AuthProviders{Saml: &schema.SAMLAuthProvider{
				Type: "saml",
				IdentityProviderMetadataURL: "a2",
				ServiceProviderCertificate:  "b2",
				ServiceProviderPrivateKey:   "c2",
			}},
		},
		"saml: auth.providers": {
			input: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{{
					Saml: &schema.SAMLAuthProvider{
						Type: "saml",
						IdentityProviderMetadataURL: "a",
						ServiceProviderCertificate:  "b",
						ServiceProviderPrivateKey:   "c",
					},
				}},
			},
			want: schema.AuthProviders{Saml: &schema.SAMLAuthProvider{
				Type: "saml",
				IdentityProviderMetadataURL: "a",
				ServiceProviderCertificate:  "b",
				ServiceProviderPrivateKey:   "c",
			}},
		},
		"saml: auth.provider and auth.providers": {
			input: schema.SiteConfiguration{
				AuthProvider: "saml",
				AuthSaml: &schema.SAMLAuthProvider{
					IdentityProviderMetadataURL: "a2",
					ServiceProviderCertificate:  "b2",
					ServiceProviderPrivateKey:   "c2",
				},
				AuthProviders: []schema.AuthProviders{{
					Saml: &schema.SAMLAuthProvider{
						Type: "saml",
						IdentityProviderMetadataURL: "a",
						ServiceProviderCertificate:  "b",
						ServiceProviderPrivateKey:   "c",
					},
				}},
			},
			want: schema.AuthProviders{Saml: &schema.SAMLAuthProvider{
				Type: "saml",
				IdentityProviderMetadataURL: "a",
				ServiceProviderCertificate:  "b",
				ServiceProviderPrivateKey:   "c",
			}},
		},
		"saml: all": {
			input: schema.SiteConfiguration{
				SamlIDProviderMetadataURL: "a3",
				SamlSPCert:                "b3",
				SamlSPKey:                 "c3",
				AuthProvider:              "saml",
				AuthSaml: &schema.SAMLAuthProvider{
					IdentityProviderMetadataURL: "a2",
					ServiceProviderCertificate:  "b2",
					ServiceProviderPrivateKey:   "c2",
				},
				AuthProviders: []schema.AuthProviders{{
					Saml: &schema.SAMLAuthProvider{
						Type: "saml",
						IdentityProviderMetadataURL: "a",
						ServiceProviderCertificate:  "b",
						ServiceProviderPrivateKey:   "c",
					},
				}},
			},
			want: schema.AuthProviders{Saml: &schema.SAMLAuthProvider{
				Type: "saml",
				IdentityProviderMetadataURL: "a",
				ServiceProviderCertificate:  "b",
				ServiceProviderPrivateKey:   "c",
			}},
		},

		// http-header
		"http-header: provider not set": {
			input: schema.SiteConfiguration{AuthUserIdentityHTTPHeader: "a"},
			want:  schema.AuthProviders{},
		},
		"http-header: none": {
			// This config would produce a runtime error in the auth middleware.
			input: schema.SiteConfiguration{AuthProvider: "http-header"},
			want:  schema.AuthProviders{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: ""}},
		},
		"http-header: auth.provider": {
			input: schema.SiteConfiguration{AuthProvider: "http-header", AuthUserIdentityHTTPHeader: "a"},
			want:  schema.AuthProviders{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: "a"}},
		},
		"http-header: auth.provider only, no header": {
			input: schema.SiteConfiguration{AuthProvider: "http-header", AuthUserIdentityHTTPHeader: ""},
			want:  schema.AuthProviders{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: ""}},
		},
		"http-header: auth.userIdentityHTTPHeader only": {
			input: schema.SiteConfiguration{AuthUserIdentityHTTPHeader: "a"},
			want:  schema.AuthProviders{},
		},
		"http-header: auth.providers": {
			input: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{{
					HttpHeader: &schema.HTTPHeaderAuthProvider{
						Type:           "http-header",
						UsernameHeader: "a",
					},
				}},
			},
			want: schema.AuthProviders{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: "a"}},
		},
		"http-header: auth.provider and auth.providers": {
			input: schema.SiteConfiguration{
				AuthProvider:               "http-header",
				AuthUserIdentityHTTPHeader: "a2",
				AuthProviders: []schema.AuthProviders{{
					HttpHeader: &schema.HTTPHeaderAuthProvider{
						Type:           "http-header",
						UsernameHeader: "a",
					},
				}},
			},
			want: schema.AuthProviders{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: "a"}},
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
				t.Errorf("got != want\ngot  %+v\nwant %+v", got, test.want)
			}
		})
	}
}

func TestAuthPublic(t *testing.T) {
	tests := map[string]struct {
		input schema.SiteConfiguration
		want  bool
	}{
		"false": {
			input: schema.SiteConfiguration{AuthPublic: false},
			want:  false,
		},
		"true, no auth provider": {
			input: schema.SiteConfiguration{AuthPublic: true},
			want:  false,
		},
		"true, non-builtin auth provider": {
			input: schema.SiteConfiguration{AuthPublic: true, AuthProviders: []schema.AuthProviders{{Openidconnect: &schema.OpenIDConnectAuthProvider{}}}},
			want:  false,
		},
		"true, builtin auth provider": {
			input: schema.SiteConfiguration{AuthPublic: true, AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}},
			want:  true,
		},
		"false, builtin auth provider": {
			input: schema.SiteConfiguration{AuthPublic: false, AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}},
			want:  false,
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			got := authPublic(&test.input)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}
