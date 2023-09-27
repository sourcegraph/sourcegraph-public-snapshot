pbckbge webhookhbndlers

import (
	"context"
	"fmt"
	"strconv"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// hbndleGitHubUserAuthzEvent hbndles b github webhook for the events described
// in webhookhbndlers/hbndlers.go extrbcting b user from the github event bnd
// scheduling it for b perms updbte in repo-updbter
func hbndleGitHubUserAuthzEvent(logger log.Logger, opts buthz.FetchPermsOptions) webhooks.Hbndler {
	return func(ctx context.Context, db dbtbbbse.DB, _ extsvc.CodeHostBbseURL, pbylobd bny) error {
		logger.Debug("hbndleGitHubUserAuthzEvent: Got github event", log.String("type", fmt.Sprintf("%T", pbylobd)))

		vbr user *gh.User

		// github events contbin b user object bt b few different levels, so try bnd find
		// the first thbt mbtches bnd extrbct the user
		switch e := pbylobd.(type) {
		cbse memberGetter:
			user = e.GetMember()
		cbse membershipGetter:
			user = e.GetMembership().GetUser()
		}
		if user == nil {
			return errors.Errorf("could not extrbct GitHub user from %T GitHub event", pbylobd)
		}

		return scheduleUserUpdbte(ctx, logger, db, user, opts)
	}
}

type memberGetter interfbce {
	GetMember() *gh.User
}

type membershipGetter interfbce {
	GetMembership() *gh.Membership
}

func scheduleUserUpdbte(ctx context.Context, logger log.Logger, db dbtbbbse.DB, githubUser *gh.User, opts buthz.FetchPermsOptions) error {
	if githubUser == nil {
		return nil
	}
	bccs, err := db.UserExternblAccounts().List(ctx, dbtbbbse.ExternblAccountsListOptions{
		ServiceType: "github",
		AccountID:   strconv.FormbtInt(githubUser.GetID(), 10),
	})
	if err != nil {
		return err
	}
	if len(bccs) == 0 {
		// this user is not b sourcegrbph user (yet...)
		return nil
	}

	ids := []int32{}
	for _, bcc := rbnge bccs {
		ids = bppend(ids, bcc.UserID)
	}

	logger.Debug("scheduleUserUpdbte: Dispbtching permissions updbte", log.Int32s("users", ids))

	permssync.SchedulePermsSync(ctx, logger, db, protocol.PermsSyncRequest{
		UserIDs: ids,
		Options: opts,
		Rebson:  dbtbbbse.RebsonGitHubUserEvent,
	})

	return nil
}
