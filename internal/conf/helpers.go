package conf

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
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

func GetDeduplicatedForksIndex() map[api.RepoName]struct{} {
	index := map[api.RepoName]struct{}{}

	repoConf := Get().Repositories
	if repoConf == nil {
		return index
	}

	for _, name := range repoConf.DeduplicateForks {
		index[api.RepoName(name)] = struct{}{}
	}

	return index
}
