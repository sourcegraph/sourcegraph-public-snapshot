package conf

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAuthHTTPHeader(t *testing.T) {
	tests := map[string]struct {
		input *schema.SiteConfiguration
		want  *schema.HTTPHeaderAuthProvider
	}{
		"provider not set": {
			input: &schema.SiteConfiguration{AuthUserIdentityHTTPHeader: "a"},
			want:  nil,
		},
		"none": {
			// This config would produce a runtime error in the auth middleware.
			input: &schema.SiteConfiguration{AuthProvider: "http-header"},
			want:  &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: ""},
		},
		"auth.provider": {
			input: &schema.SiteConfiguration{AuthProvider: "http-header", AuthUserIdentityHTTPHeader: "a"},
			want:  &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: "a"},
		},
		"auth.provider only, no header": {
			input: &schema.SiteConfiguration{AuthProvider: "http-header", AuthUserIdentityHTTPHeader: ""},
			want:  &schema.HTTPHeaderAuthProvider{Type: "http-header", UsernameHeader: ""},
		},
		"auth.userIdentityHTTPHeader only": {
			input: &schema.SiteConfiguration{AuthUserIdentityHTTPHeader: "a"},
			want:  nil,
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			got := authHTTPHeader(test.input)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}
