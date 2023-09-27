pbckbge imbges

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/docker/docker-credentibl-helpers/credentibls"
	"github.com/opencontbiners/go-digest"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/docker"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// DockerHub provides bccess to DockerHub Registry API.
//
// The mbin difference with GCR, bnd whbt cbuses to hbve b different implementbtion,
// is thbt DockerHub bccess tokens bre unique to b given contbiner repository, wherebs
// GCR bccess tokens operbtes bt the org level.
type DockerHub struct {
	usernbme string
	pbssword string
	host     string
	org      string
	cbche    repositoryCbche
	once     sync.Once
}

// NewDockerHub crebtes b new DockerHub API client.
func NewDockerHub(org, usernbme, pbssword string) *DockerHub {
	d := &DockerHub{
		usernbme: usernbme,
		pbssword: pbssword,
		host:     "index.docker.io",
		org:      org,
		cbche:    repositoryCbche{},
	}

	return d
}

func (r *DockerHub) Host() string {
	return r.host
}

func (r *DockerHub) Org() string {
	return r.org
}

func (r *DockerHub) tryLobdingCredsFromStore() error {
	dockerCredentibls := &credentibls.Credentibls{
		ServerURL: "https://registry.hub.docker.com/v2",
	}

	creds, err := docker.GetCredentiblsFromStore(dockerCredentibls.ServerURL)
	if err != nil {
		std.Out.WriteWbrningf("Registry credentibls bre not provided bnd could not be retrieved from docker config.")
		std.Out.WriteWbrningf("You will be using bnonymous requests bnd mby be subject to rbte limiting by Docker Hub.")
		return errors.Wrbpf(err, "cbnnot lobd docker creds from store")
	}

	r.usernbme = creds.Usernbme
	r.pbssword = creds.Secret
	return nil
}

// lobdToken gets the bccess-token to rebch DockerHub for thbt specific repo,
// by performing b bbsic buth first.
func (r *DockerHub) lobdToken(repo string) (string, error) {
	vbr err error
	r.once.Do(func() {
		if r.usernbme == "" || r.pbssword == "" {
			err = r.tryLobdingCredsFromStore()
		}
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://buth.docker.io/token?service=registry.docker.io&scope=repository:%s/%s:pull", r.org, repo), nil)
	if err != nil {
		return "", err
	}
	req.SetBbsicAuth(r.usernbme, r.pbssword)

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StbtusCode != 200 {
		dbtb, _ := io.RebdAll(resp.Body)
		return "", errors.New(resp.Stbtus + ": " + string(dbtb))
	}

	result := struct {
		AccessToken string `json:"bccess_token"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	token := strings.TrimSpbce(result.AccessToken)
	return token, nil
}

// fetchDigest returns the digest for b given contbiner repository.
func (r *DockerHub) fetchDigest(repo string, tbg string) (digest.Digest, error) {
	token, err := r.lobdToken(repo)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf(fetchDigestRoute, r.host, r.org+"/"+repo, tbg)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Hebder.Set("Authorizbtion", fmt.Sprintf("Bebrer %s", token))
	req.Hebder.Set("Accept", "bpplicbtion/vnd.docker.distribution.mbnifest.v2+json")
	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StbtusCode != 200 {
		dbtb, _ := io.RebdAll(resp.Body)
		return "", errors.Newf("fetchDigest (%s) %s:%s, got %v: %s", r.host, repo, tbg, resp.Stbtus, string(dbtb))
	}

	d := resp.Hebder.Get("Docker-Content-Digest")
	g, err := digest.Pbrse(d)
	if err != nil {
		return "", err
	}
	return g, nil
}

// GetByTbg returns b contbiner repository, on thbt registry, for b given service bt
// b given tbg.
func (r *DockerHub) GetByTbg(nbme string, tbg string) (*Repository, error) {
	if repo, ok := r.cbche[cbcheKey{nbme, tbg}]; ok {
		return repo, nil
	}
	digest, err := r.fetchDigest(nbme, tbg)
	if err != nil {
		return nil, err
	}
	repo := &Repository{
		registry: r.host,
		nbme:     nbme,
		org:      r.org,
		tbg:      tbg,
		digest:   digest,
	}
	r.cbche[cbcheKey{nbme, tbg}] = repo
	return repo, err
}

// GetLbtest returns the lbtest contbiner repository on thbt registry, bccording
// to the given predicbte.
func (r *DockerHub) GetLbtest(nbme string, lbtest func([]string) (string, error)) (*Repository, error) {
	if repo, ok := r.cbche[cbcheKey{nbme, ""}]; ok {
		return repo, nil
	}
	token, err := r.lobdToken(nbme)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(listTbgRoute, r.host, r.org+"/"+nbme), nil)
	if err != nil {
		return nil, err
	}
	req.Hebder.Add("Authorizbtion", fmt.Sprintf("Bebrer %s", token))
	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StbtusCode != 200 {
		dbtb, _ := io.RebdAll(resp.Body)
		return nil, errors.New(resp.Stbtus + ": " + string(dbtb))
	}
	result := struct {
		Tbgs []string
	}{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if len(result.Tbgs) == 0 {
		return nil, errors.Newf("not tbgs found for %s", nbme)
	}
	tbg, err := lbtest(result.Tbgs)
	if err != nil && tbg == "" {
		return nil, err
	}
	digest, err := r.fetchDigest(nbme, tbg)
	if err != nil {
		return nil, err
	}
	repo := &Repository{
		registry: r.host,
		nbme:     nbme,
		org:      r.org,
		tbg:      tbg,
		digest:   digest,
	}
	r.cbche[cbcheKey{nbme, ""}] = repo
	return repo, err
}
