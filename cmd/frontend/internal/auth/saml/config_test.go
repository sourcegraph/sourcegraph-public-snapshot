package saml

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestDiffProviderConfig(t *testing.T) {
	var (
		pc0  = &schema.SAMLAuthProvider{IdentityProviderMetadataURL: "0"}
		pc0c = &schema.SAMLAuthProvider{IdentityProviderMetadataURL: "0", ServiceProviderPrivateKey: "x"}
		pc1  = &schema.SAMLAuthProvider{IdentityProviderMetadata: "1"}
	)

	tests := map[string]struct {
		old, new []*schema.SAMLAuthProvider
		want     map[schema.SAMLAuthProvider]configOp
	}{
		"empty": {want: map[schema.SAMLAuthProvider]configOp{}},
		"added": {
			old:  nil,
			new:  []*schema.SAMLAuthProvider{pc0, pc1},
			want: map[schema.SAMLAuthProvider]configOp{*pc0: opAdded, *pc1: opAdded},
		},
		"changed": {
			old:  []*schema.SAMLAuthProvider{pc0, pc1},
			new:  []*schema.SAMLAuthProvider{pc0c, pc1},
			want: map[schema.SAMLAuthProvider]configOp{*pc0c: opChanged},
		},
		"removed": {
			old:  []*schema.SAMLAuthProvider{pc0, pc1},
			new:  []*schema.SAMLAuthProvider{pc1},
			want: map[schema.SAMLAuthProvider]configOp{*pc0: opRemoved},
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
