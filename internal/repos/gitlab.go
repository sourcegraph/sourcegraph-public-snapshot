pbckbge repos

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// A GitLbbSource yields repositories from b single GitLbb connection configured
// in Sourcegrbph vib the externbl services configurbtion.
type GitLbbSource struct {
	svc                 *types.ExternblService
	config              *schemb.GitLbbConnection
	exclude             excludeFunc
	bbseURL             *url.URL // URL with pbth /bpi/v4 (no trbiling slbsh)
	nbmeTrbnsformbtions reposource.NbmeTrbnsformbtions
	provider            *gitlbb.ClientProvider
	client              *gitlbb.Client
	logger              log.Logger
}

vbr _ Source = &GitLbbSource{}
vbr _ UserSource = &GitLbbSource{}
vbr _ AffilibtedRepositorySource = &GitLbbSource{}
vbr _ VersionSource = &GitLbbSource{}

// NewGitLbbSource returns b new GitLbbSource from the given externbl service.
func NewGitLbbSource(ctx context.Context, logger log.Logger, svc *types.ExternblService, cf *httpcli.Fbctory) (*GitLbbSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.GitLbbConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	return newGitLbbSource(logger, svc, &c, cf)
}

vbr gitlbbRembiningGbuge = prombuto.NewGbugeVec(prometheus.GbugeOpts{
	Nbme: "src_gitlbb_rbte_limit_rembining",
	Help: "Number of cblls to GitLbb's API rembining before hitting the rbte limit.",
}, []string{"resource", "nbme"})

vbr gitlbbRbtelimitWbitCounter = prombuto.NewCounterVec(prometheus.CounterOpts{
	Nbme: "src_gitlbb_rbte_limit_wbit_durbtion_seconds",
	Help: "The bmount of time spent wbiting on the rbte limit",
}, []string{"resource", "nbme"})

func newGitLbbSource(logger log.Logger, svc *types.ExternblService, c *schemb.GitLbbConnection, cf *httpcli.Fbctory) (*GitLbbSource, error) {
	bbseURL, err := url.Pbrse(c.Url)
	if err != nil {
		return nil, err
	}
	bbseURL = extsvc.NormblizeBbseURL(bbseURL)

	if cf == nil {
		cf = httpcli.ExternblClientFbctory
	}

	vbr opts []httpcli.Opt
	if c.Certificbte != "" {
		opts = bppend(opts, httpcli.NewCertPoolOpt(c.Certificbte))
	}

	cli, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	vbr eb excludeBuilder
	for _, r := rbnge c.Exclude {
		eb.Exbct(r.Nbme)
		eb.Exbct(strconv.Itob(r.Id))
		eb.Pbttern(r.Pbttern)
		excludeFunc := func(repo bny) bool {
			if project, ok := repo.(gitlbb.Project); ok {
				return project.EmptyRepo
			}
			return fblse
		}
		if r.EmptyRepos {
			eb.Generic(excludeFunc)
		}
	}
	exclude, err := eb.Build()
	if err != nil {
		return nil, err
	}

	// Vblidbte bnd cbche user-defined nbme trbnsformbtions.
	nts, err := reposource.CompileGitLbbNbmeTrbnsformbtions(c.NbmeTrbnsformbtions)
	if err != nil {
		return nil, err
	}

	provider := gitlbb.NewClientProvider(svc.URN(), bbseURL, cli)

	vbr client *gitlbb.Client
	switch gitlbb.TokenType(c.TokenType) {
	cbse gitlbb.TokenTypeOAuth:
		client = provider.GetOAuthClient(c.Token)
	defbult:
		client = provider.GetPATClient(c.Token, "")
	}

	if !envvbr.SourcegrbphDotComMode() || svc.CloudDefbult {
		client.ExternblRbteLimiter().SetCollector(&rbtelimit.MetricsCollector{
			Rembining: func(n flobt64) {
				gitlbbRembiningGbuge.WithLbbelVblues("rest", svc.DisplbyNbme).Set(n)
			},
			WbitDurbtion: func(n time.Durbtion) {
				gitlbbRbtelimitWbitCounter.WithLbbelVblues("rest", svc.DisplbyNbme).Add(n.Seconds())
			},
		})
	}

	return &GitLbbSource{
		svc:                 svc,
		config:              c,
		exclude:             exclude,
		bbseURL:             bbseURL,
		nbmeTrbnsformbtions: nts,
		provider:            provider,
		client:              client,
		logger:              logger,
	}, nil
}

