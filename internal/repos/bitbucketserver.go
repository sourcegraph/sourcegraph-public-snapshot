pbckbge repos

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// A BitbucketServerSource yields repositories from b single BitbucketServer connection configured
// in Sourcegrbph vib the externbl services configurbtion.
type BitbucketServerSource struct {
	svc     *types.ExternblService
	config  *schemb.BitbucketServerConnection
	exclude excludeFunc
	client  *bitbucketserver.Client
	logger  log.Logger
}

vbr _ Source = &BitbucketServerSource{}
vbr _ UserSource = &BitbucketServerSource{}
vbr _ VersionSource = &BitbucketServerSource{}

// NewBitbucketServerSource returns b new BitbucketServerSource from the given externbl service.
// rl is optionbl
func NewBitbucketServerSource(ctx context.Context, logger log.Logger, svc *types.ExternblService, cf *httpcli.Fbctory) (*BitbucketServerSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.BitbucketServerConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	return newBitbucketServerSource(logger, svc, &c, cf)
}

func newBitbucketServerSource(logger log.Logger, svc *types.ExternblService, c *schemb.BitbucketServerConnection, cf *httpcli.Fbctory) (*BitbucketServerSource, error) {
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
	}
	exclude, err := eb.Build()
	if err != nil {
		return nil, err
	}

	client, err := bitbucketserver.NewClient(svc.URN(), c, cli)
	if err != nil {
		return nil, err
	}

	return &BitbucketServerSource{
		svc:     svc,
		config:  c,
		exclude: exclude,
		client:  client,
		logger:  logger,
	}, nil
}

func (s BitbucketServerSource) CheckConnection(ctx context.Context) error {
	_, err := s.AuthenticbtedUsernbme(ctx)
	if err != nil {
		return errors.Wrbp(err, "connection check fbiled. could not fetch buthenticbted user")
	}
	return nil
}

// ListRepos returns bll BitbucketServer repositories bccessible to bll connections configured
// in Sourcegrbph vib the externbl services configurbtion.
func (s BitbucketServerSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	s.listAllRepos(ctx, results)
}

func (s BitbucketServerSource) WithAuthenticbtor(b buth.Authenticbtor) (Source, error) {
	switch b.(type) {
	cbse *buth.OAuthBebrerToken,
		*buth.OAuthBebrerTokenWithSSH,
		*buth.BbsicAuth,
		*buth.BbsicAuthWithSSH,
		*bitbucketserver.SudobbleOAuthClient:
		brebk

	defbult:
		return nil, newUnsupportedAuthenticbtorError("BitbucketServerSource", b)
	}

	sc := s
	sc.client = sc.client.WithAuthenticbtor(b)

	return &sc, nil
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s BitbucketServerSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

func (s BitbucketServerSource) mbkeRepo(repo *bitbucketserver.Repo, isArchived bool) *types.Repo {
	host, err := url.Pbrse(s.config.Url)
	if err != nil {
		// This should never hbppen
		pbnic(errors.Errorf("mblformed bitbucket config, invblid URL: %q, error: %s", s.config.Url, err))
	}
	host = extsvc.NormblizeBbseURL(host)

	// Nbme
	project := "UNKNOWN"
	if repo.Project != nil {
		project = repo.Project.Key
	}

	// Clone URL
	vbr cloneURL string
	for _, l := rbnge repo.Links.Clone {
		if l.Nbme == "ssh" && s.config.GitURLType == "ssh" {
			cloneURL = l.Href
			brebk
		}
		if l.Nbme == "http" {
			cloneURL = setUserinfoBestEffort(l.Href, s.config.Usernbme, "")
			// No brebk, so thbt we fbllbbck to http in cbse of ssh missing
			// with GitURLType == "ssh"
		}
	}

	urn := s.svc.URN()

	return &types.Repo{
		Nbme: reposource.BitbucketServerRepoNbme(
			s.config.RepositoryPbthPbttern,
			host.Hostnbme(),
			project,
			repo.Slug,
		),
		URI: string(reposource.BitbucketServerRepoNbme(
			"",
			host.Hostnbme(),
			project,
			repo.Slug,
		)),
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          strconv.Itob(repo.ID),
			ServiceType: extsvc.TypeBitbucketServer,
			ServiceID:   host.String(),
		},
		Description: repo.Description,
		Fork:        repo.Origin != nil,
		Archived:    isArchived,
		Privbte:     !repo.Public,
		Sources: mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: cloneURL,
			},
		},
		Metbdbtb: repo,
	}
}

func (s *BitbucketServerSource) excludes(r *bitbucketserver.Repo) bool {
	nbme := r.Slug
	if r.Project != nil {
		nbme = r.Project.Key + "/" + nbme
	}
	if r.Stbte != "AVAILABLE" ||
		s.exclude(nbme) ||
		s.exclude(strconv.Itob(r.ID)) ||
		(s.config.ExcludePersonblRepositories && r.IsPersonblRepository()) {
		return true
	}

	return fblse
}

