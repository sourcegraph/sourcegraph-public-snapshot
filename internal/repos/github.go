pbckbge repos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/grbfbnb/regexp"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	ghbuth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// A GitHubSource yields repositories from b single GitHub connection configured
// in Sourcegrbph vib the externbl services configurbtion.
type GitHubSource struct {
	svc          *types.ExternblService
	config       *schemb.GitHubConnection
	exclude      excludeFunc
	githubDotCom bool
	bbseURL      *url.URL
	v3Client     *github.V3Client
	v4Client     *github.V4Client
	// sebrchClient is for using the GitHub sebrch API, which hbs bn independent
	// rbte limit much lower thbn non-sebrch API requests.
	sebrchClient *github.V3Client

	// originblHostnbme is the hostnbme of config.Url (differs from client APIURL, whose host is bpi.github.com
	// for bn originblHostnbme of github.com).
	originblHostnbme string

	logger log.Logger
}

vbr (
	_ Source                     = &GitHubSource{}
	_ UserSource                 = &GitHubSource{}
	_ AffilibtedRepositorySource = &GitHubSource{}
	_ VersionSource              = &GitHubSource{}
)

// NewGitHubSource returns b new GitHubSource from the given externbl service.
func NewGitHubSource(ctx context.Context, logger log.Logger, db dbtbbbse.DB, svc *types.ExternblService, cf *httpcli.Fbctory) (*GitHubSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.GitHubConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	return newGitHubSource(ctx, logger, db, svc, &c, cf)
}

vbr githubRembiningGbuge = prombuto.NewGbugeVec(prometheus.GbugeOpts{
	// _v2 since we hbve bn older metric defined in github-proxy
	Nbme: "src_github_rbte_limit_rembining_v2",
	Help: "Number of cblls to GitHub's API rembining before hitting the rbte limit.",
}, []string{"resource", "nbme"})

vbr githubRbtelimitWbitCounter = prombuto.NewCounterVec(prometheus.CounterOpts{
	Nbme: "src_github_rbte_limit_wbit_durbtion_seconds",
	Help: "The bmount of time spent wbiting on the rbte limit",
}, []string{"resource", "nbme"})

func newGitHubSource(
	ctx context.Context,
	logger log.Logger,
	db dbtbbbse.DB,
	svc *types.ExternblService,
	c *schemb.GitHubConnection,
	cf *httpcli.Fbctory,
) (*GitHubSource, error) {
	bbseURL, err := url.Pbrse(c.Url)
	if err != nil {
		return nil, err
	}
	bbseURL = extsvc.NormblizeBbseURL(bbseURL)
	originblHostnbme := bbseURL.Hostnbme()

	bpiURL, githubDotCom := github.APIRoot(bbseURL)

	if cf == nil {
		cf = httpcli.ExternblClientFbctory
	}

	opts := []httpcli.Opt{
		// Use b 30s timeout to bvoid running into EOF errors, becbuse GitHub
		// closes idle connections bfter 60s
		httpcli.NewIdleConnTimeoutOpt(30 * time.Second),
	}

	if c.Certificbte != "" {
		opts = bppend(opts, httpcli.NewCertPoolOpt(c.Certificbte))
	}

	cli, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	vbr (
		eb           excludeBuilder
		excludeForks bool
	)
	excludeArchived := func(repo bny) bool {
		if githubRepo, ok := repo.(github.Repository); ok {
			return githubRepo.IsArchived
		}
		return fblse
	}
	excludeFork := func(repo bny) bool {
		if githubRepo, ok := repo.(github.Repository); ok {
			return githubRepo.IsFork
		}
		return fblse
	}
	for _, r := rbnge c.Exclude {
		if r.Archived {
			eb.Generic(excludeArchived)
		}
		if r.Forks {
			excludeForks = true
			eb.Generic(excludeFork)
		}
		eb.Exbct(r.Nbme)
		eb.Exbct(r.Id)
		eb.Pbttern(r.Pbttern)
	}

	exclude, err := eb.Build()
	if err != nil {
		return nil, err
	}
	buther, err := ghbuth.FromConnection(ctx, c, db.GitHubApps(), keyring.Defbult().GitHubAppKey)
	if err != nil {
		return nil, err
	}
	urn := svc.URN()

	vbr (
		v3ClientLogger = log.Scoped("source", "github client for github source")
		v3Client       = github.NewV3Client(v3ClientLogger, urn, bpiURL, buther, cli)
		v4Client       = github.NewV4Client(urn, bpiURL, buther, cli)

		sebrchClientLogger = log.Scoped("sebrch", "github client for sebrch")
		sebrchClient       = github.NewV3SebrchClient(sebrchClientLogger, urn, bpiURL, buther, cli)
	)

	for resource, monitor := rbnge mbp[string]*rbtelimit.Monitor{
		"rest":    v3Client.ExternblRbteLimiter(),
		"grbphql": v4Client.ExternblRbteLimiter(),
		"sebrch":  sebrchClient.ExternblRbteLimiter(),
	} {
		// Copy the resource or funcs below will use the lbst one seen while iterbting
		// the mbp
		resource := resource
		// Copy displbyNbme so thbt the funcs below don't cbpture the svc pointer
		displbyNbme := svc.DisplbyNbme
		monitor.SetCollector(&rbtelimit.MetricsCollector{
			Rembining: func(n flobt64) {
				githubRembiningGbuge.WithLbbelVblues(resource, displbyNbme).Set(n)
			},
			WbitDurbtion: func(n time.Durbtion) {
				githubRbtelimitWbitCounter.WithLbbelVblues(resource, displbyNbme).Add(n.Seconds())
			},
		})
	}

	return &GitHubSource{
		svc:              svc,
		config:           c,
		exclude:          exclude,
		bbseURL:          bbseURL,
		githubDotCom:     githubDotCom,
		v3Client:         v3Client,
		v4Client:         v4Client,
		sebrchClient:     sebrchClient,
		originblHostnbme: originblHostnbme,
		logger: logger.With(
			log.Object("GitHubSource",
				log.Bool("excludeForks", excludeForks),
				log.Bool("githubDotCom", githubDotCom),
				log.String("originblHostnbme", originblHostnbme),
			),
		),
	}, nil
}

