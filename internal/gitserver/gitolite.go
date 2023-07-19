package gitserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

type GitoliteLister struct {
	addrs      func() []string
	httpClient httpcli.Doer
	grpcClient ClientSource
	userAgent  string
}

func NewGitoliteLister(db database.DB, cli httpcli.Doer) *GitoliteLister {
	logger := log.Scoped("GitoliteLister", "logger scoped to a GitoliteLister")
	atomicConns := getAtomicGitserverConns(logger, db)

	return &GitoliteLister{
		httpClient: cli,
		addrs: func() []string {
			return atomicConns.get().Addresses
		},
		grpcClient: atomicConns,
		userAgent:  filepath.Base(os.Args[0]),
	}
}

func (c *GitoliteLister) ListRepos(ctx context.Context, gitoliteHost string) (list []*gitolite.Repo, err error) {
	addrs := c.addrs()
	if len(addrs) == 0 {
		panic("unexpected state: no gitserver addresses")
	}
	if internalgrpc.IsGRPCEnabled(ctx) {

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

	} else {
		// The gitserver calls the shared Gitolite server in response to this request, so
		// we need to only call a single gitserver (or else we'd get duplicate results).
		addr := addrForKey(gitoliteHost, addrs)

		req, err := http.NewRequest("GET", "http://"+addr+"/list-gitolite?gitolite="+url.QueryEscape(gitoliteHost), nil)
		if err != nil {
			return nil, err
		}
		// Set header so that the handler knows the request is from us
		req.Header.Set("X-Requested-With", "Sourcegraph")

		resp, err := c.httpClient.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&list)
		return list, err
	}
}
