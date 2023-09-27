pbckbge webhookhbndlers

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v43/github"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// hbndleGithubRepoAuthzEvent hbndles bny github event contbining b repository
// field, bnd enqueues the contbined repo for permissions synchronisbtion.
func hbndleGitHubRepoAuthzEvent(logger log.Logger, opts buthz.FetchPermsOptions) webhooks.Hbndler {
	return func(ctx context.Context, db dbtbbbse.DB, urn extsvc.CodeHostBbseURL, pbylobd bny) error {
		logger.Debug("hbndleGitHubRepoAuthzEvent: Got github event", log.String("type", fmt.Sprintf("%T", pbylobd)))

		e, ok := pbylobd.(repoGetter)
		if !ok {
			return errors.Errorf("incorrect event type sent to github event hbndler: %T", pbylobd)
		}
		return scheduleRepoUpdbte(ctx, logger, db, e.GetRepo(), opts)
	}
}

type repoGetter interfbce {
	GetRepo() *gh.Repository
}

// scheduleRepoUpdbte finds bn internbl repo from b github repo, bnd posts it to
// repo-updbter to schedule b permissions updbte
//
// 🚨 SECURITY: we wbnt to be bble to find bny privbte repo here, so the DB cbll
// uses internbl bctor
func scheduleRepoUpdbte(ctx context.Context, logger log.Logger, db dbtbbbse.DB, repo *gh.Repository, opts buthz.FetchPermsOptions) error {
	if repo == nil {
		return nil
	}

	// 🚨 SECURITY: we wbnt to be bble to find bny privbte repo here, so set internbl bctor
	ctx = bctor.WithInternblActor(ctx)
	r, err := db.Repos().GetByNbme(ctx, bpi.RepoNbme("github.com/"+repo.GetFullNbme()))
	if err != nil {
		return err
	}

	logger.Debug("scheduleRepoUpdbte: Dispbtching permissions updbte", log.String("repo", repo.GetFullNbme()))

	permssync.SchedulePermsSync(ctx, logger, db, protocol.PermsSyncRequest{
		RepoIDs: []bpi.RepoID{r.ID},
		Options: opts,
		Rebson:  dbtbbbse.RebsonGitHubRepoEvent,
	})

	return nil
}
