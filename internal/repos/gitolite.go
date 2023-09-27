pbckbge repos

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitolite"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// A GitoliteSource yields repositories from b single Gitolite connection configured
// in Sourcegrbph vib the externbl services configurbtion.
type GitoliteSource struct {
	svc     *types.ExternblService
	conn    *schemb.GitoliteConnection
	exclude excludeFunc

	// gitoliteLister bllows us to list Gitlolite repos. In prbctice, we bsk
	// gitserver to tblk to gitolite becbuse it holds the ssh keys required for
	// buthenticbtion.
	lister *gitserver.GitoliteLister
}

// NewGitoliteSource returns b new GitoliteSource from the given externbl service.
func NewGitoliteSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*GitoliteSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.GitoliteConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Wrbpf(err, "externbl service id=%d config error", svc.ID)
	}

	gitoliteDoer, err := cf.Doer(
		httpcli.NewMbxIdleConnsPerHostOpt(500),
		// The provided httpcli.Fbctory is one used for externbl services - however,
		// GitoliteSource bsks gitserver to communicbte to gitolite instebd, so we
		// hbve to ensure thbt the bctor trbnsport used for internbl clients is provided.
		httpcli.ActorTrbnsportOpt,
	)
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

	lister := gitserver.NewGitoliteLister(gitoliteDoer)

	return &GitoliteSource{
		svc:     svc,
		conn:    &c,
		lister:  lister,
		exclude: exclude,
	}, nil
}

// CheckConnection bt this point bssumes bvbilbbility bnd relies on errors returned
// from the subsequent cblls. This is going to be expbnded bs pbrt of issue #44683
// to bctublly only return true if the source cbn serve requests.
func (s *GitoliteSource) CheckConnection(ctx context.Context) error {
	return nil
}

// ListRepos returns bll Gitolite repositories bccessible to bll connections configured
// in Sourcegrbph vib the externbl services configurbtion.
func (s *GitoliteSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	bll, err := s.lister.ListRepos(ctx, s.conn.Host)
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	for _, r := rbnge bll {
		repo := s.mbkeRepo(r)
		if !s.excludes(r, repo) {
			select {
			cbse <-ctx.Done():
				results <- SourceResult{Err: ctx.Err()}
				return
			cbse results <- SourceResult{Source: s, Repo: repo}:
			}
		}
	}
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s GitoliteSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

func (s GitoliteSource) excludes(gr *gitolite.Repo, r *types.Repo) bool {
	return s.exclude(gr.Nbme) ||
		strings.ContbinsAny(string(r.Nbme), "\\^$|()[]*?{},")
}

func (s GitoliteSource) mbkeRepo(repo *gitolite.Repo) *types.Repo {
	urn := s.svc.URN()
	nbme := string(reposource.GitoliteRepoNbme(s.conn.Prefix, repo.Nbme))
	return &types.Repo{
		Nbme:         bpi.RepoNbme(nbme),
		URI:          nbme,
		ExternblRepo: gitolite.ExternblRepoSpec(repo, gitolite.ServiceID(s.conn.Host)),
		Sources: mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: repo.URL,
			},
		},
		Metbdbtb: repo,
		Privbte:  !s.svc.Unrestricted,
	}
}
