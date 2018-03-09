package conf

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

func TestAuthHTTPHeader(t *testing.T) {
	tests := map[string]struct {
		input *schema.SiteConfiguration
		want  string
	}{
		"provider not set": {
			input: &schema.SiteConfiguration{AuthUserIdentityHTTPHeader: "a"},
			want:  "",
		},
		"none": {
			input: &schema.SiteConfiguration{AuthProvider: "http-header"},
			want:  "",
		},
		"old": {
			input: &schema.SiteConfiguration{AuthProvider: "http-header", SsoUserHeader: "a"},
			want:  "a",
		},
		"new": {
			input: &schema.SiteConfiguration{AuthProvider: "http-header", AuthUserIdentityHTTPHeader: "a"},
			want:  "a",
		},
		"both": {
			input: &schema.SiteConfiguration{
				AuthProvider:               "http-header",
				SsoUserHeader:              "a",
				AuthUserIdentityHTTPHeader: "a2",
			},
			want: "a2",
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
