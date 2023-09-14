package gerrit

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestParseConfig(t *testing.T) {
	testCases := map[string]struct {
		cfg           *conf.Unified
		wantProviders []Provider
		wantProblems  []string
	}{
		"no configs": {
			cfg:           &conf.Unified{},
			wantProviders: []Provider(nil),
		},
		"1 gerrit config": {
			cfg: &conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{{
					Gerrit: &schema.GerritAuthProvider{
						Url:  "https://gerrit.example.com",
						Type: extsvc.TypeGerrit,
					},
				}},
			}},
			wantProviders: []Provider{{
				ServiceID:   "https://gerrit.example.com",
				ServiceType: extsvc.TypeGerrit,
			}},
		},
		"2 gerrit configs with same URL causes conflict": {
			cfg: &conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{
						Gerrit: &schema.GerritAuthProvider{
							Url:  "https://gerrit.example.com",
							Type: extsvc.TypeGerrit,
						},
					},
					{
						Gerrit: &schema.GerritAuthProvider{
							Url:  "https://gerrit.example.com",
							Type: extsvc.TypeGerrit,
						},
					},
				},
			}},
			wantProviders: []Provider{{
				ServiceID:   "https://gerrit.example.com",
				ServiceType: extsvc.TypeGerrit,
			}},
			wantProblems: []string{
				`Cannot have more than one Gerrit auth provider with url "https://gerrit.example.com"`,
			},
		},
		"2 gerrit configs with different URLs is okay": {
			cfg: &conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{
						Gerrit: &schema.GerritAuthProvider{
							Url:  "https://gerrit.example.com",
							Type: extsvc.TypeGerrit,
						},
					},
					{
						Gerrit: &schema.GerritAuthProvider{
							Url:  "https://gerrit.different.com",
							Type: extsvc.TypeGerrit,
						},
					},
				},
			}},
			wantProviders: []Provider{
				{
					ServiceID:   "https://gerrit.example.com",
					ServiceType: extsvc.TypeGerrit,
				},
				{
					ServiceID:   "https://gerrit.different.com",
					ServiceType: extsvc.TypeGerrit,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			gotProviders, gotProblems := parseConfig(tc.cfg)
			if diff := cmp.Diff(tc.wantProviders, gotProviders); diff != "" {
				t.Errorf("providers mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantProblems, gotProblems.Messages()); diff != "" {
				t.Errorf("problems mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