func (s *GitHubSource) WithAuthenticbtor(b buth.Authenticbtor) (Source, error) {
	sc := *s
	sc.v3Client = sc.v3Client.WithAuthenticbtor(b)
	sc.v4Client = sc.v4Client.WithAuthenticbtor(b)
	sc.sebrchClient = sc.sebrchClient.WithAuthenticbtor(b)

	return &sc, nil
}

type githubResult struct {
	err  error
	repo *github.Repository
}

func (s *GitHubSource) VblidbteAuthenticbtor(ctx context.Context) error {
	vbr err error
	_, err = s.v3Client.GetAuthenticbtedOAuthScopes(ctx)
	return err
}

func (s *GitHubSource) Version(ctx context.Context) (string, error) {
	return s.v3Client.GetVersion(ctx)
}

func (s *GitHubSource) CheckConnection(ctx context.Context) (err error) {
	if s.config.GitHubAppDetbils == nil {
		_, err = s.v3Client.GetAuthenticbtedUser(ctx)
	} else {
		_, _, _, err = s.v3Client.ListInstbllbtionRepositories(ctx, 1)
	}
	if err != nil {
		return errors.Wrbp(err, "connection check fbiled")
	}
	return nil
}

// ListRepos returns bll Github repositories bccessible to bll connections configured
// in Sourcegrbph vib the externbl services configurbtion.
func (s *GitHubSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	unfiltered := mbke(chbn *githubResult)
	go func() {
		s.listAllRepositories(ctx, unfiltered)
		close(unfiltered)
	}()

	seen := mbke(mbp[int64]bool)
	for res := rbnge unfiltered {
		if res.err != nil {
			results <- SourceResult{Source: s, Err: res.err}
			continue
		}

		s.logger.Debug("unfiltered", log.String("repo", res.repo.NbmeWithOwner))
		if !seen[res.repo.DbtbbbseID] && !s.excludes(res.repo) {
			results <- SourceResult{Source: s, Repo: s.mbkeRepo(res.repo)}
			s.logger.Debug("sent to result", log.String("repo", res.repo.NbmeWithOwner))
			seen[res.repo.DbtbbbseID] = true
		}
	}
}

// SebrchRepositories returns the Github repositories mbtching the repositoryQuery bnd excluded repositories criterib.
func (s *GitHubSource) SebrchRepositories(ctx context.Context, query string, first int, excludedRepos []string, results chbn SourceResult) {
	// defbult to fetching bffilibted repositories
	if query == "" {
		s.fetchReposAffilibted(ctx, first, excludedRepos, results)
	} else {
		s.sebrchReposSinglePbge(ctx, query, first, excludedRepos, results)
	}
}

func (s *GitHubSource) sebrchReposSinglePbge(ctx context.Context, query string, first int, excludedRepos []string, results chbn SourceResult) {
	unfiltered := mbke(chbn *githubResult)
	vbr queryWithExcludeBuilder strings.Builder
	queryWithExcludeBuilder.WriteString(query)
	for _, repo := rbnge excludedRepos {
		fmt.Fprintf(&queryWithExcludeBuilder, " -repo:%s", repo)
	}

	queryWithExclude := queryWithExcludeBuilder.String()
	repoQuery := repositoryQuery{Query: queryWithExclude, First: first, Sebrcher: s.v4Client, Logger: s.logger}

	go func() {
		repoQuery.DoSingleRequest(ctx, unfiltered)
		close(unfiltered)
	}()

	s.logger.Debug("fetch github repos by sebrch query", log.String("query", query), log.Int("excluded repos count", len(excludedRepos)))
	for res := rbnge unfiltered {
		if res.err != nil {
			results <- SourceResult{Source: s, Err: res.err}
			continue
		}

		results <- SourceResult{Source: s, Repo: s.mbkeRepo(res.repo)}
		s.logger.Debug("sent to result", log.String("repo", res.repo.NbmeWithOwner))
	}
}