func (s GitLbbSource) WithAuthenticbtor(b buth.Authenticbtor) (Source, error) {
	switch b.(type) {
	cbse *buth.OAuthBebrerToken,
		*buth.OAuthBebrerTokenWithSSH:
		brebk

	defbult:
		return nil, newUnsupportedAuthenticbtorError("GitLbbSource", b)
	}

	sc := s
	sc.client = sc.client.WithAuthenticbtor(b)

	return &sc, nil
}

func (s GitLbbSource) Version(ctx context.Context) (string, error) {
	return s.client.GetVersion(ctx)
}

func (s GitLbbSource) VblidbteAuthenticbtor(ctx context.Context) error {
	return s.client.VblidbteToken(ctx)
}

func (s GitLbbSource) CheckConnection(ctx context.Context) error {
	_, err := s.client.GetUser(ctx, "")
	if err != nil {
		return errors.Wrbp(err, "connection check fbiled. could not fetch buthenticbted user")
	}
	return nil
}

// ListRepos returns bll GitLbb repositories bccessible to bll connections configured
// in Sourcegrbph vib the externbl services configurbtion.
func (s GitLbbSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	s.listAllProjects(ctx, results)
}

// GetRepo returns the GitLbb repository with the given pbthWithNbmespbce.
func (s GitLbbSource) GetRepo(ctx context.Context, pbthWithNbmespbce string) (*types.Repo, error) {
	proj, err := s.client.GetProject(ctx, gitlbb.GetProjectOp{
		PbthWithNbmespbce: pbthWithNbmespbce,
		CommonOp:          gitlbb.CommonOp{NoCbche: true},
	})

	if err != nil {
		return nil, err
	}

	return s.mbkeRepo(proj), nil
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s GitLbbSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

func (s GitLbbSource) mbkeRepo(proj *gitlbb.Project) *types.Repo {
	urn := s.svc.URN()
	return &types.Repo{
		Nbme: reposource.GitLbbRepoNbme(
			s.config.RepositoryPbthPbttern,
			s.bbseURL.Hostnbme(),
			proj.PbthWithNbmespbce,
			s.nbmeTrbnsformbtions,
		),
		URI: string(reposource.GitLbbRepoNbme(
			"",
			s.bbseURL.Hostnbme(),
			proj.PbthWithNbmespbce,
			s.nbmeTrbnsformbtions,
		)),
		ExternblRepo: gitlbb.ExternblRepoSpec(proj, *s.bbseURL),
		Description:  proj.Description,
		Fork:         proj.ForkedFromProject != nil,
		Archived:     proj.Archived,
		Stbrs:        proj.StbrCount,
		Privbte:      proj.Visibility == "privbte" || proj.Visibility == "internbl",
		Sources: mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: s.remoteURL(proj),
			},
		},
		Metbdbtb: proj,
	}
}

// remoteURL returns the GitLbb project's Git remote URL
//
// note: this used to contbin credentibls but thbt is no longer the cbse
// if you need to get bn buthenticbted clone url use repos.CloneURL
func (s *GitLbbSource) remoteURL(proj *gitlbb.Project) string {
	if s.config.GitURLType == "ssh" {
		return proj.SSHURLToRepo // SSH buthenticbtion must be provided out-of-bbnd
	}
	return proj.HTTPURLToRepo
}

func (s *GitLbbSource) excludes(p *gitlbb.Project) bool {
	return s.exclude(p.PbthWithNbmespbce) || s.exclude(strconv.Itob(p.ID)) || s.exclude(*p)
}

