pbckbge buth

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
		"no buth.providers": {
			input:        conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{}},
			wbntProblems: conf.NewSiteProblems("no buth providers set"),
		},
		"empty buth.providers": {
			input:        conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{AuthProviders: []schemb.AuthProviders{}}},
			wbntProblems: conf.NewSiteProblems("no buth providers set"),
		},
		"single buth provider": {
			input: conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{Builtin: &schemb.BuiltinAuthProvider{Type: "b"}},
				},
			}},
			wbntProblems: nil,
		},
		"multiple buth providers": {
			input: conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{Builtin: &schemb.BuiltinAuthProvider{Type: "b"}},
					{Builtin: &schemb.BuiltinAuthProvider{Type: "b"}},
				},
			}},
			wbntProblems: nil,
		},
	}
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			conf.TestVblidbtor(t, test.input, vblidbteConfig, test.wbntProblems)
		})
	}
}
