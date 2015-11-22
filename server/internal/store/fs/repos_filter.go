package fs

import (
	"sort"
	"strings"
	"time"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// reposFilter filters, sorts, and paginates repos according to
// opt. It may modify repos.
func reposFilter(repos []*sourcegraph.Repo, opt *sourcegraph.RepoListOptions) []*sourcegraph.Repo {
	if opt == nil {
		opt = &sourcegraph.RepoListOptions{}
	}

	// Filter
	repos2 := make([]*sourcegraph.Repo, 0, len(repos))
	for _, repo := range repos {
		if repoSatisfiesOpts(repo, opt) {
			repos2 = append(repos2, repo)
		}
	}
	repos = repos2

	// Sort.
	v := reposSorter{repos: repos}
	switch opt.Sort {
	case "pushed":
		v.less = func(a, b *sourcegraph.Repo) bool {
			var at, bt time.Time
			if a.PushedAt != nil {
				at = a.PushedAt.Time()
			}
			if b.PushedAt != nil {
				bt = b.PushedAt.Time()
			}
			return at.Before(bt)
		}
	case "path":
		fallthrough
	default:
		v.less = func(a, b *sourcegraph.Repo) bool {
			return a.URI < b.URI
		}
	}
	if opt.Direction == "desc" {
		sort.Sort(sort.Reverse(v))
	} else {
		sort.Sort(v)
	}

	// Paginate.
	offset, limit := opt.ListOptions.Offset(), opt.ListOptions.Limit()
	if offset > len(repos) {
		offset = len(repos)
	}
	repos = repos[offset:]
	if len(repos) > limit {
		repos = repos[:limit]
	}

	return repos
}

func repoSatisfiesOpts(repo *sourcegraph.Repo, opt *sourcegraph.RepoListOptions) bool {
	if opt == nil {
		return true
	}

	if query := opt.Query; query != "" {
		ok := func() bool {
			query = strings.ToLower(query)
			uri, name := strings.ToLower(repo.URI), strings.ToLower(repo.Name)

			if query == uri || strings.HasPrefix(name, query) {
				return true
			}

			// Match any path component prefix.
			for _, pc := range strings.Split(uri, "/") {
				if strings.HasPrefix(pc, query) {
					return true
				}
			}

			return false
		}()
		if !ok {
			return false
		}
	}

	if len(opt.URIs) > 0 {
		uriMatch := false
		for _, uri := range opt.URIs {
			if strings.EqualFold(uri, repo.URI) {
				uriMatch = true
				break
			}
		}
		if !uriMatch {
			return false
		}
	}

	return true
}

type reposSorter struct {
	repos []*sourcegraph.Repo
	less  func(a, b *sourcegraph.Repo) bool
}

func (bs reposSorter) Len() int           { return len(bs.repos) }
func (bs reposSorter) Swap(i, j int)      { bs.repos[i], bs.repos[j] = bs.repos[j], bs.repos[i] }
func (bs reposSorter) Less(i, j int) bool { return bs.less(bs.repos[i], bs.repos[j]) }
