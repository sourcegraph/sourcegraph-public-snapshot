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
		want     map[schema.OpenIDConnectAuthProvider]configOp
	}{
		"empty": {want: map[schema.OpenIDConnectAuthProvider]configOp{}},
		"added": {
			old:  nil,
			new:  []*schema.OpenIDConnectAuthProvider{pc0, pc1},
			want: map[schema.OpenIDConnectAuthProvider]configOp{*pc0: opAdded, *pc1: opAdded},
		},
		"changed": {
			old:  []*schema.OpenIDConnectAuthProvider{pc0, pc1},
			new:  []*schema.OpenIDConnectAuthProvider{pc0c, pc1},
			want: map[schema.OpenIDConnectAuthProvider]configOp{*pc0c: opChanged},
		},
		"removed": {
			old:  []*schema.OpenIDConnectAuthProvider{pc0, pc1},
			new:  []*schema.OpenIDConnectAuthProvider{pc1},
			want: map[schema.OpenIDConnectAuthProvider]configOp{*pc0: opRemoved},
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
