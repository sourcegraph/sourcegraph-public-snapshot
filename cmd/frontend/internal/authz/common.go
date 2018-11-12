package authz

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
)

func ToRepos(src []*types.Repo) (dst map[Repo]struct{}) {
	dst = make(map[Repo]struct{})
	for _, r := range src {
		rp := Repo{RepoName: r.Name}
		if r.ExternalRepo != nil {
			rp.ExternalRepoSpec = *r.ExternalRepo
		}
		dst[rp] = struct{}{}
	}
	return dst
}

func GetCodeHostRepos(c extsvc.CodeHost, repos map[Repo]struct{}) (mine map[Repo]struct{}, others map[Repo]struct{}) {
	mine, others = make(map[Repo]struct{}), make(map[Repo]struct{})
	for repo := range repos {
		if extsvc.IsHostOf(c, &repo.ExternalRepoSpec) {
			mine[repo] = struct{}{}
		} else {
			others[repo] = struct{}{}
		}
	}
	return mine, others
}
