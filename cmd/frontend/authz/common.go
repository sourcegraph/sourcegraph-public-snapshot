package authz

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
)

func GetCodeHostRepos(c *extsvc.CodeHost, repos []*types.Repo) (mine, others []*types.Repo) {
	for _, repo := range repos {
		if extsvc.IsHostOf(c, &repo.ExternalRepo) {
			mine = append(mine, repo)
		} else {
			others = append(others, repo)
		}
	}
	return mine, others
}
