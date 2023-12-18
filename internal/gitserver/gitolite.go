package gitserver

import (
	"context"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

type GitoliteLister struct {
	addrs      func() []string
	httpClient httpcli.Doer
	grpcClient ClientSource
	userAgent  string
}

func NewGitoliteLister(cli httpcli.Doer) *GitoliteLister {
	return &GitoliteLister{
		httpClient: cli,
		addrs: func() []string {
			return conns.get().Addresses
		},
		grpcClient: conns,
		userAgent:  filepath.Base(os.Args[0]),
	}
}

func (c *GitoliteLister) ListRepos(ctx context.Context, gitoliteHost string) (list []*gitolite.Repo, err error) {
	addrs := c.addrs()
	if len(addrs) == 0 {
		panic("unexpected state: no gitserver addresses")
	}

	client, err := c.grpcClient.ClientForRepo(ctx, c.userAgent, "")
	if err != nil {
		return nil, err
	}
	req := &proto.ListGitoliteRequest{
		GitoliteHost: gitoliteHost,
	}
	grpcResp, err := client.ListGitolite(ctx, req)
	if err != nil {
		return nil, err
	}

	list = make([]*gitolite.Repo, len(grpcResp.Repos))

	for i, r := range grpcResp.GetRepos() {
		list[i] = &gitolite.Repo{
			Name: r.GetName(),
			URL:  r.GetUrl(),
		}
	}
	return list, nil
}