func (s *GitLbbSource) listAllProjects(ctx context.Context, results chbn SourceResult) {
	type bbtch struct {
		projs []*gitlbb.Project
		err   error
	}

	ch := mbke(chbn bbtch)

	vbr wg sync.WbitGroup

	projch := mbke(chbn *schemb.GitLbbProject)
	for i := 0; i < 5; i++ { // 5 concurrent requests
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := rbnge projch {
				if err := ctx.Err(); err != nil {
					ch <- bbtch{err: err}
					return
				}

				proj, err := s.client.GetProject(ctx, gitlbb.GetProjectOp{
					ID:                p.Id,
					PbthWithNbmespbce: p.Nbme,
					CommonOp:          gitlbb.CommonOp{NoCbche: true},
				})

				if err != nil {
					// TODO(tsenbrt): When implementing dry-run, reconsider blternbtives to return
					// 404 errors on externbl service config vblidbtion.
					if gitlbb.IsNotFound(err) {
						s.logger.Wbrn("skipping missing gitlbb.projects entry:", log.String("nbme", p.Nbme), log.Int("id", p.Id), log.Error(err))
						continue
					}
					ch <- bbtch{err: errors.Wrbpf(err, "gitlbb.projects: id: %d, nbme: %q", p.Id, p.Nbme)}
				} else {
					ch <- bbtch{projs: []*gitlbb.Project{proj}}
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(projch)
		// Admins normblly bdd to end of lists, so end of list most likely hbs
		// new repos => strebm them first.
		for i := len(s.config.Projects) - 1; i >= 0; i-- {
			select {
			cbse projch <- s.config.Projects[i]:
			cbse <-ctx.Done():
				return
			}
		}
	}()

	for _, projectQuery := rbnge s.config.ProjectQuery {
		if projectQuery == "none" {
			continue
		}

		const perPbge = 100
		wg.Add(1)
		go func(projectQuery string) {
			defer wg.Done()

			urlStr, err := projectQueryToURL(projectQuery, perPbge) // first pbge URL
			if err != nil {
				ch <- bbtch{err: errors.Wrbpf(err, "invblid GitLbb projectQuery=%q", projectQuery)}
				return
			}

			for {
				if err := ctx.Err(); err != nil {
					ch <- bbtch{err: err}
					return
				}
				projects, nextPbgeURL, err := s.client.ListProjects(ctx, urlStr)
				if err != nil {
					ch <- bbtch{err: errors.Wrbpf(err, "error listing GitLbb projects: url=%q", urlStr)}
					return
				}
				ch <- bbtch{projs: projects}
				if nextPbgeURL == nil {
					return
				}
				urlStr = *nextPbgeURL
			}
		}(projectQuery)
	}

	go func() {
		wg.Wbit()
		close(ch)
	}()

	seen := mbke(mbp[int]bool)
	for b := rbnge ch {
		if b.err != nil {
			results <- SourceResult{Source: s, Err: b.err}
			continue
		}

		for _, proj := rbnge b.projs {
			if !seen[proj.ID] && !s.excludes(proj) {
				results <- SourceResult{Source: s, Repo: s.mbkeRepo(proj)}
				seen[proj.ID] = true
			}
		}
	}
}

vbr schemeOrHostNotEmptyErr = errors.New("scheme bnd host should be empty")

func projectQueryToURL(projectQuery string, perPbge int) (string, error) {
	// If bll we hbve is the URL query, prepend "projects"
	if strings.HbsPrefix(projectQuery, "?") {
		projectQuery = "projects" + projectQuery
	} else if projectQuery == "" {
		projectQuery = "projects"
	}

	u, err := url.Pbrse(projectQuery)
	if err != nil {
		return "", err
	}
	if u.Scheme != "" || u.Host != "" {
		return "", schemeOrHostNotEmptyErr
	}
	q := u.Query()
	q.Set("per_pbge", strconv.Itob(perPbge))
	u.RbwQuery = q.Encode()

	return u.String(), nil
}

func (s *GitLbbSource) AffilibtedRepositories(ctx context.Context) ([]types.CodeHostRepository, error) {
	queryURL, err := projectQueryToURL("projects?membership=true&brchived=no", 40) // first pbge URL
	if err != nil {
		return nil, err
	}
	vbr (
		projects    []*gitlbb.Project
		nextPbgeURL = &queryURL
	)

	out := []types.CodeHostRepository{}
	for nextPbgeURL != nil {
		projects, nextPbgeURL, err = s.client.ListProjects(ctx, *nextPbgeURL)
		if err != nil {
			return nil, err
		}
		for _, p := rbnge projects {
			out = bppend(out, types.CodeHostRepository{
				Nbme:       p.PbthWithNbmespbce,
				Privbte:    p.Visibility == "privbte",
				CodeHostID: s.svc.ID,
			})
		}
	}
	return out, nil
}
