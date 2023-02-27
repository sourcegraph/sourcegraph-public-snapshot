package releasecache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestParseSiteConfig(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		for name, tc := range map[string]schema.SiteConfiguration{
			"no dotcom":             {},
			"no SrcCliVersionCache": {Dotcom: &schema.Dotcom{}},
			"explicitly disabled": {
				Dotcom: &schema.Dotcom{
					SrcCliVersionCache: &schema.SrcCliVersionCache{Enabled: false},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				config, err := parseSiteConfig(&conf.Unified{SiteConfiguration: tc})
				assert.False(t, config.enabled)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("erroneous configurations", func(t *testing.T) {
		for name, tc := range map[string]schema.SiteConfiguration{
			"no GitHub": {
				Dotcom: &schema.Dotcom{
					SrcCliVersionCache: &schema.SrcCliVersionCache{
						Enabled: true,
					},
				},
			},
			"no token": {
				Dotcom: &schema.Dotcom{
					SrcCliVersionCache: &schema.SrcCliVersionCache{
						Enabled: true,
						Github: schema.Github{
							WebhookSecret: "a fake secret",
						},
					},
				},
			},
			"no webhook secret": {
				Dotcom: &schema.Dotcom{
					SrcCliVersionCache: &schema.SrcCliVersionCache{
						Enabled: true,
						Github: schema.Github{
							Token: "a fake token",
						},
					},
				},
			},
			"invalid interval": {
				Dotcom: &schema.Dotcom{
					SrcCliVersionCache: &schema.SrcCliVersionCache{
						Enabled:  true,
						Interval: "not a duration",
						Github: schema.Github{
							Token:         "a fake token",
							WebhookSecret: "a fake secret",
						},
					},
				},
			},
			"invalid uri": {
				Dotcom: &schema.Dotcom{
					SrcCliVersionCache: &schema.SrcCliVersionCache{
						Enabled: true,
						Github: schema.Github{
							Token:         "a fake token",
							WebhookSecret: "a fake secret",
							Uri:           " http://foo.com",
						},
					},
				},
			}} {
			t.Run(name, func(t *testing.T) {
				config, err := parseSiteConfig(&conf.Unified{SiteConfiguration: tc})
				assert.Nil(t, config)
				assert.Error(t, err)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			conf schema.SiteConfiguration
			want config
		}{
			"default values": {
				conf: schema.SiteConfiguration{
					Dotcom: &schema.Dotcom{
						SrcCliVersionCache: &schema.SrcCliVersionCache{
							Enabled: true,
							Github: schema.Github{
								Token:         "a fake token",
								WebhookSecret: "a fake secret",
							},
						},
					},
				},
				want: config{
					enabled:       true,
					interval:      1 * time.Hour,
					api:           mustParseUrl(t, "https://api.github.com/"),
					owner:         "sourcegraph",
					name:          "src-cli",
					urn:           "https://github.com",
					token:         "a fake token",
					webhookSecret: "a fake secret",
				},
			},
			"overridden values": {
				conf: schema.SiteConfiguration{
					Dotcom: &schema.Dotcom{
						SrcCliVersionCache: &schema.SrcCliVersionCache{
							Enabled:  true,
							Interval: "30m",
							Github: schema.Github{
								Uri: "https://ghe.sgdev.org",
								Repository: &schema.Repository{
									Owner: "foo",
									Name:  "bar",
								},
								Token:         "a fake token",
								WebhookSecret: "a fake secret",
							},
						},
					},
				},
				want: config{
					enabled:       true,
					interval:      30 * time.Minute,
					api:           mustParseUrl(t, "https://ghe.sgdev.org/api/v3"),
					owner:         "foo",
					name:          "bar",
					urn:           "https://ghe.sgdev.org",
					token:         "a fake token",
					webhookSecret: "a fake secret",
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				config, err := parseSiteConfig(&conf.Unified{SiteConfiguration: tc.conf})
				assert.True(t, config.enabled)
				assert.Equal(t, tc.want.api, config.api)
				assert.Equal(t, tc.want.owner, config.owner)
				assert.Equal(t, tc.want.name, config.name)
				assert.Equal(t, tc.want.interval, config.interval)
				assert.Equal(t, tc.want.token, config.token)
				assert.Equal(t, tc.want.webhookSecret, config.webhookSecret)
				assert.NoError(t, err)
			})
		}
	})
}
