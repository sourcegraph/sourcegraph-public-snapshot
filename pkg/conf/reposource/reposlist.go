package reposource

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	// reposListInstance is the global instance of the repos.list repoSource. Do NOT reference this
	// directly; use GetReposListInstance() instead.
	reposListInstance          atomic.Value
	reposListInstanceReadyOnce sync.Once
	reposListInstanceReady     = make(chan struct{})
)

func init() {
	go func() {
		conf.Watch(func() {
			reposListInstance.Store(newReposList(conf.Get().ReposList))
		})

		reposListInstanceReadyOnce.Do(func() {
			close(reposListInstanceReady)
		})
	}()
}

func GetReposListInstance() RepoSource {
	<-reposListInstanceReady
	return reposListInstance.Load().(*reposList)
}

type reposList struct {
	// cloneURLToName records the map from clone URL to repo name. It is read-only after construction,
	// so does not require synchronization.
	cloneURLToName map[string]string
}

var _ RepoSource = (*reposList)(nil)

func newReposList(repos []*schema.Repository) *reposList {
	cloneURLToName := map[string]string{}
	for _, rp := range repos {
		cloneURLToName[normalizeCloneURL(rp.Url)] = rp.Path
	}
	return &reposList{
		cloneURLToName: cloneURLToName,
	}
}

func (c *reposList) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	return api.RepoName(c.cloneURLToName[normalizeCloneURL(cloneURL)]), nil
}

// normalizeCloneURL attempts to reduce the cloneURL to a normalized form using some simple
// heuristics. If it finds the heuristics don't apply, it returns the original clone URL.
func normalizeCloneURL(cloneURL string) string {
	var (
		u   *url.URL
		err error
	)
	if strings.HasPrefix(cloneURL, "https://") || strings.HasPrefix(cloneURL, "http://") || strings.HasPrefix(cloneURL, "ssh://") {
		u, err = url.Parse(cloneURL)
	} else { // SCP-like case
		u, err = url.Parse("fake://" + strings.Replace(cloneURL, ":", "/", 1))
	}
	if err != nil {
		return cloneURL
	}
	return fmt.Sprintf("%s/%s", u.Hostname(), strings.TrimPrefix(strings.TrimSuffix(u.Path, ".git"), "/"))
}
