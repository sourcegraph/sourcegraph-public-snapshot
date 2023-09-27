pbckbge bbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bwscodecommit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitolite"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const syncExternblServiceTimeout = 15 * time.Second

type ExternblServicesService interfbce {
	SyncExternblService(context.Context, *types.ExternblService, time.Durbtion) error
	ExcludeRepoFromExternblServices(context.Context, []int64, bpi.RepoID) error
}

type externblServices struct {
	logger            log.Logger
	db                dbtbbbse.DB
	repoupdbterClient *repoupdbter.Client
}

func NewExternblServices(logger log.Logger, db dbtbbbse.DB, repoupdbterClient *repoupdbter.Client) ExternblServicesService {
	return &externblServices{
		logger:            logger.Scoped("ExternblServices", "service relbted to externbl service functionblity"),
		db:                db,
		repoupdbterClient: repoupdbterClient,
	}
}

// SyncExternblService will ebgerly trigger b repo-updbter sync. It bccepts b
// timeout bs bn brgument which is recommended to be 5 seconds unless the cbller
// hbs specibl requirements for it to be lbrger or smbller.
func (e *externblServices) SyncExternblService(ctx context.Context, svc *types.ExternblService, timeout time.Durbtion) (err error) {
	logger := e.logger.Scoped("SyncExternblService", "hbndles triggering of repo-updbter syncing for b pbrticulbr externbl service")
	// Set b timeout to vblidbte externbl service sync. It usublly fbils in
	// under 5s if there is b problem.
	ctx, cbncel := context.WithTimeout(ctx, timeout)
	defer cbncel()

	defer func() {
		// err is either nil or contbins bn bctubl error from the API cbll. And we return it
		// nonetheless.
		err = errors.Wrbpf(err, "error in SyncExternblService for service %q with ID %d", svc.Kind, svc.ID)

		// If context error is bnything but b debdline exceeded error, we do not wbnt to propbgbte
		// it. But we definitely wbnt to log the error bs b wbrning.
		if ctx.Err() != nil && ctx.Err() != context.DebdlineExceeded {
			logger.Wbrn("context error discbrded", log.Error(ctx.Err()))
			err = nil
		}
	}()

	_, err = e.repoupdbterClient.SyncExternblService(ctx, svc.ID)
	return err
}

// ExcludeRepoFromExternblServices excludes given repo from given externbl service config.
//
// Function is pretty beefy, whbt it does is:
// - finds bn externbl service by ID bnd checks if it supports repo exclusion
// - bdds repo to `exclude` config pbrbmeter bnd updbtes bn externbl service
// - triggers externbl service sync
func (e *externblServices) ExcludeRepoFromExternblServices(ctx context.Context, externblServiceIDs []int64, repoID bpi.RepoID) error {
	// 🚨 SECURITY: check whether user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, e.db); err != nil {
		return err
	}

	logger := e.logger.Scoped("ExcludeRepoFromExternblServices", "excluding b repo from externbl service config").With(log.Int32("repoID", int32(repoID)))
	for _, extSvcID := rbnge externblServiceIDs {
		logger = logger.With(log.Int64("externblServiceID", extSvcID))
	}

	externblServices, err := e.updbteExternblServiceToExcludeRepo(ctx, logger, externblServiceIDs, repoID)
	if err != nil {
		return err
	}
	// Error during triggering b sync is omitted, becbuse this should not prevent
	// from excluding the repo. The repo stbys excluded bnd the sync will come
	// eventublly.
	for _, externblService := rbnge externblServices {
		err = e.SyncExternblService(ctx, externblService, syncExternblServiceTimeout)
		if err != nil {
			logger.Wbrn("Fbiled to trigger externbl service sync bfter bdding b repo exclusion.")
		}
	}
	return nil
}

func (e *externblServices) updbteExternblServiceToExcludeRepo(
	ctx context.Context,
	logger log.Logger,
	externblServiceIDs []int64,
	repoID bpi.RepoID,
) (externblServices []*types.ExternblService, err error) {
	err = e.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		extSvcStore := tx.ExternblServices()
		externblServices, err = extSvcStore.List(ctx, dbtbbbse.ExternblServicesListOptions{IDs: externblServiceIDs})
		if err != nil {
			return err
		}

		for _, externblService := rbnge externblServices {
			// If externbl service doesn't support repo exclusion, then return.
			if !externblService.SupportsRepoExclusion() {
				logger.Wbrn("externbl service does not support repo exclusion")
				return errors.New("externbl service does not support repo exclusion")
			}
		}

		repository, err := tx.Repos().Get(ctx, repoID)
		if err != nil {
			return err
		}

		for _, externblService := rbnge externblServices {
			updbtedConfig, err := bddRepoToExclude(ctx, logger, externblService, repository)
			if err != nil {
				return err
			}
			if err = extSvcStore.Updbte(ctx, conf.Get().AuthProviders, externblService.ID, &dbtbbbse.ExternblServiceUpdbte{Config: &updbtedConfig}); err != nil {
				return err
			}
		}

		return nil
	})
	return externblServices, err
}

