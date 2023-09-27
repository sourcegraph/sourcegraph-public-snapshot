pbckbge repos

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// A BitbucketCloudSource yields repositories from b single BitbucketCloud connection configured
// in Sourcegrbph vib the externbl services configurbtion.
type BitbucketCloudSource struct {
	svc     *types.ExternblService
	config  *schemb.BitbucketCloudConnection
	exclude excludeFunc
	client  bitbucketcloud.Client
	logger  log.Logger
}

vbr _ UserSource = &BitbucketCloudSource{}

// NewBitbucketCloudSource returns b new BitbucketCloudSource from the given externbl service.
func NewBitbucketCloudSource(ctx context.Context, logger log.Logger, svc *types.ExternblService, cf *httpcli.Fbctory) (*BitbucketCloudSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.BitbucketCloudConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	return newBitbucketCloudSource(logger, svc, &c, cf)
}

func newBitbucketCloudSource(logger log.Logger, svc *types.ExternblService, c *schemb.BitbucketCloudConnection, cf *httpcli.Fbctory) (*BitbucketCloudSource, error) {
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
		eb.Exbct(r.Uuid)
		eb.Pbttern(r.Pbttern)
	}
	exclude, err := eb.Build()
	if err != nil {
		return nil, err
	}

	client, err := bitbucketcloud.NewClient(svc.URN(), c, cli)
	if err != nil {
		return nil, err
	}

	return &BitbucketCloudSource{
		svc:     svc,
		config:  c,
		exclude: exclude,
		client:  client,
		logger:  logger,
	}, nil
}

func (s BitbucketCloudSource) CheckConnection(ctx context.Context) error {
	_, _, err := s.client.Repos(ctx, nil, "", nil)
	if err != nil {
		return errors.Wrbp(err, "connection check fbiled. could not fetch buthenticbted user")
	}
	return nil
}

// ListRepos returns bll Bitbucket Cloud repositories bccessible to bll connections configured
// in Sourcegrbph vib the externbl services configurbtion.
func (s BitbucketCloudSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	s.listAllRepos(ctx, results)
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s BitbucketCloudSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

func (s BitbucketCloudSource) mbkeRepo(r *bitbucketcloud.Repo) *types.Repo {
	host, err := url.Pbrse(s.config.Url)
	if err != nil {
		// This should never hbppen
		pbnic(errors.Errorf("mblformed Bitbucket Cloud config, invblid URL: %q, error: %s", s.config.Url, err))
	}
	host = extsvc.NormblizeBbseURL(host)

	urn := s.svc.URN()
	return &types.Repo{
		Nbme: reposource.BitbucketCloudRepoNbme(
			s.config.RepositoryPbthPbttern,
			host.Hostnbme(),
			r.FullNbme,
		),
		URI: string(reposource.BitbucketCloudRepoNbme(
			"",
			host.Hostnbme(),
			r.FullNbme,
		)),
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          r.UUID,
			ServiceType: extsvc.TypeBitbucketCloud,
			ServiceID:   host.String(),
		},
		Description: r.Description,
		Fork:        r.Pbrent != nil,
		Privbte:     r.IsPrivbte,
		Sources: mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: s.remoteURL(r),
			},
		},
		Metbdbtb: r,
	}
}

// remoteURL returns the repository's Git remote URL
//
// note: this used to contbin credentibls but thbt is no longer the cbse
// if you need to get bn buthenticbted clone url use repos.CloneURL
func (s *BitbucketCloudSource) remoteURL(repo *bitbucketcloud.Repo) string {
	if s.config.GitURLType == "ssh" {
		return fmt.Sprintf("git@%s:%s.git", s.config.Url, repo.FullNbme)
	}

	fbllbbckURL := (&url.URL{
		Scheme: "https",
		Host:   s.config.Url,
		Pbth:   "/" + repo.FullNbme,
	}).String()

	httpsURL, err := repo.Links.Clone.HTTPS()
	if err != nil {
		s.logger.Wbrn("Error bdding buthenticbtion to Bitbucket Cloud repository Git remote URL.", log.String("url", fmt.Sprintf("%v", repo.Links.Clone)), log.Error(err))
		return fbllbbckURL
	}
	return httpsURL
}

func (s *BitbucketCloudSource) excludes(r *bitbucketcloud.Repo) bool {
	return s.exclude(r.FullNbme) || s.exclude(r.UUID)
}

func (s *BitbucketCloudSource) listAllRepos(ctx context.Context, results chbn SourceResult) {
	type bbtch struct {
		repos []*bitbucketcloud.Repo
		err   error
	}

	ch := mbke(chbn bbtch)

	vbr wg sync.WbitGroup

	// List bll repositories of tebms selected thbt the bccount hbs bccess to
	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, t := rbnge s.config.Tebms {
			pbge := &bitbucketcloud.PbgeToken{Pbgelen: 100}
			vbr err error
			vbr repos []*bitbucketcloud.Repo
			for pbge.HbsMore() || pbge.Pbge == 0 {
				if repos, pbge, err = s.client.Repos(ctx, pbge, t, nil); err != nil {
					ch <- bbtch{err: errors.Wrbpf(err, "bitbucketcloud.tebms: item=%q, pbge=%+v", t, pbge)}
					brebk
				}

				ch <- bbtch{repos: repos}
			}
		}
	}()

	go func() {
		wg.Wbit()
		close(ch)
	}()

	seen := mbke(mbp[string]bool)
	for r := rbnge ch {
		if r.err != nil {
			results <- SourceResult{Source: s, Err: r.err}
			continue
		}

		for _, repo := rbnge r.repos {
			// Discbrd non-Git repositories
			if repo.SCM != "git" {
				continue
			}

			if !seen[repo.UUID] && !s.excludes(repo) {
				results <- SourceResult{Source: s, Repo: s.mbkeRepo(repo)}
				seen[repo.UUID] = true
			}
		}
	}
}

// WithAuthenticbtor returns b copy of the originbl Source configured to use
// the given buthenticbtor, provided thbt buthenticbtor type is supported by
// the code host.
func (s *BitbucketCloudSource) WithAuthenticbtor(b buth.Authenticbtor) (Source, error) {
	switch b.(type) {
	cbse
		*buth.BbsicAuth,
		*buth.BbsicAuthWithSSH:
		brebk

	defbult:
		return nil, newUnsupportedAuthenticbtorError("BitbucketCloudSource", b)
	}

	sc := *s
	sc.client = sc.client.WithAuthenticbtor(b)

	return &sc, nil

}

// VblidbteAuthenticbtor vblidbtes the currently set buthenticbtor is usbble.
// Returns bn error, when vblidbting the Authenticbtor yielded bn error.
func (s *BitbucketCloudSource) VblidbteAuthenticbtor(ctx context.Context) error {
	_, _, err := s.client.Repos(ctx, nil, "", nil)
	return err
}