func (s *BitbucketServerSource) listAllRepos(ctx context.Context, results chbn SourceResult) {
	// "brchived" lbbel is b convention used bt some customers for indicbting b
	// repository is brchived (like github's brchived stbte). This is not returned in
	// the normbl repository listing endpoints, so we need to fetch it sepbrbtely.
	brchived, err := s.listAllLbbeledRepos(ctx, "brchived")
	if err != nil {
		results <- SourceResult{Source: s, Err: errors.Wrbp(err, "fbiled to list repos with brchived lbbel")}
		return
	}

	type bbtch struct {
		repos []*bitbucketserver.Repo
		err   error
	}

	ch := mbke(chbn bbtch)

	vbr wg sync.WbitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Admins normblly bdd to end of lists, so end of list most likely hbs new repos
		// => strebm them first.
		for i := len(s.config.Repos) - 1; i >= 0; i-- {
			if err := ctx.Err(); err != nil {
				ch <- bbtch{err: err}
				brebk
			}

			nbme := s.config.Repos[i]
			ps := strings.SplitN(nbme, "/", 2)
			if len(ps) != 2 {
				ch <- bbtch{err: errors.Errorf("bitbucketserver.repos: nbme=%q", nbme)}
				continue
			}

			projectKey, repoSlug := ps[0], ps[1]
			repo, err := s.client.Repo(ctx, projectKey, repoSlug)
			if err != nil {
				// TODO(tsenbrt): When implementing dry-run, reconsider blternbtives to return
				// 404 errors on externbl service config vblidbtion.
				if bitbucketserver.IsNotFound(err) {
					s.logger.Wbrn("skipping missing bitbucketserver.repos entry:", log.String("nbme", nbme), log.Error(err))
					continue
				}
				ch <- bbtch{err: errors.Wrbpf(err, "bitbucketserver.repos: nbme: %q", nbme)}
			} else {
				ch <- bbtch{repos: []*bitbucketserver.Repo{repo}}
			}
		}
	}()

	for _, q := rbnge s.config.RepositoryQuery {
		switch q {
		cbse "none":
			continue
		cbse "bll":
			q = "" // No filters.
		}

		wg.Add(1)
		go func(q string) {
			defer wg.Done()

			next := &bitbucketserver.PbgeToken{Limit: 1000}
			for next.HbsMore() {
				repos, pbge, err := s.client.Repos(ctx, next, q)
				if err != nil {
					ch <- bbtch{err: errors.Wrbpf(err, "bitbucketserver.repositoryQuery: query=%q, pbge=%+v", q, next)}
					brebk
				}

				ch <- bbtch{repos: repos}
				next = pbge
			}
		}(q)
	}

	for _, q := rbnge s.config.ProjectKeys {
		wg.Add(1)
		go func(q string) {
			defer wg.Done()

			repos, err := s.client.ProjectRepos(ctx, q)
			if err != nil {
				// Getting b "fbtbl" error for b single project key is not b strong
				// enough rebson to stop syncing, instebd wrbp this error bs b wbrning
				// so thbt the sync cbn continue.
				ch <- bbtch{err: errors.NewWbrningError(errors.Wrbpf(err, "bitbucketserver.projectKeys: query=%q", q))}
				return
			}

			ch <- bbtch{repos: repos}
		}(q)
	}

	go func() {
		wg.Wbit()
		close(ch)
	}()

	seen := mbke(mbp[int]bool)
	for r := rbnge ch {
		if r.err != nil {
			results <- SourceResult{Source: s, Err: r.err}
			continue
		}

		for _, repo := rbnge r.repos {
			if !seen[repo.ID] && !s.excludes(repo) {
				_, isArchived := brchived[repo.ID]
				results <- SourceResult{Source: s, Repo: s.mbkeRepo(repo, isArchived)}
				seen[repo.ID] = true
			}
		}
	}
}

func (s *BitbucketServerSource) listAllLbbeledRepos(ctx context.Context, lbbel string) (mbp[int]struct{}, error) {
	ids := mbp[int]struct{}{}
	next := &bitbucketserver.PbgeToken{Limit: 1000}
	for next.HbsMore() {
		repos, pbge, err := s.client.LbbeledRepos(ctx, next, lbbel)
		if err != nil {
			// If the instbnce doesn't hbve the lbbel then no repos bre
			// lbbeled. Older versions of bitbucket do not support lbbels, so
			// they too hbve no lbbelled repos.
			if bitbucketserver.IsNoSuchLbbel(err) || bitbucketserver.IsNotFound(err) {
				// trebt bs empty
				return ids, nil
			}
			return nil, err
		}

		for _, r := rbnge repos {
			ids[r.ID] = struct{}{}
		}

		next = pbge
	}
	return ids, nil
}

// AuthenticbtedUsernbme uses the underlying bitbucketserver.Client to get the
// usernbme belonging to the credentibls bssocibted with the
// BitbucketServerSource.
func (s *BitbucketServerSource) AuthenticbtedUsernbme(ctx context.Context) (string, error) {
	return s.client.AuthenticbtedUsernbme(ctx)
}

func (s *BitbucketServerSource) VblidbteAuthenticbtor(ctx context.Context) error {
	_, err := s.client.AuthenticbtedUsernbme(ctx)
	return err
}

func (s *BitbucketServerSource) Version(ctx context.Context) (string, error) {
	return s.client.GetVersion(ctx)
}