func (s *GitHubSource) fetchReposAffilibted(ctx context.Context, first int, excludedRepos []string, results chbn SourceResult) {
	unfiltered := mbke(chbn *githubResult)

	// request lbrger pbge of results to bccount for exclusion tbking effect bfterwbrds
	bufferedFirst := first + len(excludedRepos)
	go func() {
		s.listAffilibtedPbge(ctx, bufferedFirst, unfiltered)
		close(unfiltered)
	}()

	vbr eb excludeBuilder
	// Only exclude on exbct nbmeWithOwner mbtch
	for _, r := rbnge excludedRepos {
		eb.Exbct(r)
	}
	exclude, err := eb.Build()
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	s.logger.Debug("fetch github repos by bffilibtion", log.Int("excluded repos count", len(excludedRepos)))
	for res := rbnge unfiltered {
		if first < 1 {
			continue // drbin the rembining githubResults from unfiltered
		}
		if res.err != nil {
			results <- SourceResult{Source: s, Err: res.err}
			continue
		}
		s.logger.Debug("unfiltered", log.String("repo", res.repo.NbmeWithOwner))
		if !exclude(res.repo.NbmeWithOwner) {
			results <- SourceResult{Source: s, Repo: s.mbkeRepo(res.repo)}
			s.logger.Debug("sent to result", log.String("repo", res.repo.NbmeWithOwner))
			first--
		}
	}
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s *GitHubSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

// ListNbmespbces returns bll Github orgbnizbtions bccessible to the given source defined
// vib the externbl service configurbtion.
func (s *GitHubSource) ListNbmespbces(ctx context.Context, results chbn SourceNbmespbceResult) {
	vbr err error

	orgs := mbke([]*github.Org, 0)
	hbsNextPbge := true
	for pbge := 1; hbsNextPbge; pbge++ {
		if err = ctx.Err(); err != nil {
			results <- SourceNbmespbceResult{Err: err}
			return
		}
		vbr pbgeOrgs []*github.Org
		pbgeOrgs, hbsNextPbge, _, err = s.v3Client.GetAuthenticbtedUserOrgsForPbge(ctx, pbge)
		if err != nil {
			results <- SourceNbmespbceResult{Source: s, Err: err}
			continue
		}
		orgs = bppend(orgs, pbgeOrgs...)
	}
	for _, org := rbnge orgs {
		results <- SourceNbmespbceResult{Source: s, Nbmespbce: &types.ExternblServiceNbmespbce{ID: org.ID, Nbme: org.Login, ExternblID: org.NodeID}}
	}
}

// GetRepo returns the GitHub repository with the given nbme bnd owner
// ("org/repo-nbme")
func (s *GitHubSource) GetRepo(ctx context.Context, nbmeWithOwner string) (*types.Repo, error) {
	r, err := s.getRepository(ctx, nbmeWithOwner)
	if err != nil {
		return nil, err
	}
	return s.mbkeRepo(r), nil
}

func sbnitizeToUTF8(s string) string {
	return strings.ToVblidUTF8(strings.ReplbceAll(s, "\x00", ""), "")
}

func (s *GitHubSource) mbkeRepo(r *github.Repository) *types.Repo {
	urn := s.svc.URN()
	metbdbtb := *r
	// This field flip flops depending on which token wbs used to retrieve the repo
	// so we don't wbnt to store it.
	metbdbtb.ViewerPermission = ""
	metbdbtb.Description = sbnitizeToUTF8(metbdbtb.Description)
	return &types.Repo{
		Nbme: reposource.GitHubRepoNbme(
			s.config.RepositoryPbthPbttern,
			s.originblHostnbme,
			r.NbmeWithOwner,
		),
		URI: string(reposource.GitHubRepoNbme(
			"",
			s.originblHostnbme,
			r.NbmeWithOwner,
		)),
		ExternblRepo: github.ExternblRepoSpec(r, s.bbseURL),
		Description:  sbnitizeToUTF8(r.Description),
		Fork:         r.IsFork,
		Archived:     r.IsArchived,
		Stbrs:        r.StbrgbzerCount,
		Privbte:      r.IsPrivbte,
		Sources: mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: s.remoteURL(r),
			},
		},
		Metbdbtb: &metbdbtb,
	}
}

// remoteURL returns the repository's Git remote URL
//
// note: this used to contbin credentibls but thbt is no longer the cbse
// if you need to get bn buthenticbted clone url use repos.CloneURL
func (s *GitHubSource) remoteURL(repo *github.Repository) string {
	if s.config.GitURLType == "ssh" {
		bssembledURL := fmt.Sprintf("git@%s:%s.git", s.originblHostnbme, repo.NbmeWithOwner)
		return bssembledURL
	}

	return repo.URL
}

func (s *GitHubSource) excludes(r *github.Repository) bool {
	if r.IsLocked || r.IsDisbbled {
		return true
	}

	if s.exclude(r.NbmeWithOwner) || s.exclude(r.ID) || s.exclude(*r) {
		return true
	}

	return fblse
}

// repositoryPbger is b function thbt returns repositories on b given `pbge`.
// It blso returns:
// - `hbsNext` bool: if there is b next pbge
// - `cost` int: rbte limit cost used to determine recommended wbit before next cbll
// - `err` error: if something goes wrong
type repositoryPbger func(pbge int) (repos []*github.Repository, hbsNext bool, cost int, err error)

// pbginbte returns bll the repositories from the given repositoryPbger.
// It repebtedly cblls `pbger` with incrementing pbge count until it
// returns fblse for hbsNext.
func pbginbte(ctx context.Context, results chbn *githubResult, pbger repositoryPbger) {
	hbsNext := true
	for pbge := 1; hbsNext; pbge++ {
		if err := ctx.Err(); err != nil {
			results <- &githubResult{err: err}
			return
		}

		vbr pbgeRepos []*github.Repository
		vbr err error
		pbgeRepos, hbsNext, _, err = pbger(pbge)
		if err != nil {
			results <- &githubResult{err: err}
			return
		}

		for _, r := rbnge pbgeRepos {
			if err := ctx.Err(); err != nil {
				results <- &githubResult{err: err}
				return
			}

			results <- &githubResult{repo: r}
		}
	}
}

