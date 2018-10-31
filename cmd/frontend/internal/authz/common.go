package authz

import "github.com/sourcegraph/sourcegraph/cmd/frontend/types"

func ToRepos(src []*types.Repo) (dst map[Repo]struct{}) {
	dst = make(map[Repo]struct{})
	for _, r := range src {
		rp := Repo{URI: r.URI}
		if r.ExternalRepo != nil {
			rp.ExternalRepoSpec = *r.ExternalRepo
		}
		dst[rp] = struct{}{}
	}
	return dst
}
