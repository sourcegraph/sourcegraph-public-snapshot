package gitserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

type GitoliteLister struct {
	addrs      func() []string
	httpClient httpcli.Doer
}

func NewGitoliteLister(cli httpcli.Doer) *GitoliteLister {
	return &GitoliteLister{
		httpClient: cli,
		addrs: func() []string {
			return conf.Get().ServiceConnections().GitServers
		},
	}
}

func (c *GitoliteLister) ListRepos(ctx context.Context, gitoliteHost string) (list []*gitolite.Repo, err error) {
	addrs := c.addrs()
	if len(addrs) == 0 {
		panic("unexpected state: no gitserver addresses")
	}
	// The gitserver calls the shared Gitolite server in response to this request, so
	// we need to only call a single gitserver (or else we'd get duplicate results).
	addr := addrForKey(gitoliteHost, addrs)

	req, err := http.NewRequest("GET", "http://"+addr+"/list-gitolite?gitolite="+url.QueryEscape(gitoliteHost), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&list)
	return list, err
}
