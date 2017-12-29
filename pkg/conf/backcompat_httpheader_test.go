package conf

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

func TestAuthHTTPHeader(t *testing.T) {
	tests := map[string]struct {
		input schema.SiteConfiguration
		want  string
	}{
		"none": {
			input: schema.SiteConfiguration{},
			want:  "",
		},
		"old": {
			input: schema.SiteConfiguration{SsoUserHeader: "a"},
			want:  "a",
		},
		"new": {
			input: schema.SiteConfiguration{AuthUserIdentityHTTPHeader: "a"},
			want:  "a",
		},
		"both": {
			input: schema.SiteConfiguration{
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
