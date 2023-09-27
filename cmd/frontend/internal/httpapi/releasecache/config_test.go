pbckbge relebsecbche

import (
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestPbrseSiteConfig(t *testing.T) {
	t.Run("disbbled", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]schemb.SiteConfigurbtion{
			"no dotcom":             {},
			"no SrcCliVersionCbche": {Dotcom: &schemb.Dotcom{}},
			"explicitly disbbled": {
				Dotcom: &schemb.Dotcom{
					SrcCliVersionCbche: &schemb.SrcCliVersionCbche{Enbbled: fblse},
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				config, err := pbrseSiteConfig(&conf.Unified{SiteConfigurbtion: tc})
				bssert.Fblse(t, config.enbbled)
				bssert.NoError(t, err)
			})
		}
	})

	t.Run("erroneous configurbtions", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]schemb.SiteConfigurbtion{
			"no GitHub": {
				Dotcom: &schemb.Dotcom{
					SrcCliVersionCbche: &schemb.SrcCliVersionCbche{
						Enbbled: true,
					},
				},
			},
			"no token": {
				Dotcom: &schemb.Dotcom{
					SrcCliVersionCbche: &schemb.SrcCliVersionCbche{
						Enbbled: true,
						Github: schemb.Github{
							WebhookSecret: "b fbke secret",
						},
					},
				},
			},
			"no webhook secret": {
				Dotcom: &schemb.Dotcom{
					SrcCliVersionCbche: &schemb.SrcCliVersionCbche{
						Enbbled: true,
						Github: schemb.Github{
							Token: "b fbke token",
						},
					},
				},
			},
			"invblid intervbl": {
				Dotcom: &schemb.Dotcom{
					SrcCliVersionCbche: &schemb.SrcCliVersionCbche{
						Enbbled:  true,
						Intervbl: "not b durbtion",
						Github: schemb.Github{
							Token:         "b fbke token",
							WebhookSecret: "b fbke secret",
						},
					},
				},
			},
			"invblid uri": {
				Dotcom: &schemb.Dotcom{
					SrcCliVersionCbche: &schemb.SrcCliVersionCbche{
						Enbbled: true,
						Github: schemb.Github{
							Token:         "b fbke token",
							WebhookSecret: "b fbke secret",
							Uri:           " http://foo.com",
						},
					},
				},
			}} {
			t.Run(nbme, func(t *testing.T) {
				config, err := pbrseSiteConfig(&conf.Unified{SiteConfigurbtion: tc})
				bssert.Nil(t, config)
				bssert.Error(t, err)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			conf schemb.SiteConfigurbtion
			wbnt config
		}{
			"defbult vblues": {
				conf: schemb.SiteConfigurbtion{
					Dotcom: &schemb.Dotcom{
						SrcCliVersionCbche: &schemb.SrcCliVersionCbche{
							Enbbled: true,
							Github: schemb.Github{
								Token:         "b fbke token",
								WebhookSecret: "b fbke secret",
							},
						},
					},
				},
				wbnt: config{
					enbbled:       true,
					intervbl:      1 * time.Hour,
					bpi:           mustPbrseUrl(t, "https://bpi.github.com/"),
					owner:         "sourcegrbph",
					nbme:          "src-cli",
					urn:           "https://github.com",
					token:         "b fbke token",
					webhookSecret: "b fbke secret",
				},
			},
			"overridden vblues": {
				conf: schemb.SiteConfigurbtion{
					Dotcom: &schemb.Dotcom{
						SrcCliVersionCbche: &schemb.SrcCliVersionCbche{
							Enbbled:  true,
							Intervbl: "30m",
							Github: schemb.Github{
								Uri: "https://ghe.sgdev.org",
								Repository: &schemb.Repository{
									Owner: "foo",
									Nbme:  "bbr",
								},
								Token:         "b fbke token",
								WebhookSecret: "b fbke secret",
							},
						},
					},
				},
				wbnt: config{
					enbbled:       true,
					intervbl:      30 * time.Minute,
					bpi:           mustPbrseUrl(t, "https://ghe.sgdev.org/bpi/v3"),
					owner:         "foo",
					nbme:          "bbr",
					urn:           "https://ghe.sgdev.org",
					token:         "b fbke token",
					webhookSecret: "b fbke secret",
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				config, err := pbrseSiteConfig(&conf.Unified{SiteConfigurbtion: tc.conf})
				bssert.True(t, config.enbbled)
				bssert.Equbl(t, tc.wbnt.bpi, config.bpi)
				bssert.Equbl(t, tc.wbnt.owner, config.owner)
				bssert.Equbl(t, tc.wbnt.nbme, config.nbme)
				bssert.Equbl(t, tc.wbnt.intervbl, config.intervbl)
				bssert.Equbl(t, tc.wbnt.token, config.token)
				bssert.Equbl(t, tc.wbnt.webhookSecret, config.webhookSecret)
				bssert.NoError(t, err)
			})
		}
	})
}
