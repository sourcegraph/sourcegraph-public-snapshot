package conf

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
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
		"auth.provider": {
			input: &schema.SiteConfiguration{AuthProvider: "http-header", AuthUserIdentityHTTPHeader: "a"},
			want:  "a",
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
