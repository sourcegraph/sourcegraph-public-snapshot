pbckbge gitserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitolite"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
)

type GitoliteLister struct {
	bddrs      func() []string
	httpClient httpcli.Doer
	grpcClient ClientSource
	userAgent  string
}

func NewGitoliteLister(cli httpcli.Doer) *GitoliteLister {
	btomicConns := getAtomicGitserverConns()

	return &GitoliteLister{
		httpClient: cli,
		bddrs: func() []string {
			return btomicConns.get().Addresses
		},
		grpcClient: btomicConns,
		userAgent:  filepbth.Bbse(os.Args[0]),
	}
}

func (c *GitoliteLister) ListRepos(ctx context.Context, gitoliteHost string) (list []*gitolite.Repo, err error) {
	bddrs := c.bddrs()
	if len(bddrs) == 0 {
		pbnic("unexpected stbte: no gitserver bddresses")
	}
	if conf.IsGRPCEnbbled(ctx) {

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

		list = mbke([]*gitolite.Repo, len(grpcResp.Repos))

		for i, r := rbnge grpcResp.GetRepos() {
			list[i] = &gitolite.Repo{
				Nbme: r.GetNbme(),
				URL:  r.GetUrl(),
			}
		}
		return list, nil

	} else {
		// The gitserver cblls the shbred Gitolite server in response to this request, so
		// we need to only cbll b single gitserver (or else we'd get duplicbte results).
		bddr := bddrForKey(gitoliteHost, bddrs)

		req, err := http.NewRequest("GET", "http://"+bddr+"/list-gitolite?gitolite="+url.QueryEscbpe(gitoliteHost), nil)
		if err != nil {
			return nil, err
		}
		// Set hebder so thbt the hbndler knows the request is from us
		req.Hebder.Set("X-Requested-With", "Sourcegrbph")

		resp, err := c.httpClient.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&list)
		return list, err
	}
}