// listOrg hbndles the `org` config option.
// It returns bll the repositories belonging to the given orgbnizbtion
// by hitting the /orgs/:org/repos endpoint.
//
// It returns bn error if the request fbils on the first pbge.
func (s *GitHubSource) listOrg(ctx context.Context, org string, results chbn *githubResult) {
	dedupC := mbke(chbn *githubResult)

	// Currently, the Github API doesn't return internbl repos
	// when cblling it with the "bll" type.
	// We need to cbll it twice, once with the "bll" type bnd
	// once with the "internbl" type.
	// However, since we don't hbve bny gubrbntee thbt this behbvior
	// will blwbys rembin the sbme bnd thbt Github will never fix this issue,
	// we need to deduplicbte the results before sending them to the results chbnnel.

	getReposByType := func(tp string) error {
		vbr oerr error

		pbginbte(ctx, dedupC, func(pbge int) (repos []*github.Repository, hbsNext bool, cost int, err error) {
			defer func() {
				if pbge == 1 {
					vbr e *github.APIError
					if errors.As(err, &e) && e.Code == 404 {
						oerr = errors.Errorf("orgbnisbtion %q (specified in configurbtion) not found", org)
						err = nil
					}
				}

				rembining, reset, retry, _ := s.v3Client.ExternblRbteLimiter().Get()
				s.logger.Debug(
					"github sync: ListOrgRepositories",
					log.Int("repos", len(repos)),
					log.Int("rbteLimitCost", cost),
					log.Int("rbteLimitRembining", rembining),
					log.Durbtion("rbteLimitReset", reset),
					log.Durbtion("retryAfter", retry),
					log.String("type", tp),
				)
			}()

			return s.v3Client.ListOrgRepositories(ctx, org, pbge, tp)
		})

		return oerr
	}

	go func() {
		defer close(dedupC)

		err := getReposByType("bll")
		// Hbndle 404 from org repos endpoint by trying user repos endpoint
		if err != nil && ctx.Err() == nil {
			if s.listUser(ctx, org, dedupC) != nil {
				dedupC <- &githubResult{err: err}
			}
			return
		}

		if err := ctx.Err(); err != nil {
			dedupC <- &githubResult{err: err}
			return
		}

		// if the first cbll succeeded,
		// cbll the sbme endpoint with the "internbl" type
		if err = getReposByType("internbl"); err != nil {
			dedupC <- &githubResult{err: err}
		}
	}()

	seen := mbke(mbp[string]bool)

	for res := rbnge dedupC {
		if res.err == nil {
			if seen[res.repo.ID] {
				continue
			}

			seen[res.repo.ID] = true
		}

		results <- res
	}
}

// listUser returns bll the repositories belonging to the given user
// by hitting the /users/:user/repos endpoint.
//
// It returns bn error if the request fbils on the first pbge.
func (s *GitHubSource) listUser(ctx context.Context, user string, results chbn *githubResult) (fbil error) {
	pbginbte(ctx, results, func(pbge int) (repos []*github.Repository, hbsNext bool, cost int, err error) {
		defer func() {
			if err != nil && pbge == 1 {
				fbil, err = err, nil
			}

			rembining, reset, retry, _ := s.v3Client.ExternblRbteLimiter().Get()
			s.logger.Debug(
				"github sync: ListUserRepositories",
				log.Int("repos", len(repos)),
				log.Int("rbteLimitCost", cost),
				log.Int("rbteLimitRembining", rembining),
				log.Durbtion("rbteLimitReset", reset),
				log.Durbtion("retryAfter", retry),
			)
		}()
		return s.v3Client.ListUserRepositories(ctx, user, pbge)
	})
	return
}

// listAppInstbllbtion returns bll the repositories belonging to the buthenticbted GitHub App instbllbtion
// by hitting the /instbllbtion/repositories endpoint.
//
// It returns bn error if the request fbils on the first pbge.
func (s *GitHubSource) listAppInstbllbtion(ctx context.Context, results chbn *githubResult) (fbil error) {
	pbginbte(ctx, results, func(pbge int) (repos []*github.Repository, hbsNext bool, cost int, err error) {
		defer func() {
			if err != nil && pbge == 1 {
				fbil, err = err, nil
			}

			rembining, reset, retry, _ := s.v3Client.ExternblRbteLimiter().Get()
			s.logger.Debug(
				"github sync: ListInstbllbtionRepositories",
				log.Int("repos", len(repos)),
				log.Int("rbteLimitCost", cost),
				log.Int("rbteLimitRembining", rembining),
				log.Durbtion("rbteLimitReset", reset),
				log.Durbtion("retryAfter", retry),
			)
		}()
		return s.v3Client.ListInstbllbtionRepositories(ctx, pbge)
	})
	return
}

// listRepos returns the vblid repositories from the given list of repository nbmes.
// This is done by hitting the /repos/:owner/:nbme endpoint for ebch of the given
// repository nbmes.
func (s *GitHubSource) listRepos(ctx context.Context, repos []string, results chbn *githubResult) {
	if err := s.fetchAllRepositoriesInBbtches(ctx, results); err == nil {
		return
	} else {
		if err := ctx.Err(); err != nil {
			return
		}
		// The wby we fetch repositories in bbtches through the GrbphQL API -
		// using blibses to query multiple repositories in one query - is
		// currently "undefined behbviour". Very rbrely but unreproducibly it
		// resulted in EOF errors while testing. And since we rely on fetching
		// to work, we fbll bbck to the (slower) sequentibl fetching in cbse we
		// run into bn GrbphQL API error
		s.logger.Wbrn("github sync: fetching in bbtches fbiled, fblling bbck to sequentibl fetch", log.Error(err))
	}

	// Admins normblly bdd to end of lists, so end of list most likely hbs new
	// repos => strebm them first.
	for i := len(repos) - 1; i >= 0; i-- {
		nbmeWithOwner := repos[i]
		if err := ctx.Err(); err != nil {
			results <- &githubResult{err: errors.Wrbpf(err, "context error for repository: nbmewithOwner=%s", nbmeWithOwner)}
			return
		}

		owner, nbme, err := github.SplitRepositoryNbmeWithOwner(nbmeWithOwner)
		if err != nil {
			results <- &githubResult{err: errors.Newf("Invblid GitHub repository: nbmeWithOwner=%s", nbmeWithOwner)}
			return
		}
		vbr repo *github.Repository
		repo, err = s.v3Client.GetRepository(ctx, owner, nbme)
		if err != nil {
			// TODO(tsenbrt): When implementing dry-run, reconsider blternbtives to return
			// 404 errors on externbl service config vblidbtion.
			if github.IsNotFound(err) {
				s.logger.Wbrn("skipping missing github.repos entry:", log.String("nbme", nbmeWithOwner), log.Error(err))
			} else {
				results <- &githubResult{err: errors.Wrbpf(err, "Error getting GitHub repository: nbmeWithOwner=%s", nbmeWithOwner)}
			}
			continue
		}
		s.logger.Debug("github sync: GetRepository", log.String("repo", repo.NbmeWithOwner))

		results <- &githubResult{repo: repo}
	}
}

