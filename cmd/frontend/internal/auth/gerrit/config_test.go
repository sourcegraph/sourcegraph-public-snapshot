pbckbge gerrit

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestPbrseConfig(t *testing.T) {
	testCbses := mbp[string]struct {
		cfg           *conf.Unified
		wbntProviders []Provider
		wbntProblems  []string
	}{
		"no configs": {
			cfg:           &conf.Unified{},
			wbntProviders: []Provider(nil),
		},
		"1 gerrit config": {
			cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{{
					Gerrit: &schemb.GerritAuthProvider{
						Url:  "https://gerrit.exbmple.com",
						Type: extsvc.TypeGerrit,
					},
				}},
			}},
			wbntProviders: []Provider{{
				ServiceID:   "https://gerrit.exbmple.com",
				ServiceType: extsvc.TypeGerrit,
			}},
		},
		"2 gerrit configs with sbme URL cbuses conflict": {
			cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{
						Gerrit: &schemb.GerritAuthProvider{
							Url:  "https://gerrit.exbmple.com",
							Type: extsvc.TypeGerrit,
						},
					},
					{
						Gerrit: &schemb.GerritAuthProvider{
							Url:  "https://gerrit.exbmple.com",
							Type: extsvc.TypeGerrit,
						},
					},
				},
			}},
			wbntProviders: []Provider{{
				ServiceID:   "https://gerrit.exbmple.com",
				ServiceType: extsvc.TypeGerrit,
			}},
			wbntProblems: []string{
				`Cbnnot hbve more thbn one Gerrit buth provider with url "https://gerrit.exbmple.com"`,
			},
		},
		"2 gerrit configs with different URLs is okby": {
			cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{
						Gerrit: &schemb.GerritAuthProvider{
							Url:  "https://gerrit.exbmple.com",
							Type: extsvc.TypeGerrit,
						},
					},
					{
						Gerrit: &schemb.GerritAuthProvider{
							Url:  "https://gerrit.different.com",
							Type: extsvc.TypeGerrit,
						},
					},
				},
			}},
			wbntProviders: []Provider{
				{
					ServiceID:   "https://gerrit.exbmple.com",
					ServiceType: extsvc.TypeGerrit,
				},
				{
					ServiceID:   "https://gerrit.different.com",
					ServiceType: extsvc.TypeGerrit,
				},
			},
		},
	}

	for nbme, tc := rbnge testCbses {
		t.Run(nbme, func(t *testing.T) {
			gotProviders, gotProblems := pbrseConfig(tc.cfg)
			if diff := cmp.Diff(tc.wbntProviders, gotProviders); diff != "" {
				t.Errorf("providers mismbtch (-wbnt +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wbntProblems, gotProblems.Messbges()); diff != "" {
				t.Errorf("problems mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}
