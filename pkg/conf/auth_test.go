package conf

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAuthProvider(t *testing.T) {
	tests := map[string]struct {
		input schema.SiteConfiguration
		want  []schema.AuthProviders
	}{
		"no auth.provider": {
			input: schema.SiteConfiguration{},
			want:  nil,
		},

		// unrecognized
		"unrecognized: auth.provider": {
			input: schema.SiteConfiguration{AuthProvider: "asdf"},
			want:  []schema.AuthProviders(nil),
		},

		// builtin
		"auth.provider builtin": {
			input: schema.SiteConfiguration{AuthProvider: "builtin"},
			want:  []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{Type: "builtin"}}},
		},
		"auth.provider builtin with allowSignup": {
			input: schema.SiteConfiguration{AuthProvider: "builtin", AuthAllowSignup: true},
			want:  []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{Type: "builtin", AllowSignup: true}}},
		},

		// openidconnect
		"openidconnect: provider not set": {
			input: schema.SiteConfiguration{
				AuthOpenIDConnect: &schema.OpenIDConnectAuthProvider{
					Issuer:             "a",
					ClientID:           "b",
					ClientSecret:       "c",
					RequireEmailDomain: "d",
				},
			},
			want: nil,
		},
		"openidconnect: auth.provider none": {
			input: schema.SiteConfiguration{AuthProvider: "openidconnect"},
			want: []schema.AuthProviders{{Openidconnect: &schema.OpenIDConnectAuthProvider{
				Type: "openidconnect",
			}}},
		},
		"openidconnect: auth.provider": {
			input: schema.SiteConfiguration{
				AuthProvider: "openidconnect",
				AuthOpenIDConnect: &schema.OpenIDConnectAuthProvider{
					Issuer:             "a",
					ClientID:           "b",
					ClientSecret:       "c",
					RequireEmailDomain: "d",
				},
			},
			want: []schema.AuthProviders{{Openidconnect: &schema.OpenIDConnectAuthProvider{
				Type:               "openidconnect",
				Issuer:             "a",
				ClientID:           "b",
				ClientSecret:       "c",
				RequireEmailDomain: "d",
			}}},
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
					},
				}},
			},
			want: []schema.AuthProviders{{Openidconnect: &schema.OpenIDConnectAuthProvider{
				Type:               "openidconnect",
				Issuer:             "a",
				ClientID:           "b",
				ClientSecret:       "c",
				RequireEmailDomain: "d",
			}}},
		},
		"openidconnect: auth.provider and auth.providers": {
			input: schema.SiteConfiguration{
				AuthProvider: "openidconnect",
				AuthOpenIDConnect: &schema.OpenIDConnectAuthProvider{
					Issuer:             "a2",
					ClientID:           "b2",
					ClientSecret:       "c2",
					RequireEmailDomain: "d2",
				},
				AuthProviders: []schema.AuthProviders{{
					Openidconnect: &schema.OpenIDConnectAuthProvider{
						Type:               "openidconnect",
						Issuer:             "a",
						ClientID:           "b",
						ClientSecret:       "c",
						RequireEmailDomain: "d",
					},
				}},
			},
			want: []schema.AuthProviders{{Openidconnect: &schema.OpenIDConnectAuthProvider{
				Type:               "openidconnect",
				Issuer:             "a",
				ClientID:           "b",
				ClientSecret:       "c",
				RequireEmailDomain: "d",
			}}},
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
			want: nil,
		},
		"saml: auth.provider none": {
			input: schema.SiteConfiguration{AuthProvider: "saml"},
			want: []schema.AuthProviders{{Saml: &schema.SAMLAuthProvider{
				Type: "saml",
			}}},
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
			want: []schema.AuthProviders{{Saml: &schema.SAMLAuthProvider{
				Type:                        "saml",
				IdentityProviderMetadataURL: "a",
				ServiceProviderCertificate:  "b",
				ServiceProviderPrivateKey:   "c",
			}}},
		},
		"saml: auth.providers": {
			input: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{{
					Saml: &schema.SAMLAuthProvider{
						Type:                        "saml",
						IdentityProviderMetadataURL: "a",
						ServiceProviderCertificate:  "b",
						ServiceProviderPrivateKey:   "c",
					},
				}},
			},
			want: []schema.AuthProviders{{Saml: &schema.SAMLAuthProvider{
				Type:                        "saml",
				IdentityProviderMetadataURL: "a",
				ServiceProviderCertificate:  "b",
				ServiceProviderPrivateKey:   "c",
			}}},
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
						Type:                        "saml",
						IdentityProviderMetadataURL: "a",
						ServiceProviderCertificate:  "b",
						ServiceProviderPrivateKey:   "c",
					},
				}},
			},
			want: []schema.AuthProviders{{Saml: &schema.SAMLAuthProvider{
				Type:                        "saml",
				IdentityProviderMetadataURL: "a",
				ServiceProviderCertificate:  "b",
				ServiceProviderPrivateKey:   "c",
			}}},
		},

		// http-header
		"http-header: provider not set": {
			input: schema.SiteConfiguration{AuthUserIdentityHTTPHeader: "a"},
			want:  nil,
		},
		"http-header: none": {
			// This config would produce a runtime error in the auth middleware.
			input: schema.SiteConfiguration{AuthProvider: "http-header"},
			want:  []schema.AuthProviders{{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: ""}}},
		},
		"http-header: auth.provider": {
			input: schema.SiteConfiguration{AuthProvider: "http-header", AuthUserIdentityHTTPHeader: "a"},
			want:  []schema.AuthProviders{{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: "a"}}},
		},
		"http-header: auth.provider only, no header": {
			input: schema.SiteConfiguration{AuthProvider: "http-header", AuthUserIdentityHTTPHeader: ""},
			want:  []schema.AuthProviders{{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: ""}}},
		},
		"http-header: auth.userIdentityHTTPHeader only": {
			input: schema.SiteConfiguration{AuthUserIdentityHTTPHeader: "a"},
			want:  nil,
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
			want: []schema.AuthProviders{{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: "a"}}},
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
			want: []schema.AuthProviders{{HttpHeader: &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: "a"}}},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			got := AuthProvidersFromConfig(&test.input)
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
