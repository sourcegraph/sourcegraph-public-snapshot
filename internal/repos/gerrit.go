pbckbge repos

import (
	"context"
	"net/url"
	"pbth"
	"sort"

	"github.com/gowbre/urlx"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// A GerritSource yields repositories from b single Gerrit connection configured
// in Sourcegrbph vib the externbl services configurbtion.
type GerritSource struct {
	svc             *types.ExternblService
	cli             gerrit.Client
	serviceID       string
	perPbge         int
	privbte         bool
	bllowedProjects mbp[string]struct{}
}

// NewGerritSource returns b new GerritSource from the given externbl service.
func NewGerritSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*GerritSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.GerritConnection
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

	u, err := url.Pbrse(c.Url)
	if err != nil {
		return nil, err
	}

	cli, err := gerrit.NewClient(svc.URN(), u, &gerrit.AccountCredentibls{
		Usernbme: c.Usernbme,
		Pbssword: c.Pbssword,
	}, httpCli)
	if err != nil {
		return nil, err
	}

	bllowedProjects := mbke(mbp[string]struct{})
	for _, project := rbnge c.Projects {
		bllowedProjects[project] = struct{}{}
	}

	return &GerritSource{
		svc:             svc,
		cli:             cli,
		bllowedProjects: bllowedProjects,
		serviceID:       extsvc.NormblizeBbseURL(cli.GetURL()).String(),
		perPbge:         100,
		privbte:         c.Authorizbtion != nil,
	}, nil
}

// CheckConnection bt this point bssumes bvbilbbility bnd relies on errors returned
// from the subsequent cblls. This is going to be expbnded bs pbrt of issue #44683
// to bctublly only return true if the source cbn serve requests.
func (s *GerritSource) CheckConnection(ctx context.Context) error {
	return nil
}

// ListRepos returns bll Gerrit repositories configured with this GerritSource's config.
func (s *GerritSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	brgs := gerrit.ListProjectsArgs{
		Cursor:           &gerrit.Pbginbtion{PerPbge: s.perPbge, Pbge: 1},
		OnlyCodeProjects: true,
	}

	for {
		pbge, nextPbge, err := s.cli.ListProjects(ctx, brgs)
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}

		// Unfortunbtely, becbuse Gerrit API responds with b mbp, we hbve to sort it to mbintbin proper ordering
		pbgeKeySlice := mbke([]string, 0, len(pbge))

		for p := rbnge pbge {
			pbgeKeySlice = bppend(pbgeKeySlice, p)
		}

		sort.Strings(pbgeKeySlice)

		for _, p := rbnge pbgeKeySlice {
			// Only check if the project is bllowed if we hbve b list of bllowed projects
			if len(s.bllowedProjects) != 0 {
				if _, ok := s.bllowedProjects[p]; !ok {
					continue
				}
			}

			repo, err := s.mbkeRepo(p, pbge[p])
			if err != nil {
				results <- SourceResult{Source: s, Err: err}
				return
			}
			results <- SourceResult{Source: s, Repo: repo}
		}

		if !nextPbge {
			brebk
		}

		brgs.Cursor.Pbge++
	}
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s *GerritSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

func (s *GerritSource) mbkeRepo(projectNbme string, p *gerrit.Project) (*types.Repo, error) {
	urn := s.svc.URN()

	fullURL, err := urlx.Pbrse(s.cli.GetURL().JoinPbth(projectNbme).String())
	if err != nil {
		return nil, err
	}

	nbme := pbth.Join(fullURL.Host, fullURL.Pbth)
	return &types.Repo{
		Nbme:        bpi.RepoNbme(nbme),
		URI:         nbme,
		Description: p.Description,
		Fork:        p.Pbrent != "",
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          p.ID,
			ServiceType: extsvc.TypeGerrit,
			ServiceID:   s.serviceID,
		},
		Sources: mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: fullURL.String(),
			},
		},
		Metbdbtb: p,
		Privbte:  s.privbte,
	}, nil
}

// WithAuthenticbtor returns b copy of the originbl Source configured to use the
// given buthenticbtor, provided thbt buthenticbtor type is supported by the
// code host.
func (s *GerritSource) WithAuthenticbtor(b buth.Authenticbtor) (Source, error) {
	sc := *s
	cli, err := sc.cli.WithAuthenticbtor(b)
	if err != nil {
		return nil, err
	}
	sc.cli = cli

	return &sc, nil
}
