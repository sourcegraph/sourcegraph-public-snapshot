package conf

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

func TestAuthSAML(t *testing.T) {
	tests := map[string]struct {
		input schema.SiteConfiguration
		want  *schema.SAMLAuthProvider
	}{
		"none": {
			input: schema.SiteConfiguration{},
			want:  nil,
		},
		"old": {
			input: schema.SiteConfiguration{
				SamlIDProviderMetadataURL: "a",
				SamlSPCert:                "b",
				SamlSPKey:                 "c",
			},
			want: &schema.SAMLAuthProvider{
				IdentityProviderMetadataURL: "a",
				ServiceProviderCertificate:  "b",
				ServiceProviderPrivateKey:   "c",
			},
		},
		"new": {
			input: schema.SiteConfiguration{
				AuthSaml: &schema.SAMLAuthProvider{
					IdentityProviderMetadataURL: "a",
					ServiceProviderCertificate:  "b",
					ServiceProviderPrivateKey:   "c",
				},
			},
			want: &schema.SAMLAuthProvider{
				IdentityProviderMetadataURL: "a",
				ServiceProviderCertificate:  "b",
				ServiceProviderPrivateKey:   "c",
			},
		},
		"both": {
			input: schema.SiteConfiguration{
				SamlIDProviderMetadataURL: "a",
				SamlSPCert:                "b",
				SamlSPKey:                 "c",
				AuthSaml: &schema.SAMLAuthProvider{
					IdentityProviderMetadataURL: "a2",
					ServiceProviderCertificate:  "b2",
					ServiceProviderPrivateKey:   "c2",
				},
			},
			want: &schema.SAMLAuthProvider{
				IdentityProviderMetadataURL: "a2",
				ServiceProviderCertificate:  "b2",
				ServiceProviderPrivateKey:   "c2",
			},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			got := authSAML(test.input)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}
