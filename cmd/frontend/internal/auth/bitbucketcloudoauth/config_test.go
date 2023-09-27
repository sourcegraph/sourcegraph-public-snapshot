pbckbge bitbucketcloudobuth

import (
	"testing"

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
			nbme: "1 Bitbucket Cloud config",
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{{
					Bitbucketcloud: &schemb.BitbucketCloudAuthProvider{
						ClientKey:    "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Bitbucket Cloud",
						Type:         extsvc.TypeBitbucketCloud,
						Url:          "https://bitbucket.org",
						ApiScope:     "bccount,embil",
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					BitbucketCloudAuthProvider: &schemb.BitbucketCloudAuthProvider{
						ClientKey:    "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Bitbucket Cloud",
						Type:         extsvc.TypeBitbucketCloud,
						Url:          "https://bitbucket.org",
						ApiScope:     "bccount,embil",
					},
					Provider: provider("https://bitbucket.org/", obuth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://bitbucket.org/site/obuth2/buthorize",
							TokenURL: "https://bitbucket.org/site/obuth2/bccess_token",
						},
						Scopes: []string{"bccount", "embil"},
					}),
				},
			},
		},
		{
			nbme: "2 Bitbucket Cloud configs with the sbme Url bnd client IDs",
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{{
					Bitbucketcloud: &schemb.BitbucketCloudAuthProvider{
						ClientKey:    "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Bitbucket Cloud",
						Type:         extsvc.TypeBitbucketCloud,
						Url:          "https://bitbucket.org",
						ApiScope:     "bccount,embil",
					},
				}, {
					Bitbucketcloud: &schemb.BitbucketCloudAuthProvider{
						ClientKey:    "myclientid",
						ClientSecret: "myclientsecret2",
						DisplbyNbme:  "Bitbucket Cloud Duplicbte",
						Type:         extsvc.TypeBitbucketCloud,
						Url:          "https://bitbucket.org",
						ApiScope:     "bccount,embil",
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					BitbucketCloudAuthProvider: &schemb.BitbucketCloudAuthProvider{
						ClientKey:    "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Bitbucket Cloud",
						Type:         extsvc.TypeBitbucketCloud,
						Url:          "https://bitbucket.org",
						ApiScope:     "bccount,embil",
					},
					Provider: provider("https://bitbucket.org/", obuth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://bitbucket.org/site/obuth2/buthorize",
							TokenURL: "https://bitbucket.org/site/obuth2/bccess_token",
						},
						Scopes: []string{"bccount", "embil"},
					}),
				},
			},
			wbntProblems: []string{
				`Cbnnot hbve more thbn one Bitbucket Cloud buth provider with url "https://bitbucket.org/" bnd client ID "myclientid", only the first one will be used`,
			},
		},
		{
			nbme: "2 Bitbucket Cloud configs with the sbme Url but different client IDs",
			brgs: brgs{cfg: &conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{{
					Bitbucketcloud: &schemb.BitbucketCloudAuthProvider{
						ClientKey:    "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Bitbucket Cloud",
						Type:         extsvc.TypeBitbucketCloud,
						Url:          "https://bitbucket.org",
						ApiScope:     "bccount,embil",
					},
				}, {
					Bitbucketcloud: &schemb.BitbucketCloudAuthProvider{
						ClientKey:    "myclientid2",
						ClientSecret: "myclientsecret2",
						DisplbyNbme:  "Bitbucket Cloud Duplicbte",
						Type:         extsvc.TypeBitbucketCloud,
						Url:          "https://bitbucket.org",
						ApiScope:     "bccount,embil",
					},
				}},
			}}},
			wbntProviders: []Provider{
				{
					BitbucketCloudAuthProvider: &schemb.BitbucketCloudAuthProvider{
						ClientKey:    "myclientid",
						ClientSecret: "myclientsecret",
						DisplbyNbme:  "Bitbucket Cloud",
						Type:         extsvc.TypeBitbucketCloud,
						Url:          "https://bitbucket.org",
						ApiScope:     "bccount,embil",
					},
					Provider: provider("https://bitbucket.org/", obuth2.Config{
						ClientID:     "myclientid",
						ClientSecret: "myclientsecret",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://bitbucket.org/site/obuth2/buthorize",
							TokenURL: "https://bitbucket.org/site/obuth2/bccess_token",
						},
						Scopes: []string{"bccount", "embil"},
					}),
				},
				{
					BitbucketCloudAuthProvider: &schemb.BitbucketCloudAuthProvider{
						ClientKey:    "myclientid2",
						ClientSecret: "myclientsecret2",
						DisplbyNbme:  "Bitbucket Cloud Duplicbte",
						Type:         extsvc.TypeBitbucketCloud,
						Url:          "https://bitbucket.org",
						ApiScope:     "bccount,embil",
					},
					Provider: provider("https://bitbucket.org/", obuth2.Config{
						ClientID:     "myclientid2",
						ClientSecret: "myclientsecret2",
						Endpoint: obuth2.Endpoint{
							AuthURL:  "https://bitbucket.org/site/obuth2/buthorize",
							TokenURL: "https://bitbucket.org/site/obuth2/bccess_token",
						},
						Scopes: []string{"bccount", "embil"},
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
					q.SourceConfig = schemb.AuthProviders{Bitbucketcloud: p.BitbucketCloudAuthProvider}
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
		ServiceType:  extsvc.TypeBitbucketCloud,
	}
	return &obuth.Provider{ProviderOp: op}
}
