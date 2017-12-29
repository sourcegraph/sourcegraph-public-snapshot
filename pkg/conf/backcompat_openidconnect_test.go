package conf

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

func TestAuthOpenIDConnect(t *testing.T) {
	tests := map[string]struct {
		input schema.SiteConfiguration
		want  *schema.OpenIDConnectAuthProvider
	}{
		"none": {
			input: schema.SiteConfiguration{},
			want:  nil,
		},
		"old": {
			input: schema.SiteConfiguration{
				OidcProvider:      "a",
				OidcClientID:      "b",
				OidcClientSecret:  "c",
				OidcEmailDomain:   "d",
				OidcOverrideToken: "e",
			},
			want: &schema.OpenIDConnectAuthProvider{
				Issuer:             "a",
				ClientID:           "b",
				ClientSecret:       "c",
				RequireEmailDomain: "d",
				OverrideToken:      "e",
			},
		},
		"new": {
			input: schema.SiteConfiguration{
				AuthOpenIDConnect: &schema.OpenIDConnectAuthProvider{
					Issuer:             "a",
					ClientID:           "b",
					ClientSecret:       "c",
					RequireEmailDomain: "d",
					OverrideToken:      "e",
				},
			},
			want: &schema.OpenIDConnectAuthProvider{
				Issuer:             "a",
				ClientID:           "b",
				ClientSecret:       "c",
				RequireEmailDomain: "d",
				OverrideToken:      "e",
			},
		},
		"both": {
			input: schema.SiteConfiguration{
				OidcProvider:      "a",
				OidcClientID:      "b",
				OidcClientSecret:  "c",
				OidcEmailDomain:   "d",
				OidcOverrideToken: "e",
				AuthOpenIDConnect: &schema.OpenIDConnectAuthProvider{
					Issuer:             "a2",
					ClientID:           "b2",
					ClientSecret:       "c2",
					RequireEmailDomain: "d2",
					OverrideToken:      "e2",
				},
			},
			want: &schema.OpenIDConnectAuthProvider{
				Issuer:             "a2",
				ClientID:           "b2",
				ClientSecret:       "c2",
				RequireEmailDomain: "d2",
				OverrideToken:      "e2",
			},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			got := authOpenIDConnect(test.input)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}
