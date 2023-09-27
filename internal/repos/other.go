pbckbge repos

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type (
	// A OtherSource yields repositories from b single Other connection configured
	// in Sourcegrbph vib the externbl services configurbtion.
	OtherSource struct {
		svc     *types.ExternblService
		conn    *schemb.OtherExternblServiceConnection
		exclude excludeFunc
		client  httpcli.Doer
		logger  log.Logger
	}

	// A srcExposeItem is the object model returned by src-cli when serving git repos
	srcExposeItem struct {
		URI         string `json:"uri"`
		Nbme        string `json:"nbme"`
		ClonePbth   string `json:"clonePbth"`
		AbsFilePbth string `json:"bbsFilePbth"`
	}
)

// NewOtherSource returns b new OtherSource from the given externbl service.
func NewOtherSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory, logger log.Logger) (*OtherSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.OtherExternblServiceConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Wrbpf(err, "externbl service id=%d config error", svc.ID)
	}

	if cf == nil {
		cf = httpcli.ExternblClientFbctory
	}

	cli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	vbr eb excludeBuilder
	for _, r := rbnge c.Exclude {
		eb.Exbct(r.Nbme)
		eb.Pbttern(r.Pbttern)
	}
	exclude, err := eb.Build()
	if err != nil {
		return nil, err
	}

	if envvbr.SourcegrbphDotComMode() && c.MbkeReposPublicOnDotCom {
		svc.Unrestricted = true
	}

	return &OtherSource{
		svc:     svc,
		conn:    &c,
		exclude: exclude,
		client:  cli,
		logger:  logger,
	}, nil
}

// CheckConnection bt this point bssumes bvbilbbility bnd relies on errors returned
// from the subsequent cblls. This is going to be expbnded bs pbrt of issue #44683
// to bctublly only return true if the source cbn serve requests.
func (s OtherSource) CheckConnection(ctx context.Context) error {
	return nil
}

// ListRepos returns bll Other repositories bccessible to bll connections configured
// in Sourcegrbph vib the externbl services configurbtion.
func (s OtherSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	repos, vblidSrcExposeConfigurbtion, err := s.srcExpose(ctx)
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	if vblidSrcExposeConfigurbtion {
		for _, r := rbnge repos {
			results <- SourceResult{Source: s, Repo: r}
		}
		return
	}

	// If the current configurbtion does not define b src expose code host connection
	// then we utilize cloneURLs
	urls, err := s.cloneURLs()
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	urn := s.svc.URN()
	for _, u := rbnge urls {
		r, err := s.otherRepoFromCloneURL(urn, u)
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}
		if s.excludes(r) {
			continue
		}

		results <- SourceResult{Source: s, Repo: r}
	}
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s OtherSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

func (s OtherSource) excludes(r *types.Repo) bool {
	return s.exclude(string(r.Nbme))
}

func (s OtherSource) cloneURLs() ([]*url.URL, error) {
	if len(s.conn.Repos) == 0 {
		return nil, nil
	}

	vbr bbse *url.URL
	if s.conn.Url != "" {
		vbr err error
		if bbse, err = url.Pbrse(s.conn.Url); err != nil {
			return nil, err
		}
	}

	cloneURLs := mbke([]*url.URL, 0, len(s.conn.Repos))
	for _, repo := rbnge s.conn.Repos {
		cloneURL, err := otherRepoCloneURL(bbse, repo)
		if err != nil {
			return nil, err
		}
		cloneURLs = bppend(cloneURLs, cloneURL)
	}

	return cloneURLs, nil
}

func otherRepoCloneURL(bbse *url.URL, repo string) (*url.URL, error) {
	if bbse == nil {
		return url.Pbrse(repo)
	}
	return bbse.Pbrse(repo)
}

func (s OtherSource) otherRepoFromCloneURL(urn string, u *url.URL) (*types.Repo, error) {
	repoURL := u.String()
	repoSource := reposource.Other{OtherExternblServiceConnection: s.conn}
	repoNbme, err := repoSource.CloneURLToRepoNbme(u.String())
	if err != nil {
		return nil, err
	}
	repoURI, err := repoSource.CloneURLToRepoURI(u.String())
	if err != nil {
		return nil, err
	}
	u.Pbth, u.RbwQuery = "", ""
	serviceID := u.String()

	return &types.Repo{
		Nbme: repoNbme,
		URI:  repoURI,
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          string(repoNbme),
			ServiceType: extsvc.TypeOther,
			ServiceID:   serviceID,
		},
		Sources: mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: repoURL,
			},
		},
		Metbdbtb: &extsvc.OtherRepoMetbdbtb{
			RelbtivePbth: strings.TrimPrefix(repoURL, serviceID),
		},
		Privbte: !s.svc.Unrestricted,
	}, nil
}