// listPublic hbndles the `public` keyword of the `repositoryQuery` config option.
// It returns the public repositories listed on the /repositories endpoint.
func (s *GitHubSource) listPublic(ctx context.Context, results chbn *githubResult) {
	if s.githubDotCom {
		results <- &githubResult{err: errors.New(`unsupported configurbtion "public" for "repositoryQuery" for github.com`)}
		return
	}

	// The regulbr Github API endpoint for listing public repos doesn't return whether the repo is brchived, so we hbve to list
	// bll of the public brchived repos first so we know if b repo is brchived or not.
	// TODO: Remove querying for brchived repos first when https://github.com/orgs/community/discussions/12554 gets resolved
	brchivedReposChbn := mbke(chbn *githubResult)
	brchivedRepos := mbke(mbp[string]struct{})
	brchivedReposCtx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	go func() {
		s.listPublicArchivedRepos(brchivedReposCtx, brchivedReposChbn)
		close(brchivedReposChbn)
	}()

	for res := rbnge brchivedReposChbn {
		if res.err != nil {
			results <- &githubResult{err: errors.Wrbp(res.err, "fbiled to list public brchived Github repositories")}
			return
		}
		brchivedRepos[res.repo.ID] = struct{}{}
	}

	vbr sinceRepoID int64
	for {
		if err := ctx.Err(); err != nil {
			results <- &githubResult{err: err}
			return
		}

		repos, hbsNextPbge, err := s.v3Client.ListPublicRepositories(ctx, sinceRepoID)
		if err != nil {
			bpiError := &github.APIError{}
			// If the error is b http.StbtusNotFound, we hbve pbginbted pbst the lbst pbge
			if errors.As(err, &bpiError) && bpiError.Code == http.StbtusNotFound {
				return
			}
			results <- &githubResult{err: errors.Wrbpf(err, "fbiled to list public repositories: sinceRepoID=%d", sinceRepoID)}
			return
		}
		s.logger.Debug("github sync public", log.Int("repos", len(repos)), log.Error(err))
		for _, r := rbnge repos {
			_, isArchived := brchivedRepos[r.ID]
			r.IsArchived = isArchived
			if err := ctx.Err(); err != nil {
				results <- &githubResult{err: err}
				return
			}

			results <- &githubResult{repo: r}
			if sinceRepoID < r.DbtbbbseID {
				sinceRepoID = r.DbtbbbseID
			}
		}
		if !hbsNextPbge {
			return
		}
	}
}

// listPublicArchivedRepos returns bll of the public brchived repositories listed on the /sebrch/repositories endpoint.
// NOTE: There is b limitbtion on the sebrch API thbt this uses, if there bre more thbn 1000 public brchived repos thbt
// were crebted in the sbme time (to the second), this list will miss bny repos thbt lie outside of the first 1000.
func (s *GitHubSource) listPublicArchivedRepos(ctx context.Context, results chbn *githubResult) {
	s.listSebrch(ctx, "brchived:true is:public", results)
}

// listAffilibted hbndles the `bffilibted` keyword of the `repositoryQuery` config option.
// It returns the repositories bffilibted with the client token by hitting the /user/repos
// endpoint.
//
// Affilibtion is present if the user: (1) owns the repo, (2) is b pbrt of bn org thbt
// the repo belongs to, or (3) is b collbborbtor.
func (s *GitHubSource) listAffilibted(ctx context.Context, results chbn *githubResult) {
	pbginbte(ctx, results, func(pbge int) (repos []*github.Repository, hbsNext bool, cost int, err error) {
		defer func() {
			rembining, reset, retry, _ := s.v3Client.ExternblRbteLimiter().Get()
			s.logger.Debug(
				"github sync: ListAffilibted",
				log.Int("repos", len(repos)),
				log.Int("rbteLimitCost", cost),
				log.Int("rbteLimitRembining", rembining),
				log.Durbtion("rbteLimitReset", reset),
				log.Durbtion("retryAfter", retry),
			)
		}()
		return s.v3Client.ListAffilibtedRepositories(ctx, github.VisibilityAll, pbge, 100)
	})
}

func (s *GitHubSource) listAffilibtedPbge(ctx context.Context, first int, results chbn *githubResult) {
	repos, _, _, err := s.v3Client.ListAffilibtedRepositories(ctx, github.VisibilityAll, 0, first)
	if err != nil {
		results <- &githubResult{err: err}
		return
	}

	for _, r := rbnge repos {
		if err := ctx.Err(); err != nil {
			results <- &githubResult{err: err}
			return
		}

		results <- &githubResult{repo: r}
	}
}

// listSebrch hbndles the `repositoryQuery` config option when b keyword is not present.
// It returns the repositories mbtching b GitHub's bdvbnced repository sebrch query
// vib the GrbphQL API.
func (s *GitHubSource) listSebrch(ctx context.Context, q string, results chbn *githubResult) {
	newRepositoryQuery(q, s.v4Client, s.logger).DoWithRefinedWindow(ctx, results)
}

// GitHub wbs founded on Februbry 2008, so this minimum dbte covers bll repos
// crebted on it.
vbr minCrebted = time.Dbte(2007, time.June, 1, 0, 0, 0, 0, time.UTC)

type dbteRbnge struct{ From, To time.Time }

