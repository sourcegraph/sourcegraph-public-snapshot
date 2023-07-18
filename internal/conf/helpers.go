package conf

import (
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func HasExternalAuthProvider(c conftypes.SiteConfigQuerier) bool {
	for _, p := range c.SiteConfig().AuthProviders {
		if p.Builtin == nil { // not builtin implies SSO
			return true
		}
	}
	return false
}

func GetDeduplicatedForksIndex() collections.Set[string] {
	index := collections.NewSet[string]()

	repoConf := Get().Repositories
	if repoConf == nil {
		return index
	}

	index.Add(repoConf.DeduplicateForks...)
	return index
}
