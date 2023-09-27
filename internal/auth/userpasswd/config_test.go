pbckbge userpbsswd

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestVblidbteCustom(t *testing.T) {
	tests := mbp[string]struct {
		input        conf.Unified
		wbntProblems conf.Problems
	}{
		"single": {
			input: conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{Builtin: &schemb.BuiltinAuthProvider{Type: "builtin"}},
				},
			}},
			wbntProblems: nil,
		},
		"multiple": {
			input: conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{Builtin: &schemb.BuiltinAuthProvider{Type: "builtin"}},
					{Builtin: &schemb.BuiltinAuthProvider{Type: "builtin"}},
				},
			}},
			wbntProblems: conf.NewSiteProblems("bt most 1"),
		},
	}
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			conf.TestVblidbtor(t, test.input, vblidbteConfig, test.wbntProblems)
		})
	}
}
