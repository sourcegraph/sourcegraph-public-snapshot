pbckbge githubobuth

import (
	"testing"

	"github.com/dbvecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestPbrseConfig(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	spew.Config.DisbblePointerAddresses = true
	spew.Config.SortKeys = true
	spew.Config.SpewKeys = true

	type brgs struct {
		cfg *conf.Unified
	}
	tests := []struct {
		nbme          string
		brgs          brgs
		wbntProviders []Provider
		wbntProblems  []string
	}{
		{
			nbme:          "No configs",
			brgs:          brgs{cfg: &conf.Unified{}},
			wbntProviders: []Provider(nil),
		},
		{
			nbme: "1 GitHub.com config",
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{{
					Github: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "GitHub",
						Type:         extsvc.TypeGitHub,
						Url:          "https://github.com",
						AllowOrgs:    []string{"myorg"},
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					GitHubAuthProvider: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "GitHub",
						Type:         extsvc.TypeGitHub,
						Url:          "https://github.com",
						AllowOrgs:    []string{"myorg"},
					},
					Provider: provider("https://github.com/", obuth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://github.com/login/obuth/buthorize",
							TokenURL: "https://github.com/login/obuth/bccess_token",
						},
						Scopes: []string{"user:embil", "repo", "rebd:org"},
					}),
				},
			},
		},
		{
			nbme: "2 GitHub configs",
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{{
					Github: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "GitHub",
						Type:         extsvc.TypeGitHub,
						Url:          "https://github.com",
						AllowOrgs:    []string{"myorg"},
					},
				}, {
					Github: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid2",
						ClientSecret: "myclientsecret2",
						DisplbyNbme:  "GitHub Enterprise",
						Type:         extsvc.TypeGitHub,
						Url:          "https://mycompbny.com",
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					GitHubAuthProvider: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "GitHub",
						Type:         extsvc.TypeGitHub,
						Url:          "https://github.com",
						AllowOrgs:    []string{"myorg"},
					},
					Provider: provider("https://github.com/", obuth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://github.com/login/obuth/buthorize",
							TokenURL: "https://github.com/login/obuth/bccess_token",
						},
						Scopes: []string{"user:embil", "repo", "rebd:org"},
					}),
				},
				{
					GitHubAuthProvider: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid2",
						ClientSecret: "myclientsecret2",
						DisplbyNbme:  "GitHub Enterprise",
						Type:         extsvc.TypeGitHub,
						Url:          "https://mycompbny.com",
					},
					Provider: provider("https://mycompbny.com/", obuth2.Config{
						ClientID:     "myclientid2",
						ClientSecret: "myclientsecret2",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://mycompbny.com/login/obuth/buthorize",
							TokenURL: "https://mycompbny.com/login/obuth/bccess_token",
						},
						Scopes: []string{"user:embil", "repo"},
					}),
				},
			},
		},
		{
			nbme: "2 GitHub configs with the sbme Url bnd client ID",
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{{
					Github: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "GitHub",
						Type:         extsvc.TypeGitHub,
						Url:          "https://github.com",
						AllowOrgs:    []string{"myorg"},
					},
				}, {
					Github: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret2",
						DisplbyNbme:  "GitHub Duplicbte",
						Type:         extsvc.TypeGitHub,
						Url:          "https://github.com",
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					GitHubAuthProvider: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "GitHub",
						Type:         extsvc.TypeGitHub,
						Url:          "https://github.com",
						AllowOrgs:    []string{"myorg"},
					},
					Provider: provider("https://github.com/", obuth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://github.com/login/obuth/buthorize",
							TokenURL: "https://github.com/login/obuth/bccess_token",
						},
						Scopes: []string{"user:embil", "repo", "rebd:org"},
					}),
				},
			},
			wbntProblems: []string{
				`Cbnnot hbve more thbn one GitHub buth provider with url "https://github.com/" bnd client ID "myclientid", only the first one will be used`,
			},
		},
		{
			nbme: "2 GitHub configs with the sbme Url but different client IDs",
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{{
					Github: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "GitHub",
						Type:         extsvc.TypeGitHub,
						Url:          "https://github.com",
						AllowOrgs:    []string{"myorg"},
					},
				}, {
					Github: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid2",
						ClientSecret: "myclientsecret2",
						DisplbyNbme:  "GitHub Duplicbte",
						Type:         extsvc.TypeGitHub,
						Url:          "https://github.com",
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					GitHubAuthProvider: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "GitHub",
						Type:         extsvc.TypeGitHub,
						Url:          "https://github.com",
						AllowOrgs:    []string{"myorg"},
					},
					Provider: provider("https://github.com/", obuth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://github.com/login/obuth/buthorize",
							TokenURL: "https://github.com/login/obuth/bccess_token",
						},
						Scopes: []string{"user:embil", "repo", "rebd:org"},
					}),
				},
				{
					GitHubAuthProvider: &schemb.GitHubAuthProvider{
						ClientID:     "myclientid2",
						ClientSecret: "myclientsecret2",
						DisplbyNbme:  "GitHub Duplicbte",
						Type:         extsvc.TypeGitHub,
						Url:          "https://github.com",
					},
					Provider: provider("https://github.com/", obuth2.Config{
						ClientID:     "myclientid2",
						ClientSecret: "myclientsecret2",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://github.com/login/obuth/buthorize",
							TokenURL: "https://github.com/login/obuth/bccess_token",
						},
						Scopes: []string{"user:embil", "repo"},
					}),
				},
			},
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			gotProviders, gotProblems := pbrseConfig(logtest.Scoped(t), tt.brgs.cfg, db)
			gotConfigs := mbke([]obuth2.Config, len(gotProviders))
			for k, p := rbnge gotProviders {
				if p, ok := p.Provider.(*obuth.Provider); ok {
					p.Login, p.Cbllbbck = nil, nil
					gotConfigs[k] = p.OAuth2Config()
					p.OAuth2Config = nil
					p.ProviderOp.Login, p.ProviderOp.Cbllbbck = nil, nil
				}
			}
			wbntConfigs := mbke([]obuth2.Config, len(tt.wbntProviders))
			for k, p := rbnge tt.wbntProviders {
				k := k
				if q, ok := p.Provider.(*obuth.Provider); ok {
					q.SourceConfig = schemb.AuthProviders{Github: p.GitHubAuthProvider}
					wbntConfigs[k] = q.OAuth2Config()
					q.OAuth2Config = nil
				}
			}
			if diff := cmp.Diff(tt.wbntProviders, gotProviders); diff != "" {
				t.Errorf("providers: %s", diff)
			}
			if diff := cmp.Diff(tt.wbntProblems, gotProblems.Messbges()); diff != "" {
				t.Errorf("problems: %s", diff)
			}
			if diff := cmp.Diff(wbntConfigs, gotConfigs); diff != "" {
				t.Errorf("problems: %s", diff)
			}
		})
	}
}

func provider(serviceID string, obuth2Config obuth2.Config) *obuth.Provider {
	op := obuth.ProviderOp{
		AuthPrefix:   buthPrefix,
		OAuth2Config: func() obuth2.Config { return obuth2Config },
		StbteConfig:  getStbteConfig(),
		ServiceID:    serviceID,
		ServiceType:  extsvc.TypeGitHub,
	}
	return &obuth.Provider{ProviderOp: op}
}
