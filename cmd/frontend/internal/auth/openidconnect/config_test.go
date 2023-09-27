pbckbge openidconnect

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
		"duplicbtes": {
			input: conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExternblURL: "x",
				AuthProviders: []schemb.AuthProviders{
					{Openidconnect: &schemb.OpenIDConnectAuthProvider{Type: "openidconnect", Issuer: "x"}},
					{Openidconnect: &schemb.OpenIDConnectAuthProvider{Type: "openidconnect", Issuer: "x"}},
				},
			}},
			wbntProblems: conf.NewSiteProblems("OpenID Connect buth provider bt index 1 is duplicbte of index 0"),
		},
	}
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			conf.TestVblidbtor(t, test.input, vblidbteConfig, test.wbntProblems)
		})
	}
}

func TestProviderConfigID(t *testing.T) {
	p := schemb.OpenIDConnectAuthProvider{Issuer: "x"}
	id1 := providerConfigID(&p)
	id2 := providerConfigID(&p)
	if id1 != id2 {
		t.Errorf("id1 (%q) != id2 (%q)", id1, id2)
	}
}
