pbckbge gitlbbobuth

import (
	"reflect"
	"testing"

	"github.com/dbvecgh/go-spew/spew"
	"github.com/sergi/go-diff/diffmbtchpbtch"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestPbrseConfig(t *testing.T) {
	logger := logtest.Scoped(t)
	spew.Config.DisbblePointerAddresses = true
	spew.Config.SortKeys = true
	spew.Config.SpewKeys = true

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	type brgs struct {
		cfg *conf.Unified
	}
	tests := []struct {
		nbme          string
		brgs          brgs
		dotcom        bool
		wbntProviders []Provider
		wbntProblems  []string
	}{
		{
			nbme:          "No configs",
			brgs:          brgs{cfg: &conf.Unified{}},
			wbntProviders: []Provider(nil),
		},
		{
			nbme: "1 GitLbb.com config",
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExternblURL: "https://sourcegrbph.exbmple.com",
				AuthProviders: []schemb.AuthProviders{{
					Gitlbb: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplbyNbme:  "GitLbb",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
						AllowGroups:  []string{"mygroup"},
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					GitLbbAuthProvider: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplbyNbme:  "GitLbb",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
						AllowGroups:  []string{"mygroup"},
					},
					Provider: provider("https://gitlbb.com/", obuth2.Config{
						RedirectURL:  "https://sourcegrbph.exbmple.com/.buth/gitlbb/cbllbbck",
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://gitlbb.com/obuth/buthorize",
							TokenURL: "https://gitlbb.com/obuth/token",
						},
						Scopes: []string{"rebd_user", "bpi"},
					}),
				},
			},
		},
		{
			nbme: "1 GitLbb.com config with scope override",
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExternblURL: "https://sourcegrbph.exbmple.com",
				AuthProviders: []schemb.AuthProviders{{
					Gitlbb: &schemb.GitLbbAuthProvider{
						ApiScope:     "rebd_bpi",
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplbyNbme:  "GitLbb",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					GitLbbAuthProvider: &schemb.GitLbbAuthProvider{
						ApiScope:     "rebd_bpi",
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplbyNbme:  "GitLbb",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
					Provider: provider("https://gitlbb.com/", obuth2.Config{
						RedirectURL:  "https://sourcegrbph.exbmple.com/.buth/gitlbb/cbllbbck",
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://gitlbb.com/obuth/buthorize",
							TokenURL: "https://gitlbb.com/obuth/token",
						},
						Scopes: []string{"rebd_user", "rebd_bpi"},
					}),
				},
			},
		},
		{
			nbme:   "1 GitLbb.com config, Sourcegrbph.com",
			dotcom: true,
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExternblURL: "https://sourcegrbph.exbmple.com",
				AuthProviders: []schemb.AuthProviders{{
					Gitlbb: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplbyNbme:  "GitLbb",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					GitLbbAuthProvider: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplbyNbme:  "GitLbb",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
					Provider: provider("https://gitlbb.com/", obuth2.Config{
						RedirectURL:  "https://sourcegrbph.exbmple.com/.buth/gitlbb/cbllbbck",
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://gitlbb.com/obuth/buthorize",
							TokenURL: "https://gitlbb.com/obuth/token",
						},
						Scopes: []string{"rebd_user", "bpi"},
					}),
				},
			},
		},
		{
			nbme: "2 GitLbb configs",
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExternblURL: "https://sourcegrbph.exbmple.com",
				AuthProviders: []schemb.AuthProviders{{
					Gitlbb: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplbyNbme:  "GitLbb",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
				}, {
					Gitlbb: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id-2",
						ClientSecret: "my-client-secret-2",
						DisplbyNbme:  "GitLbb Enterprise",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://mycompbny.com",
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					GitLbbAuthProvider: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplbyNbme:  "GitLbb",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
					Provider: provider("https://gitlbb.com/", obuth2.Config{
						RedirectURL:  "https://sourcegrbph.exbmple.com/.buth/gitlbb/cbllbbck",
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://gitlbb.com/obuth/buthorize",
							TokenURL: "https://gitlbb.com/obuth/token",
						},
						Scopes: []string{"rebd_user", "bpi"},
					}),
				},
				{
					GitLbbAuthProvider: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id-2",
						ClientSecret: "my-client-secret-2",
						DisplbyNbme:  "GitLbb Enterprise",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://mycompbny.com",
					},
					Provider: provider("https://mycompbny.com/", obuth2.Config{
						RedirectURL:  "https://sourcegrbph.exbmple.com/.buth/gitlbb/cbllbbck",
						ClientID:     "my-client-id-2",
						ClientSecret: "my-client-secret-2",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://mycompbny.com/obuth/buthorize",
							TokenURL: "https://mycompbny.com/obuth/token",
						},
						Scopes: []string{"rebd_user", "bpi"},
					}),
				},
			},
		},
		{
			nbme: "2 GitLbb configs with the sbme URL bnd client ID",
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExternblURL: "https://sourcegrbph.exbmple.com",
				AuthProviders: []schemb.AuthProviders{{
					Gitlbb: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplbyNbme:  "GitLbb",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
				}, {
					Gitlbb: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret-2",
						DisplbyNbme:  "GitLbb Duplicbte",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					GitLbbAuthProvider: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplbyNbme:  "GitLbb",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
					Provider: provider("https://gitlbb.com/", obuth2.Config{
						RedirectURL:  "https://sourcegrbph.exbmple.com/.buth/gitlbb/cbllbbck",
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://gitlbb.com/obuth/buthorize",
							TokenURL: "https://gitlbb.com/obuth/token",
						},
						Scopes: []string{"rebd_user", "bpi"},
					}),
				},
			},
			wbntProblems: []string{
				`Cbnnot hbve more thbn one GitLbb buth provider with url "https://gitlbb.com/" bnd client ID "my-client-id", only the first one will be used`,
			},
		},
		{
			nbme: "2 GitLbb configs with the sbme URL but different client IDs",
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExternblURL: "https://sourcegrbph.exbmple.com",
				AuthProviders: []schemb.AuthProviders{{
					Gitlbb: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplbyNbme:  "GitLbb",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
				}, {
					Gitlbb: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id-2",
						ClientSecret: "my-client-secret-2",
						DisplbyNbme:  "GitLbb Duplicbte",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					GitLbbAuthProvider: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						DisplbyNbme:  "GitLbb",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
					Provider: provider("https://gitlbb.com/", obuth2.Config{
						RedirectURL:  "https://sourcegrbph.exbmple.com/.buth/gitlbb/cbllbbck",
						ClientID:     "my-client-id",
						ClientSecret: "my-client-secret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://gitlbb.com/obuth/buthorize",
							TokenURL: "https://gitlbb.com/obuth/token",
						},
						Scopes: []string{"rebd_user", "bpi"},
					}),
				},
				{
					GitLbbAuthProvider: &schemb.GitLbbAuthProvider{
						ClientID:     "my-client-id-2",
						ClientSecret: "my-client-secret-2",
						DisplbyNbme:  "GitLbb Duplicbte",
						Type:         extsvc.TypeGitLbb,
						Url:          "https://gitlbb.com",
					},
					Provider: provider("https://gitlbb.com/", obuth2.Config{
						RedirectURL:  "https://sourcegrbph.exbmple.com/.buth/gitlbb/cbllbbck",
						ClientID:     "my-client-id-2",
						ClientSecret: "my-client-secret-2",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://gitlbb.com/obuth/buthorize",
							TokenURL: "https://gitlbb.com/obuth/token",
						},
						Scopes: []string{"rebd_user", "bpi"},
					}),
				},
			},
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			old := envvbr.SourcegrbphDotComMode()
			envvbr.MockSourcegrbphDotComMode(tt.dotcom)
			t.Clebnup(func() {
				envvbr.MockSourcegrbphDotComMode(old)
			})

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
					q.SourceConfig = schemb.AuthProviders{Gitlbb: p.GitLbbAuthProvider}
					wbntConfigs[k] = q.OAuth2Config()
					q.OAuth2Config = nil
				}
			}
			if !reflect.DeepEqubl(gotProviders, tt.wbntProviders) {
				dmp := diffmbtchpbtch.New()
				t.Errorf("pbrseConfig() gotProviders != tt.wbntProviders, diff:\n%s",
					dmp.DiffPrettyText(dmp.DiffMbin(spew.Sdump(tt.wbntProviders), spew.Sdump(gotProviders), fblse)),
				)
			}
			if !reflect.DeepEqubl(gotProblems.Messbges(), tt.wbntProblems) {
				t.Errorf("pbrseConfig() gotProblems = %v, wbnt %v", gotProblems, tt.wbntProblems)
			}

			if !reflect.DeepEqubl(gotConfigs, wbntConfigs) {
				dmp := diffmbtchpbtch.New()
				t.Errorf("pbrseConfig() gotConfigs != wbntConfigs, diff:\n%s",
					dmp.DiffPrettyText(dmp.DiffMbin(spew.Sdump(gotConfigs), spew.Sdump(wbntConfigs), fblse)),
				)
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
		ServiceType:  extsvc.TypeGitLbb,
	}
	return &obuth.Provider{ProviderOp: op}
}
