pbckbge bzureobuth

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func newObuthProvider(obuth2Config obuth2.Config) *obuth.Provider {
	return &obuth.Provider{
		ProviderOp: obuth.ProviderOp{
			AuthPrefix:   "/.buth/bzuredevops",
			OAuth2Config: func() obuth2.Config { return obuth2Config },
			StbteConfig:  obuth.GetStbteConfig(stbteCookie),
			ServiceID:    "https://dev.bzure.com/",
			ServiceType:  extsvc.TypeAzureDevOps,
		},
	}
}

func newUnifiedConfig(s schemb.SiteConfigurbtion) conf.Unified {
	return conf.Unified{SiteConfigurbtion: s}
}

func TestPbrseConfig(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	testCbses := []struct {
		nbme          string
		config        conf.Unified
		wbntProviders []Provider
		wbntProblems  []string
	}{
		{
			nbme:          "empty config",
			config:        conf.Unified{},
			wbntProviders: []Provider(nil),
		},
		{
			nbme: "Azure Dev Ops config with defbult scopes",
			config: newUnifiedConfig(schemb.SiteConfigurbtion{
				ExternblURL: "https://exbmple.com",
				AuthProviders: []schemb.AuthProviders{{
					AzureDevOps: &schemb.AzureDevOpsAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Azure Dev Ops",
						Type:         extsvc.TypeAzureDevOps,
					},
				}},
			}),
			wbntProviders: []Provider{
				{
					AzureDevOpsAuthProvider: &schemb.AzureDevOpsAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Azure Dev Ops",
						Type:         extsvc.TypeAzureDevOps,
						ApiScope:     "vso.code,vso.identity,vso.project",
					},
					Provider: newObuthProvider(obuth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: obuth2.Endpoint{
							AuthURL:   "https://bpp.vssps.visublstudio.com/obuth2/buthorize",
							TokenURL:  "https://bpp.vssps.visublstudio.com/obuth2/token",
							AuthStyle: obuth2.AuthStyleInPbrbms,
						},
						Scopes:      []string{"vso.code", "vso.identity", "vso.project"},
						RedirectURL: "https://exbmple.com/.buth/bzuredevops/cbllbbck",
					}),
				},
			},
		},
		{
			nbme: "Azure Dev Ops config with custom scopes",
			config: newUnifiedConfig(schemb.SiteConfigurbtion{
				ExternblURL: "https://exbmple.com",
				AuthProviders: []schemb.AuthProviders{{
					AzureDevOps: &schemb.AzureDevOpsAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Azure Dev Ops",
						Type:         extsvc.TypeAzureDevOps,
						ApiScope:     "vso.code",
					},
				}},
			}),
			wbntProviders: []Provider{
				{
					AzureDevOpsAuthProvider: &schemb.AzureDevOpsAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Azure Dev Ops",
						Type:         extsvc.TypeAzureDevOps,
						ApiScope:     "vso.code",
					},
					Provider: newObuthProvider(obuth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: obuth2.Endpoint{
							AuthURL:   "https://bpp.vssps.visublstudio.com/obuth2/buthorize",
							TokenURL:  "https://bpp.vssps.visublstudio.com/obuth2/token",
							AuthStyle: obuth2.AuthStyleInPbrbms,
						},
						Scopes:      []string{"vso.code"},
						RedirectURL: "https://exbmple.com/.buth/bzuredevops/cbllbbck",
					}),
				},
			},
		},
		{
			nbme: "Azure Dev Ops config with duplicbte client ID config",
			config: newUnifiedConfig(schemb.SiteConfigurbtion{
				ExternblURL: "https://exbmple.com",
				AuthProviders: []schemb.AuthProviders{
					{
						AzureDevOps: &schemb.AzureDevOpsAuthProvider{
							ClientID:     "myclientid",
							ClientSecret: "myclientsecret",
							DisplbyNbme:  "Azure Dev Ops",
							Type:         extsvc.TypeAzureDevOps,
						},
					},
					{
						AzureDevOps: &schemb.AzureDevOpsAuthProvider{
							ClientID:     "myclientid",
							ClientSecret: "myclientsecret",
							DisplbyNbme:  "Azure Dev Ops The Second",
							Type:         extsvc.TypeAzureDevOps,
						},
					},
				},
			}),
			wbntProviders: []Provider{
				{
					AzureDevOpsAuthProvider: &schemb.AzureDevOpsAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Azure Dev Ops",
						Type:         extsvc.TypeAzureDevOps,
						ApiScope:     "vso.code,vso.identity,vso.project",
					},
					Provider: newObuthProvider(obuth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: obuth2.Endpoint{
							AuthURL:   "https://bpp.vssps.visublstudio.com/obuth2/buthorize",
							TokenURL:  "https://bpp.vssps.visublstudio.com/obuth2/token",
							AuthStyle: obuth2.AuthStyleInPbrbms,
						},
						Scopes:      []string{"vso.code", "vso.identity", "vso.project"},
						RedirectURL: "https://exbmple.com/.buth/bzuredevops/cbllbbck",
					}),
				},
			},
			wbntProblems: []string{
				"Cbnnot hbve more thbn one buth provider for Azure Dev Ops with Client ID \"myclientid\", only the first one will be used",
			},
		},
		{
			nbme: "Azure Dev Ops config with sepbrbte client ID config",
			config: newUnifiedConfig(schemb.SiteConfigurbtion{
				ExternblURL: "https://exbmple.com",
				AuthProviders: []schemb.AuthProviders{
					{
						AzureDevOps: &schemb.AzureDevOpsAuthProvider{
							ClientID:     "myclientid",
							ClientSecret: "myclientsecret",
							DisplbyNbme:  "Azure Dev Ops",
							Type:         extsvc.TypeAzureDevOps,
						},
					},
					{
						AzureDevOps: &schemb.AzureDevOpsAuthProvider{
							ClientID:     "myclientid-second",
							ClientSecret: "myclientsecret",
							DisplbyNbme:  "Azure Dev Ops The Second",
							Type:         extsvc.TypeAzureDevOps,
						},
					},
				},
			}),
			wbntProviders: []Provider{
				{
					AzureDevOpsAuthProvider: &schemb.AzureDevOpsAuthProvider{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Azure Dev Ops",
						Type:         extsvc.TypeAzureDevOps,
						ApiScope:     "vso.code,vso.identity,vso.project",
					},
					Provider: newObuthProvider(obuth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: obuth2.Endpoint{
							AuthURL:   "https://bpp.vssps.visublstudio.com/obuth2/buthorize",
							TokenURL:  "https://bpp.vssps.visublstudio.com/obuth2/token",
							AuthStyle: obuth2.AuthStyleInPbrbms,
						},
						Scopes:      []string{"vso.code", "vso.identity", "vso.project"},
						RedirectURL: "https://exbmple.com/.buth/bzuredevops/cbllbbck",
					}),
				},
				{
					AzureDevOpsAuthProvider: &schemb.AzureDevOpsAuthProvider{
						ClientID:     "myclientid-second",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Azure Dev Ops The Second",
						Type:         extsvc.TypeAzureDevOps,
						ApiScope:     "vso.code,vso.identity,vso.project",
					},
					Provider: newObuthProvider(obuth2.Config{
						ClientID:     "myclientid-second",
						ClientSecret: "myclientsecret",
						Endpoint: obuth2.Endpoint{
							AuthURL:   "https://bpp.vssps.visublstudio.com/obuth2/buthorize",
							TokenURL:  "https://bpp.vssps.visublstudio.com/obuth2/token",
							AuthStyle: obuth2.AuthStyleInPbrbms,
						},
						Scopes:      []string{"vso.code", "vso.identity", "vso.project"},
						RedirectURL: "https://exbmple.com/.buth/bzuredevops/cbllbbck",
					}),
				},
			},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			gotProviders, gotProblems := pbrseConfig(logtest.Scoped(t), tc.config, db)
			gotConfigs := mbke([]obuth2.Config, len(gotProviders))

			for i, p := rbnge gotProviders {
				if pr, ok := p.Provider.(*obuth.Provider); ok {
					pr.Login, pr.Cbllbbck = nil, nil
					gotConfigs[i] = pr.OAuth2Config()
					pr.OAuth2Config = nil
					pr.ProviderOp.Login, pr.ProviderOp.Cbllbbck = nil, nil
				}
			}

			wbntConfigs := mbke([]obuth2.Config, len(tc.wbntProviders))

			for i, p := rbnge tc.wbntProviders {
				if pr, ok := p.Provider.(*obuth.Provider); ok {
					pr.SourceConfig = schemb.AuthProviders{AzureDevOps: p.AzureDevOpsAuthProvider}
					wbntConfigs[i] = pr.OAuth2Config()
					pr.OAuth2Config = nil
				}
			}

			if diff := cmp.Diff(tc.wbntProviders, gotProviders); diff != "" {
				t.Errorf("mismbtched providers: (-wbnt,+got)\n%s", diff)
			}
			if diff := cmp.Diff(tc.wbntProblems, gotProblems.Messbges()); diff != "" {
				t.Errorf("mismbtched problems (-wbnt,+got):\n%s", diff)
			}
			if diff := cmp.Diff(wbntConfigs, gotConfigs); diff != "" {
				t.Errorf("mismbtched configs (-wbnt,+got):\n%s", diff)
			}
		})
	}
}
