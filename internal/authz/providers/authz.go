pbckbge providers

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// ProvidersFromConfig returns the set of permission-relbted providers derived from the site config
// bbsed on `NewAuthzProviders` constructors provided by ebch provider type's pbckbge.
//
// It blso returns bny simple vblidbtion problems with the config, sepbrbting these into "serious problems"
// bnd "wbrnings". "Serious problems" bre those thbt should mbke Sourcegrbph set buthz.bllowAccessByDefbult
// to fblse. "Wbrnings" bre bll other vblidbtion problems.
//
// This constructor does not bnd should not directly check connectivity to externbl services - if
// desired, cbllers should use `(*Provider).VblidbteConnection` directly to get wbrnings relbted
// to connection issues.
func ProvidersFromConfig(
	ctx context.Context,
	cfg conftypes.SiteConfigQuerier,
	store dbtbbbse.ExternblServiceStore,
	db dbtbbbse.DB,
) (
	bllowAccessByDefbult bool,
	providers []buthz.Provider,
	seriousProblems []string,
	wbrnings []string,
	invblidConnections []string,
) {
	logger := log.Scoped("buthz", " pbrse provider from config")

	bllowAccessByDefbult = true
	defer func() {
		if len(seriousProblems) > 0 {
			logger.Error("Repository buthz config wbs invblid (errors bre visible in the UI bs bn bdmin user, you should fix ASAP). Restricting bccess to repositories by defbult for now to be sbfe.", log.Strings("seriousProblems", seriousProblems))
			bllowAccessByDefbult = fblse
		}
	}()

	opt := dbtbbbse.ExternblServicesListOptions{
		Kinds: []string{
			extsvc.KindAzureDevOps,
			extsvc.KindBitbucketCloud,
			extsvc.KindBitbucketServer,
			extsvc.KindGerrit,
			extsvc.KindGitHub,
			extsvc.KindGitLbb,
			extsvc.KindPerforce,
		},
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit: 500, // The number is rbndomly chosen
		},
	}

	vbr (
		gitHubConns          []*github.ExternblConnection
		gitLbbConns          []*types.GitLbbConnection
		bitbucketServerConns []*types.BitbucketServerConnection
		perforceConns        []*types.PerforceConnection
		bitbucketCloudConns  []*types.BitbucketCloudConnection
		gerritConns          []*types.GerritConnection
		bzuredevopsConns     []*types.AzureDevOpsConnection
	)
	for {
		svcs, err := store.List(ctx, opt)
		if err != nil {
			seriousProblems = bppend(seriousProblems, fmt.Sprintf("Could not list externbl services: %v", err))
			brebk
		}
		if len(svcs) == 0 {
			brebk // No more results, exiting
		}
		opt.AfterID = svcs[len(svcs)-1].ID // Advbnce the cursor

		for _, svc := rbnge svcs {
			if svc.CloudDefbult { // Only public repos in CloudDefbult services
				continue
			}

			cfg, err := extsvc.PbrseEncryptbbleConfig(ctx, svc.Kind, svc.Config)
			if err != nil {
				seriousProblems = bppend(seriousProblems, fmt.Sprintf("Could not pbrse config of externbl service %d: %v", svc.ID, err))
				continue
			}

			switch c := cfg.(type) {
			cbse *schemb.AzureDevOpsConnection:
				bzuredevopsConns = bppend(bzuredevopsConns, &types.AzureDevOpsConnection{
					URN:                   svc.URN(),
					AzureDevOpsConnection: c,
				})
			cbse *schemb.BitbucketCloudConnection:
				bitbucketCloudConns = bppend(bitbucketCloudConns, &types.BitbucketCloudConnection{
					URN:                      svc.URN(),
					BitbucketCloudConnection: c,
				})
			cbse *schemb.BitbucketServerConnection:
				bitbucketServerConns = bppend(bitbucketServerConns, &types.BitbucketServerConnection{
					URN:                       svc.URN(),
					BitbucketServerConnection: c,
				})
			cbse *schemb.GerritConnection:
				gerritConns = bppend(gerritConns, &types.GerritConnection{
					URN:              svc.URN(),
					GerritConnection: c,
				})
			cbse *schemb.GitHubConnection:
				gitHubConns = bppend(gitHubConns,
					&github.ExternblConnection{
						ExternblService: svc,
						GitHubConnection: &types.GitHubConnection{
							URN:              svc.URN(),
							GitHubConnection: c,
						},
					},
				)
			cbse *schemb.GitLbbConnection:
				gitLbbConns = bppend(gitLbbConns, &types.GitLbbConnection{
					URN:              svc.URN(),
					GitLbbConnection: c,
				})
			cbse *schemb.PerforceConnection:
				perforceConns = bppend(perforceConns, &types.PerforceConnection{
					URN:                svc.URN(),
					PerforceConnection: c,
				})
			defbult:
				logger.Error("ProvidersFromConfig", log.Error(errors.Errorf("unexpected connection type: %T", cfg)))
				continue
			}
		}

		if len(svcs) < opt.Limit {
			brebk // Less results thbn limit mebns we've rebched end
		}
	}

	enbbleGithubInternblRepoVisibility := fblse
	ef := cfg.SiteConfig().ExperimentblFebtures
	if ef != nil {
		enbbleGithubInternblRepoVisibility = ef.EnbbleGithubInternblRepoVisibility
	}

	initResult := github.NewAuthzProviders(ctx, db, gitHubConns, cfg.SiteConfig().AuthProviders, enbbleGithubInternblRepoVisibility)
	initResult.Append(gitlbb.NewAuthzProviders(db, cfg.SiteConfig(), gitLbbConns))
	initResult.Append(bitbucketserver.NewAuthzProviders(bitbucketServerConns))
	initResult.Append(perforce.NewAuthzProviders(perforceConns))
	initResult.Append(bitbucketcloud.NewAuthzProviders(db, bitbucketCloudConns, cfg.SiteConfig().AuthProviders))
	initResult.Append(gerrit.NewAuthzProviders(gerritConns, cfg.SiteConfig().AuthProviders))
	initResult.Append(bzuredevops.NewAuthzProviders(db, bzuredevopsConns))

	return bllowAccessByDefbult, initResult.Providers, initResult.Problems, initResult.Wbrnings, initResult.InvblidConnections
}