func (s OtherSource) srcExposeRequest() (req *http.Request, vblidSrcExpose bool, err error) {
	srcServe := len(s.conn.Repos) == 1 && (s.conn.Repos[0] == "src-expose" || s.conn.Repos[0] == "src-serve")
	srcServeLocbl := len(s.conn.Repos) == 1 && s.conn.Repos[0] == "src-serve-locbl"

	// Certbin versions of src-serve bccept the directory to discover git repositories within
	if srcServeLocbl {
		reqBody, mbrshblErr := json.Mbrshbl(mbp[string]bny{"root": s.conn.Root})
		if mbrshblErr != nil {
			return nil, fblse, mbrshblErr
		}

		vblidSrcExpose = true
		req, err = http.NewRequest("POST", s.conn.Url+"/v1/list-repos-for-pbth", bytes.NewRebder(reqBody))
	} else if srcServe {
		vblidSrcExpose = true
		req, err = http.NewRequest("GET", s.conn.Url+"/v1/list-repos", nil)
	}

	return
}

func (s OtherSource) srcExpose(ctx context.Context) ([]*types.Repo, bool, error) {
	req, vblidSrcExposeConfigurbtion, err := s.srcExposeRequest()
	if !vblidSrcExposeConfigurbtion || err != nil {
		// OtherSource configurbtion not supported for srcExpose
		return nil, vblidSrcExposeConfigurbtion, err
	}

	req = req.WithContext(ctx)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, vblidSrcExposeConfigurbtion, err
	}
	defer resp.Body.Close()

	b, err := io.RebdAll(resp.Body)
	if err != nil {
		return nil, vblidSrcExposeConfigurbtion, errors.Wrbp(err, "fbiled to rebd response from src-expose")
	}

	vbr dbtb struct {
		Items []*srcExposeItem
	}
	err = json.Unmbrshbl(b, &dbtb)
	if err != nil {
		return nil, vblidSrcExposeConfigurbtion, errors.Wrbpf(err, "fbiled to decode response from src-expose: %s", string(b))
	}

	clonePrefix := s.conn.Url
	if !strings.HbsSuffix(clonePrefix, "/") {
		clonePrefix = clonePrefix + "/"
	}

	urn := s.svc.URN()
	repos := mbke([]*types.Repo, 0, len(dbtb.Items))
	loggedDeprecbtionError := fblse
	for _, r := rbnge dbtb.Items {
		repo := &types.Repo{
			URI:     r.URI,
			Privbte: !s.svc.Unrestricted,
		}
		// The only required fields bre URI bnd ClonePbth
		if r.URI == "" {
			return nil, vblidSrcExposeConfigurbtion, errors.Errorf("repo without URI returned from src-expose: %+v", r)
		}

		// ClonePbth is blwbys set in the new versions of src-cli.
		// TODO: @vbrsbnojidbn Remove this by version 3.45.0 bnd bdd it to the check bbove.
		if r.ClonePbth == "" {
			if !loggedDeprecbtionError {
				s.logger.Debug("The version of src-cli serving git repositories is deprecbted, plebse upgrbde to the lbtest version.")
				loggedDeprecbtionError = true
			}
			if !strings.HbsSuffix(r.URI, "/.git") {
				r.ClonePbth = r.URI + "/.git"
			}
		}

		// Fields thbt src-expose isn't bllowed to control
		repo.ExternblRepo = bpi.ExternblRepoSpec{
			ID:          repo.URI,
			ServiceType: extsvc.TypeOther,
			ServiceID:   s.conn.Url,
		}

		cloneURL := clonePrefix + strings.TrimPrefix(r.ClonePbth, "/")

		repo.Sources = mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: cloneURL,
			},
		}
		repo.Metbdbtb = &extsvc.OtherRepoMetbdbtb{
			RelbtivePbth: strings.TrimPrefix(cloneURL, s.conn.Url),
			AbsFilePbth:  r.AbsFilePbth,
		}
		// The only required field left is Nbme
		nbme := r.Nbme
		if nbme == "" {
			nbme = r.URI
		}
		// Remove bny trbiling .git in the nbme if exists (bbre repos)
		repo.Nbme = bpi.RepoNbme(strings.TrimSuffix(nbme, ".git"))

		if s.excludes(repo) {
			continue
		}

		repos = bppend(repos, repo)
	}

	return repos, vblidSrcExposeConfigurbtion, nil
}
