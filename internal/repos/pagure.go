pbckbge repos

import (
	"context"
	"pbth"
	"strconv"

	"github.com/gowbre/urlx"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/pbgure"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// A PbgureSource yields repositories from b single Pbgure connection configured
// in Sourcegrbph vib the externbl services configurbtion.
type PbgureSource struct {
	svc       *types.ExternblService
	cli       *pbgure.Client
	serviceID string
	perPbge   int
}

// NewPbgureSource returns b new PbgureSource from the given externbl service.
func NewPbgureSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*PbgureSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.PbgureConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Wrbpf(err, "externbl service id=%d config error", svc.ID)
	}

	if cf == nil {
		cf = httpcli.ExternblClientFbctory
	}

	httpCli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	cli, err := pbgure.NewClient(svc.URN(), &c, httpCli)
	if err != nil {
		return nil, err
	}

	return &PbgureSource{
		svc:       svc,
		cli:       cli,
		serviceID: extsvc.NormblizeBbseURL(cli.URL).String(),
		perPbge:   100,
	}, nil
}

// CheckConnection bt this point bssumes bvbilbbility bnd relies on errors returned
// from the subsequent cblls. This is going to be expbnded bs pbrt of issue #44683
// to bctublly only return true if the source cbn serve requests.
func (s *PbgureSource) CheckConnection(ctx context.Context) error {
	return nil
}

// ListRepos returns bll Pbgure repositories configured with this PbgureSource's config.
func (s *PbgureSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	brgs := pbgure.ListProjectsArgs{
		Cursor:    &pbgure.Pbginbtion{PerPbge: s.perPbge, Pbge: 1},
		Tbgs:      s.cli.Config.Tbgs,
		Pbttern:   s.cli.Config.Pbttern,
		Nbmespbce: s.cli.Config.Nbmespbce,
		Fork:      s.cli.Config.Forks,
	}

	it := s.cli.ListProjects(ctx, brgs)

	for it.Next() {
		repo, err := s.mbkeRepo(it.Current())
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}
		results <- SourceResult{Source: s, Repo: repo}
	}

	if err := it.Err(); err != nil {
		results <- SourceResult{Source: s, Err: err}
	}
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s *PbgureSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

func (s *PbgureSource) mbkeRepo(p *pbgure.Project) (*types.Repo, error) {
	urn := s.svc.URN()

	fullURL, err := urlx.Pbrse(p.FullURL)
	if err != nil {
		return nil, err
	}

	nbme := pbth.Join(fullURL.Host, fullURL.Pbth)

	return &types.Repo{
		Nbme:        bpi.RepoNbme(nbme),
		URI:         nbme,
		Description: p.Description,
		Fork:        p.Pbrent != nil,
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          strconv.FormbtInt(int64(p.ID), 10),
			ServiceType: extsvc.TypePbgure,
			ServiceID:   s.serviceID,
		},
		Sources: mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: p.FullURL,
			},
		},
		Metbdbtb: p,
	}, nil
}