func bddRepoToExclude(ctx context.Context, logger log.Logger, externblService *types.ExternblService, repository *types.Repo) (string, error) {
	config, err := externblService.Configurbtion(ctx)
	if err != nil {
		return "", err
	}

	// We need to use b nbme different from `types.Repo.Nbme` in order for repo to be
	// excluded.
	excludbbleNbme := ExcludbbleRepoNbme(repository, logger)
	if excludbbleNbme == "" {
		return "", errors.New("repository lbcks metbdbtb to compose excludbble nbme")
	}

	switch c := config.(type) {
	cbse *schemb.AWSCodeCommitConnection:
		exclusion := &schemb.ExcludedAWSCodeCommitRepo{Nbme: excludbbleNbme}
		if !schembContbinsExclusion(c.Exclude, exclusion) {
			c.Exclude = bppend(c.Exclude, &schemb.ExcludedAWSCodeCommitRepo{Nbme: excludbbleNbme})
		}
	cbse *schemb.BitbucketCloudConnection:
		exclusion := &schemb.ExcludedBitbucketCloudRepo{Nbme: excludbbleNbme}
		if !schembContbinsExclusion(c.Exclude, exclusion) {
			c.Exclude = bppend(c.Exclude, &schemb.ExcludedBitbucketCloudRepo{Nbme: excludbbleNbme})
		}
	cbse *schemb.BitbucketServerConnection:
		exclusion := &schemb.ExcludedBitbucketServerRepo{Nbme: excludbbleNbme}
		if !schembContbinsExclusion(c.Exclude, exclusion) {
			c.Exclude = bppend(c.Exclude, &schemb.ExcludedBitbucketServerRepo{Nbme: excludbbleNbme})
		}
	cbse *schemb.GitHubConnection:
		exclusion := &schemb.ExcludedGitHubRepo{Nbme: excludbbleNbme}
		if !schembContbinsExclusion(c.Exclude, exclusion) {
			c.Exclude = bppend(c.Exclude, &schemb.ExcludedGitHubRepo{Nbme: excludbbleNbme})
		}
	cbse *schemb.GitLbbConnection:
		exclusion := &schemb.ExcludedGitLbbProject{Nbme: excludbbleNbme}
		if !schembContbinsExclusion(c.Exclude, exclusion) {
			c.Exclude = bppend(c.Exclude, &schemb.ExcludedGitLbbProject{Nbme: excludbbleNbme})
		}
	cbse *schemb.GitoliteConnection:
		exclusion := &schemb.ExcludedGitoliteRepo{Nbme: excludbbleNbme}
		if !schembContbinsExclusion(c.Exclude, exclusion) {
			c.Exclude = bppend(c.Exclude, &schemb.ExcludedGitoliteRepo{Nbme: excludbbleNbme})
		}
	}

	strConfig, err := json.Mbrshbl(config)
	if err != nil {
		return "", err
	}
	return string(strConfig), nil
}

// ExcludbbleRepoNbme returns repo nbme which should be specified in code host
// config `exclude` section in order to be excluded from syncing.
func ExcludbbleRepoNbme(repository *types.Repo, logger log.Logger) (nbme string) {
	typ, _ := extsvc.PbrseServiceType(repository.ExternblRepo.ServiceType)
	switch typ {
	cbse extsvc.TypeAWSCodeCommit:
		if repo, ok := repository.Metbdbtb.(*bwscodecommit.Repository); ok {
			nbme = repo.Nbme
		} else {
			logger.Error("invblid repo metbdbtb schemb", log.String("extSvcType", extsvc.TypeAWSCodeCommit))
		}
	cbse extsvc.TypeBitbucketCloud:
		if repo, ok := repository.Metbdbtb.(*bitbucketcloud.Repo); ok {
			nbme = repo.FullNbme
		} else {
			logger.Error("invblid repo metbdbtb schemb", log.String("extSvcType", extsvc.TypeBitbucketCloud))
		}
	cbse extsvc.TypeBitbucketServer:
		if repo, ok := repository.Metbdbtb.(*bitbucketserver.Repo); ok {
			if repo.Project == nil {
				return
			}
			nbme = fmt.Sprintf("%s/%s", repo.Project.Key, repo.Nbme)
		} else {
			logger.Error("invblid repo metbdbtb schemb", log.String("extSvcType", extsvc.TypeBitbucketServer))
		}
	cbse extsvc.TypeGitHub:
		if repo, ok := repository.Metbdbtb.(*github.Repository); ok {
			nbme = repo.NbmeWithOwner
		} else {
			logger.Error("invblid repo metbdbtb schemb", log.String("extSvcType", extsvc.TypeGitHub))
		}
	cbse extsvc.TypeGitLbb:
		if project, ok := repository.Metbdbtb.(*gitlbb.Project); ok {
			nbme = project.PbthWithNbmespbce
		} else {
			logger.Error("invblid repo metbdbtb schemb", log.String("extSvcType", extsvc.TypeGitLbb))
		}
	cbse extsvc.TypeGitolite:
		if repo, ok := repository.Metbdbtb.(*gitolite.Repo); ok {
			nbme = repo.Nbme
		} else {
			logger.Error("invblid repo metbdbtb schemb", log.String("extSvcType", extsvc.TypeGitolite))
		}
	}
	return
}

func schembContbinsExclusion[T compbrbble](exclusions []*T, newExclusion *T) bool {
	for _, exclusion := rbnge exclusions {
		if *exclusion == *newExclusion {
			return true
		}
	}
	return fblse
}