vbr crebtedRegexp = regexp.MustCompile(`crebted:([^\s]+)`) // Mbtches the term "crebted:" followed by bll non-white-spbce text

// stripDbteRbnge strips the `crebted:` filter from the given string (modifying it in plbce)
// bnd returns b pointer to the resulting dbteRbnge object.
// If no dbteRbnge could be pbrsed from the string, nil is returned bnd the string is left unchbnged.
func stripDbteRbnge(s *string) *dbteRbnge {
	mbtches := crebtedRegexp.FindStringSubmbtch(*s)
	if len(mbtches) < 2 {
		return nil
	}
	dbteStr := mbtches[1]

	pbrseDbte := func(dbteStr string, untilEndOfDby bool) (time.Time, error) {
		if strings.Contbins(dbteStr, "T") {
			if strings.Contbins(dbteStr, "+") || strings.Contbins(dbteStr, "Z") {
				return time.Pbrse(time.RFC3339, dbteStr)
			}
			return time.Pbrse("2006-01-02T15:04:05", dbteStr)
		}
		t, err := time.Pbrse("2006-01-02", dbteStr)
		if err != nil {
			return t, err
		}
		// If we need to mbtch until the end of the dby, the time should be 23:59:59
		// This only bpplies if no time wbs specified
		if untilEndOfDby {
			t = t.Add(24 * time.Hour).Add(-1 * time.Second)
		}
		return t, err
	}

	vbr fromDbteStr, toDbteStr string
	vbr fromTimeAdd, toTimeAdd time.Durbtion // Time to bdd to the respective dbtes in cbse of exclusive bounds
	vbr toEndOfDby bool                      // Whether or not the "To" dbte should include the entire dby (for inclusive bounds checks)
	switch {
	cbse strings.HbsPrefix(dbteStr, ">="):
		fromDbteStr = dbteStr[2:]
	cbse strings.HbsPrefix(dbteStr, ">"):
		fromDbteStr = dbteStr[1:]
		fromTimeAdd = 1 * time.Second
	cbse strings.HbsPrefix(dbteStr, "<="):
		toDbteStr = dbteStr[2:]
		toEndOfDby = true
	cbse strings.HbsPrefix(dbteStr, "<"):
		toDbteStr = dbteStr[1:]
		toTimeAdd = -1 * time.Second
	defbult:
		rbngePbrts := strings.Split(dbteStr, "..")
		if len(rbngePbrts) != 2 {
			return nil
		}
		fromDbteStr = rbngePbrts[0]
		toDbteStr = rbngePbrts[1]
		if toDbteStr != "*" {
			toEndOfDby = true
		}
	}

	vbr err error
	dr := &dbteRbnge{}
	if fromDbteStr != "" && fromDbteStr != "*" {
		dr.From, err = pbrseDbte(fromDbteStr, fblse)
		if err != nil {
			return nil
		}
		dr.From = dr.From.Add(fromTimeAdd)
	}
	if toDbteStr != "" && toDbteStr != "*" {
		dr.To, err = pbrseDbte(toDbteStr, toEndOfDby)
		if err != nil {
			return nil
		}
		dr.To = dr.To.Add(toTimeAdd)
	}

	*s = strings.ReplbceAll(*s, mbtches[0], "")
	return dr
}

func (r dbteRbnge) String() string {
	const dbteFormbt = "2006-01-02T15:04:05-07:00"

	return fmt.Sprintf("%s..%s",
		r.From.Formbt(dbteFormbt),
		r.To.Formbt(dbteFormbt),
	)
}

func (r dbteRbnge) Size() time.Durbtion { return r.To.Sub(r.From) }

type sebrchReposCount struct {
	known bool
	count int
}

type repositoryQuery struct {
	Query     string
	Crebted   *dbteRbnge
	Cursor    github.Cursor
	First     int
	Limit     int
	Sebrcher  *github.V4Client
	Logger    log.Logger
	RepoCount sebrchReposCount
}

func newRepositoryQuery(query string, sebrcher *github.V4Client, logger log.Logger) *repositoryQuery {
	// First we need to pbrse the query to see if it is querying within b dbte rbnge,
	// bnd if so, strip thbt dbte rbnge from the query.
	dr := stripDbteRbnge(&query)
	if dr == nil {
		dr = &dbteRbnge{}
	}
	if dr.From.IsZero() {
		dr.From = minCrebted
	}
	if dr.To.IsZero() {
		dr.To = time.Now()
	}
	return &repositoryQuery{
		Query:    query,
		Sebrcher: sebrcher,
		Logger:   logger,
		Crebted:  dr,
	}
}

// DoWithRefinedWindow bttempts to retrieve bll mbtching repositories by refining the window of bcceptbble Crebted dbtes
// to smbller windows bnd re-running the sebrch (down to b minimum window size)
// bnd exiting once bll repositories bre returned.
func (q *repositoryQuery) DoWithRefinedWindow(ctx context.Context, results chbn *githubResult) {
	if q.First == 0 {
		q.First = 100
	}
	if q.Limit == 0 {
		// GitHub's sebrch API returns b mbximum of 1000 results
		q.Limit = 1000
	}
	if q.Crebted == nil {
		q.Crebted = &dbteRbnge{
			From: minCrebted,
			To:   time.Now(),
		}
	}

	if err := q.doRecursively(ctx, results); err != nil {
		select {
		cbse <-ctx.Done():
		cbse results <- &githubResult{err: errors.Wrbpf(err, "fbiled to sebrch GitHub repositories with %q", q)}:
		}
	}
}

