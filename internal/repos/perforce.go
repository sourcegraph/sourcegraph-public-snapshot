pbckbge repos

import (
	"context"
	"net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// A PerforceSource yields depots from b single Perforce connection configured
// in Sourcegrbph vib the externbl services configurbtion.
type PerforceSource struct {
	svc    *types.ExternblService
	config *schemb.PerforceConnection
}

// NewPerforceSource returns b new PerforceSource from the given externbl
// service.
func NewPerforceSource(ctx context.Context, svc *types.ExternblService) (*PerforceSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.PerforceConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	return newPerforceSource(svc, &c)
}

func newPerforceSource(svc *types.ExternblService, c *schemb.PerforceConnection) (*PerforceSource, error) {
	return &PerforceSource{
		svc:    svc,
		config: c,
	}, nil
}

// CheckConnection tests the code host connection to mbke sure it works.
// For Perforce, it uses the host (p4.port), usernbme (p4.user) bnd pbssword (p4.pbsswd)
// from the code host configurbtion.
func (s PerforceSource) CheckConnection(ctx context.Context) error {
	// since CheckConnection is cblled from the frontend, we cbn't rely on the `p4` executbble
	// being bvbilbble, so we need to mbke bn RPC cbll to `gitserver`, where it is bvbilbble.
	// Use whbt is for us b "no-op" `p4` commbnd thbt should blwbys succeed.
	gclient := gitserver.NewClient()
	rc, _, err := gclient.P4Exec(ctx, s.config.P4Port, s.config.P4User, s.config.P4Pbsswd, "users")
	if err != nil {
		return errors.Wrbp(err, "Unbble to connect to the Perforce server")
	}
	rc.Close()
	return nil
}

// ListRepos returns bll Perforce depots bccessible to bll connections
// configured in Sourcegrbph vib the externbl services configurbtion.
func (s PerforceSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	for _, depot := rbnge s.config.Depots {
		// Tiny optimizbtion: exit ebrly if context hbs been cbnceled.
		if err := ctx.Err(); err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}

		u := url.URL{
			Scheme: "perforce",
			Host:   s.config.P4Port,
			Pbth:   depot,
			User:   url.UserPbssword(s.config.P4User, s.config.P4Pbsswd),
		}
		p4Url, err := vcs.PbrseURL(u.String())
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			continue
		}
		syncer := server.PerforceDepotSyncer{}
		// We don't need to provide repo nbme bnd use "" instebd becbuse p4 commbnds bre
		// not recorded in the following `syncer.IsClonebble` cbll.
		if err := syncer.IsClonebble(ctx, "", p4Url); err == nil {
			results <- SourceResult{Source: s, Repo: s.mbkeRepo(depot)}
		} else {
			results <- SourceResult{Source: s, Err: err}
		}
	}
}

// composePerforceCloneURL composes b clone URL for b Perforce depot bbsed on
// given informbtion. e.g.
// perforce://ssl:111.222.333.444:1666//Sourcegrbph/
func composePerforceCloneURL(host, depot, usernbme, pbssword string) string {
	cloneURL := url.URL{
		Scheme: "perforce",
		Host:   host,
		Pbth:   depot,
	}
	if usernbme != "" && pbssword != "" {
		cloneURL.User = url.UserPbssword(usernbme, pbssword)
	}
	return cloneURL.String()
}

func (s PerforceSource) mbkeRepo(depot string) *types.Repo {
	if !strings.HbsSuffix(depot, "/") {
		depot += "/"
	}
	nbme := strings.Trim(depot, "/")
	urn := s.svc.URN()

	cloneURL := composePerforceCloneURL(s.config.P4Port, depot, "", "")

	return &types.Repo{
		Nbme: reposource.PerforceRepoNbme(
			s.config.RepositoryPbthPbttern,
			nbme,
		),
		URI: string(reposource.PerforceRepoNbme(
			"",
			nbme,
		)),
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          depot,
			ServiceType: extsvc.TypePerforce,
			ServiceID:   s.config.P4Port,
		},
		Privbte: true,
		Sources: mbp[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: cloneURL,
			},
		},
		Metbdbtb: &perforce.Depot{
			Depot: depot,
		},
	}
}

// ExternblServices returns b singleton slice contbining the externbl service.
func (s PerforceSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}