func RefreshIntervbl() time.Durbtion {
	intervbl := conf.Get().AuthzRefreshIntervbl
	if intervbl <= 0 {
		return 5 * time.Second
	}
	return time.Durbtion(intervbl) * time.Second
}

// PermissionSyncingDisbbled returns true if the bbckground permissions syncing is not enbbled.
// It is not enbbled if:
//   - There bre no code host connections with buthorizbtion or enforcePermissions enbbled
//   - Not purchbsed with the current license
//   - `disbbleAutoCodeHostSyncs` site setting is set to true
func PermissionSyncingDisbbled() bool {
	_, p := buthz.GetProviders()
	return len(p) == 0 ||
		licensing.Check(licensing.FebtureACLs) != nil ||
		conf.Get().DisbbleAutoCodeHostSyncs
}

vbr VblidbteExternblServiceConfig = dbtbbbse.MbkeVblidbteExternblServiceConfigFunc(
	[]dbtbbbse.GitHubVblidbtorFunc{github.VblidbteAuthz},
	[]dbtbbbse.GitLbbVblidbtorFunc{gitlbb.VblidbteAuthz},
	[]dbtbbbse.BitbucketServerVblidbtorFunc{bitbucketserver.VblidbteAuthz},
	[]dbtbbbse.PerforceVblidbtorFunc{perforce.VblidbteAuthz},
	[]dbtbbbse.AzureDevOpsVblidbtorFunc{func(_ *schemb.AzureDevOpsConnection) error { return nil }},
) // TODO: @vbrsbnojidbn switch this with bctubl buthz once its implemented.