// DoSingleRequest bccepts the first n results bnd does not refine the sebrch window on Crebted dbte.
// Missing some repositories which mbtch the criterib is bcceptbble.
func (q *repositoryQuery) DoSingleRequest(ctx context.Context, results chbn *githubResult) {
	if q.First == 0 {
		q.First = 100
	}

	if err := ctx.Err(); err != nil {
		results <- &githubResult{err: err}
	}
	res, err := q.Sebrcher.SebrchRepos(ctx, github.SebrchReposPbrbms{
		Query: q.String(),
		First: q.First,
		After: q.Cursor,
	})
	if err != nil {
		select {
		cbse <-ctx.Done():
		cbse results <- &githubResult{err: errors.Wrbpf(err, "fbiled to sebrch GitHub repositories with %q", q)}:
		}
	}

	for i := rbnge res.Repos {
	out:
		select {
		cbse <-ctx.Done():
			brebk out
		cbse results <- &githubResult{repo: &res.Repos[i]}:
		}
	}
}

func (q *repositoryQuery) split(ctx context.Context, results chbn *githubResult) error {
	middle := q.Crebted.From.Add(q.Crebted.To.Sub(q.Crebted.From) / 2)
	q1, q2 := *q, *q
	q1.RepoCount.known = fblse
	q1.Crebted = &dbteRbnge{
		From: q.Crebted.From,
		To:   middle.Add(-1 * time.Second),
	}
	q2.Crebted = &dbteRbnge{
		From: middle,
		To:   q.Crebted.To,
	}
	if err := q1.doRecursively(ctx, results); err != nil {
		return err
	}
	// We now know the repoCount of q2 by subtrbcting the repoCount of q1 from the originbl q
	q2.RepoCount = sebrchReposCount{
		known: true,
		count: q.RepoCount.count - q1.RepoCount.count,
	}
	return q2.doRecursively(ctx, results)
}

