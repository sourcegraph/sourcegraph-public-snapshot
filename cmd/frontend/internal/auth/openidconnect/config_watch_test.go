package openidconnect

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestDiffProviderConfig(t *testing.T) {
	var (
		pc0  = &schema.OpenIDConnectAuthProvider{Issuer: "0"}
		pc0c = &schema.OpenIDConnectAuthProvider{Issuer: "0", ClientSecret: "x"}
		pc1  = &schema.OpenIDConnectAuthProvider{Issuer: "1"}
	)

	tests := map[string]struct {
		old, new []*schema.OpenIDConnectAuthProvider
		want     map[schema.OpenIDConnectAuthProvider]bool
	}{
		"empty": {want: map[schema.OpenIDConnectAuthProvider]bool{}},
		"added": {
			old:  nil,
			new:  []*schema.OpenIDConnectAuthProvider{pc0, pc1},
			want: map[schema.OpenIDConnectAuthProvider]bool{*pc0: true, *pc1: true},
		},
		"changed": {
			old:  []*schema.OpenIDConnectAuthProvider{pc0, pc1},
			new:  []*schema.OpenIDConnectAuthProvider{pc0c, pc1},
			want: map[schema.OpenIDConnectAuthProvider]bool{*pc0: false, *pc0c: true},
		},
		"removed": {
			old:  []*schema.OpenIDConnectAuthProvider{pc0, pc1},
			new:  []*schema.OpenIDConnectAuthProvider{pc1},
			want: map[schema.OpenIDConnectAuthProvider]bool{*pc0: false},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			diff := diffProviderConfig(test.old, test.new)
			if !reflect.DeepEqual(diff, test.want) {
				t.Errorf("got != want\n got %+v\nwant %+v", diff, test.want)
			}
		})
	}
}
