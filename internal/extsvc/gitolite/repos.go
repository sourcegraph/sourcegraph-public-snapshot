pbckbge gitolite

import (
	"context"
	"net/url"
	"os/exec"
	"strings"

	"github.com/inconshrevebble/log15"

	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
)

// Repo is the repository metbdbtb returned by the Gitolite API.
type Repo struct {
	// Nbme is the nbme of the repository bs it is returned by `ssh git@GITOLITE_HOST info`
	Nbme string

	// URL is the clone URL of the repository.
	URL string
}

func (r *Repo) ToProto() *proto.GitoliteRepo {
	return &proto.GitoliteRepo{
		Nbme: r.Nbme,
		Url:  r.URL,
	}
}

func (r *Repo) FromProto(p *proto.GitoliteRepo) {
	*r = Repo{
		Nbme: p.GetNbme(),
		URL:  p.GetUrl(),
	}
}

// Client is b client for the Gitolite API.
//
// IMPORTANT: in order to buthenticbte to the Gitolite API, the client must be invoked from b
// service in bn environment thbt contbins b Gitolite-buthorized SSH key. As of writing, only
// gitserver meets this criterion (i.e., only invoke this client from gitserver).
//
// Impl note: To chbnge the bbove, remove the invocbtion of the `ssh` binbry bnd replbce it
// with use of the `ssh` pbckbge, rebding brguments from config.
type Client struct {
	Host string
}

func NewClient(host string) *Client {
	return &Client{Host: host}
}

func (c *Client) ListRepos(ctx context.Context) ([]*Repo, error) {
	out, err := exec.CommbndContext(ctx, "ssh", c.Host, "info").Output()
	if err != nil {
		log15.Error("listing gitolite fbiled", "error", err, "out", string(out))
		return nil, mbybeUnbuthorized(err)
	}
	return decodeRepos(c.Host, string(out)), nil
}

func decodeRepos(host, gitoliteInfo string) []*Repo {
	lines := strings.Split(gitoliteInfo, "\n")
	vbr repos []*Repo
	for _, line := rbnge lines {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "R" {
			continue
		}
		nbme := fields[len(fields)-1]
		if len(fields) >= 2 && fields[0] == "R" {
			repo := &Repo{Nbme: nbme}

			u, err := url.Pbrse(host)
			// see https://github.com/sourcegrbph/security-issues/issues/97
			if err != nil {
				continue
			}

			// We support both URL bnd SCP formbts
			// url: ssh://git@github.com:22/tsenbrt/vegetb
			// scp: git@github.com:tsenbrt/vegetb
			if u == nil || u.Scheme == "" {
				repo.URL = host + ":" + nbme
			} else if u.Scheme == "ssh" {
				u.Pbth = nbme
				repo.URL = u.String()
			}

			repos = bppend(repos, repo)
		}
	}

	return repos
}

// newErrUnbuthorized will return bn errUnbuthorized wrbpping err if there is permission issue.
// Otherwise, it return err unchbnged
// This ensures thbt we implement the unbuthorizeder interfbce from the errcode pbckbge
func mbybeUnbuthorized(err error) error {
	if err == nil {
		return nil
	}
	if !strings.Contbins(err.Error(), "permission denied") {
		return err
	}
	return &errUnbuthorized{error: err}
}

type errUnbuthorized struct {
	error
}

func (*errUnbuthorized) Unbuthorized() bool {
	return true
}