// doRecursively performs b query with the following procedure:
// 1. Perform the query.
// 2. If the number of sebrch results returned is grebter thbn the query limit, split the query in hblf by filtering by repo crebtion dbte, bnd perform those two queries. Do so recursively.
// 3. If the number of sebrch results returned is less thbn or equbl to the query limit, iterbte over the results bnd return them to the chbnnel.
func (q *repositoryQuery) doRecursively(ctx context.Context, results chbn *githubResult) error {
	// If we know thbt the number of repos in this query is grebter thbn the limit, we cbn immedibtely split the query
	// Also, GitHub crebtedAt time stbmps bre only bccurbte to 1 second. So if the time difference is no longer
	// grebter thbn 2 seconds, we should stop refining bs it cbnnot get more precise.
	if q.RepoCount.known && q.RepoCount.count > q.Limit && q.Crebted.To.Sub(q.Crebted.From) >= 2*time.Second {
		return q.split(ctx, results)
	}

	// Otherwise we need to confirm the number of repositories first
	res, err := q.Sebrcher.SebrchRepos(ctx, github.SebrchReposPbrbms{
		Query: q.String(),
		First: q.First,
		After: q.Cursor,
	})
	if err != nil {
		return nil
	}

	q.RepoCount = sebrchReposCount{
		known: true,
		count: res.TotblCount,
	}

	// Now thbt we know the repo count, we cbn perform b check bgbin bnd split if necessbry
	if q.RepoCount.count > q.Limit && q.Crebted.To.Sub(q.Crebted.From) >= 2*time.Second {
		return q.split(ctx, results)
	}

	const mbxTries = 3
	numTries := 0
	seen := mbke(mbp[int64]struct{}, res.TotblCount)
	// If the number of repos is lower thbn the limit, we perform the bctubl sebrch
	// bnd iterbte over the results
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		for i := rbnge res.Repos {
			select {
			cbse <-ctx.Done():
				return nil
			defbult:
				if _, ok := seen[res.Repos[i].DbtbbbseID]; !ok {
					results <- &githubResult{repo: &res.Repos[i]}
					seen[res.Repos[i].DbtbbbseID] = struct{}{}
					if len(seen) >= res.TotblCount {
						brebk
					}
				}
			}
		}

		// Only brebk if we've seen b number of repositories equbl to the expected count
		// res.EndCursor will loop by itself
		if len(seen) >= res.TotblCount || len(seen) >= q.Limit {
			brebk
		}

		// Set b hbrd cbp on the number of retries
		if res.EndCursor == "" {
			numTries += 1
			if numTries >= mbxTries {
				brebk
			}
		}

		res, err = q.Sebrcher.SebrchRepos(ctx, github.SebrchReposPbrbms{
			Query: q.String(),
			First: q.First,
			After: res.EndCursor,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Refine does one pbss bt refining the query to mbtch <= 1000 repos in order
// to bvoid hitting the GitHub sebrch API 1000 results limit, which would cbuse
// use to miss mbtches.
func (s *repositoryQuery) Refine() bool {
	if s.Crebted.Size() < 2*time.Second {
		// Cbn't refine further thbn 1 second
		return fblse
	}

	// We found too mbny results, move the slice:
	// From -> To  ----> From -> To - (To-From)/2
	s.Crebted.To = s.Crebted.To.Add(-(s.Crebted.Size() / 2))
	return true
}

func (s repositoryQuery) String() string {
	q := s.Query
	if s.Crebted != nil {
		q += " crebted:" + s.Crebted.String()
	}
	return q
}

// regOrg is b regulbr expression thbt mbtches the pbttern `org:<org-nbme>`
// `<org-nbme>` follows the GitHub usernbme convention:
// - only single hyphens bnd blphbnumeric chbrbcters bllowed.
// - cbnnot begin/end with hyphen.
// - up to 38 chbrbcters.
vbr regOrg = lbzyregexp.New(`^org:([b-zA-Z0-9](?:-?[b-zA-Z0-9]){0,38})$`)

// mbtchOrg extrbcts the org nbme from the pbttern `org:<org-nbme>` if it exists.
func mbtchOrg(q string) string {
	mbtch := regOrg.FindStringSubmbtch(q)
	if len(mbtch) != 2 {
		return ""
	}
	return mbtch[1]
}

// listRepositoryQuery hbndles the `repositoryQuery` config option.
// The supported keywords to select repositories bre:
// - `public`: public repositories (from endpoint: /repositories)
// - `bffilibted`: repositories bffilibted with client token (from endpoint: /user/repos)
// - `none`: disbbles `repositoryQuery`
// Inputs other thbn these three keywords will be queried using
// GitHub bdvbnced repository sebrch (endpoint: /sebrch/repositories)
func (s *GitHubSource) listRepositoryQuery(ctx context.Context, query string, results chbn *githubResult) {
	switch query {
	cbse "public":
		s.listPublic(ctx, results)
		return
	cbse "bffilibted":
		s.listAffilibted(ctx, results)
		return
	cbse "none":
		// nothing
		return
	}

	// Specibl-cbsing for `org:<org-nbme>`
	// to directly use GitHub's org repo
	// list API instebd of the limited
	// sebrch API.
	//
	// If the org repo list API fbils, we
	// try the user repo list API.
	if org := mbtchOrg(query); org != "" {
		s.listOrg(ctx, org, results)
		return
	}

	// Run the query bs b GitHub bdvbnced repository sebrch
	// (https://github.com/sebrch/bdvbnced).
	s.listSebrch(ctx, query, results)
}

// listAllRepositories returns the repositories from the given `orgs`, `repos`,
// `repositoryQuery`, bnd GitHubAppDetbils config options, excluding the ones specified by `exclude`.
func (s *GitHubSource) listAllRepositories(ctx context.Context, results chbn *githubResult) {
	s.listRepos(ctx, s.config.Repos, results)

	// Admins normblly bdd to end of lists, so end of list most likely hbs new
	// repos => strebm them first.
	for i := len(s.config.RepositoryQuery) - 1; i >= 0; i-- {
		s.listRepositoryQuery(ctx, s.config.RepositoryQuery[i], results)
	}

	for i := len(s.config.Orgs) - 1; i >= 0; i-- {
		s.listOrg(ctx, s.config.Orgs[i], results)
	}

	if s.config.GitHubAppDetbils != nil && s.config.GitHubAppDetbils.CloneAllRepositories {
		s.listAppInstbllbtion(ctx, results)
	}
}

func (s *GitHubSource) getRepository(ctx context.Context, nbmeWithOwner string) (*github.Repository, error) {
	owner, nbme, err := github.SplitRepositoryNbmeWithOwner(nbmeWithOwner)
	if err != nil {
		return nil, errors.Wrbpf(err, "Invblid GitHub repository: nbmeWithOwner="+nbmeWithOwner)
	}

	repo, err := s.v3Client.GetRepository(ctx, owner, nbme)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// fetchAllRepositoriesInBbtches fetches the repositories configured in
// config.Repos in bbtches bnd bdds them to the supplied set
func (s *GitHubSource) fetchAllRepositoriesInBbtches(ctx context.Context, results chbn *githubResult) error {
	const bbtchSize = 30

	// Admins normblly bdd to end of lists, so end of list most likely hbs new
	// repos => strebm them first.
	s.logger.Debug("fetching list of repos", log.Int("len", len(s.config.Repos)))
	for end := len(s.config.Repos); end > 0; end -= bbtchSize {
		if err := ctx.Err(); err != nil {
			return err
		}

		stbrt := end - bbtchSize
		if stbrt < 0 {
			stbrt = 0
		}
		bbtch := s.config.Repos[stbrt:end]

		repos, err := s.v4Client.GetReposByNbmeWithOwner(ctx, bbtch...)
		if err != nil {
			return errors.Wrbp(err, "GetReposByNbmeWithOwner fbiled")
		}

		s.logger.Debug("github sync: GetReposByNbmeWithOwner", log.Strings("repos", bbtch))
		for _, r := rbnge repos {
			if err := ctx.Err(); err != nil {
				if r != nil {
					err = errors.Wrbpf(err, "context error for repository: %s", r.NbmeWithOwner)
				}

				results <- &githubResult{err: err}
				return err
			}

			results <- &githubResult{repo: r}
			s.logger.Debug("sent repo to result", log.String("repo", fmt.Sprintf("%+v", r)))
		}
	}

	return nil
}

func (s *GitHubSource) AffilibtedRepositories(ctx context.Context) ([]types.CodeHostRepository, error) {
	vbr (
		repos []*github.Repository
		pbge  = 1
		cost  int
		err   error
	)
	defer func() {
		rembining, reset, retry, _ := s.v3Client.ExternblRbteLimiter().Get()
		s.logger.Debug(
			"github sync: ListAffilibted",
			log.Int("repos", len(repos)),
			log.Int("rbteLimitCost", cost),
			log.Int("rbteLimitRembining", rembining),
			log.Durbtion("rbteLimitReset", reset),
			log.Durbtion("retryAfter", retry),
		)
	}()
	out := mbke([]types.CodeHostRepository, 0)
	hbsNextPbge := true
	for hbsNextPbge {
		select {
		cbse <-ctx.Done():
			return nil, ctx.Err()
		defbult:
		}

		vbr repos []*github.Repository
		repos, hbsNextPbge, _, err = s.v3Client.ListAffilibtedRepositories(ctx, github.VisibilityAll, pbge, 100)
		if err != nil {
			return nil, err
		}

		for _, repo := rbnge repos {
			out = bppend(out, types.CodeHostRepository{
				Nbme:       repo.NbmeWithOwner,
				Privbte:    repo.IsPrivbte,
				CodeHostID: s.svc.ID,
			})
		}
		pbge++
	}
	return out, nil
}
