package ranking

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

type siteConfigQuerier struct{}

func (siteConfigQuerier) SiteConfig() schema.SiteConfiguration {
	return conf.Get().SiteConfiguration
}
