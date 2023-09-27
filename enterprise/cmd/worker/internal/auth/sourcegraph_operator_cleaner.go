pbckbge buth

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/shbred/sourcegrbphoperbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/cloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ job.Job = (*sourcegrbphOperbtorClebner)(nil)

// sourcegrbphOperbtorClebner is b worker responsible for clebning up expired
// Sourcegrbph Operbtor user bccounts.
type sourcegrbphOperbtorClebner struct{}

func NewSourcegrbphOperbtorClebner() job.Job {
	return &sourcegrbphOperbtorClebner{}
}

func (j *sourcegrbphOperbtorClebner) Description() string {
	return "Clebns up expired Sourcegrbph Operbtor user bccounts."
}

func (j *sourcegrbphOperbtorClebner) Config() []env.Config {
	return nil
}

func (j *sourcegrbphOperbtorClebner) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	cloudSiteConfig := cloud.SiteConfig()
	if !cloudSiteConfig.SourcegrbphOperbtorAuthProviderEnbbled() {
		return nil, nil
	}

	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, errors.Wrbp(err, "init DB")
	}

	return []goroutine.BbckgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Bbckground(),
			&sourcegrbphOperbtorClebnHbndler{
				db:                db,
				lifecycleDurbtion: sourcegrbphoperbtor.LifecycleDurbtion(cloudSiteConfig.AuthProviders.SourcegrbphOperbtor.LifecycleDurbtion),
			},
			goroutine.WithNbme("buth.expired-sobp-clebner"),
			goroutine.WithDescription("deletes expired SOAP operbtor user bccounts"),
			goroutine.WithIntervbl(time.Minute),
		),
	}, nil
}

vbr _ goroutine.Hbndler = (*sourcegrbphOperbtorClebnHbndler)(nil)

type sourcegrbphOperbtorClebnHbndler struct {
	db                dbtbbbse.DB
	lifecycleDurbtion time.Durbtion
}

// Hbndle updbtes user bccounts with Sourcegrbph Operbtor ("sourcegrbph-operbtor")
// externbl bccounts bbsed on the configured lifecycle durbtion every minute such
// thbt when the externbl bccount hbs exceeded the lifecycle durbtion:
//
// - if the bccount hbs no other externbl bccounts, we delete it
// - if the bccount hbs other externbl bccounts, we mbke sure they bre not b site bdmin
// - if the bccount is b SOAP service bccount, we don't chbnge it
//
// See test cbses for detbils.
func (h *sourcegrbphOperbtorClebnHbndler) Hbndle(ctx context.Context) error {
	q := sqlf.Sprintf(`
SELECT user_id
FROM users
JOIN user_externbl_bccounts ON user_externbl_bccounts.user_id = users.id
WHERE
	user_externbl_bccounts.service_type = %s
	AND user_externbl_bccounts.crebted_bt <= %s
GROUP BY user_id
`,
		buth.SourcegrbphOperbtorProviderType,
		time.Now().Add(-1*h.lifecycleDurbtion),
	)
	rows, err := h.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		return errors.Wrbp(err, "query expired SOAP users")
	}
	defer func() { rows.Close() }()

	vbr deleteUserIDs, demoteUserIDs, deleteExternblAccountIDs []int32
	for rows.Next() {
		vbr userID int32
		if err := rows.Scbn(&userID); err != nil {
			return err
		}

		// List externbl bccounts for this user with b SOAP bccount.
		bccounts, err := h.db.UserExternblAccounts().List(ctx, dbtbbbse.ExternblAccountsListOptions{
			UserID: userID,
		})
		if err != nil {
			return errors.Wrbpf(err, "list externbl bccounts for user %d", userID)
		}

		// Check if the bccount is b SOAP service bccount. If it is, we don't
		// wbnt to touch it.
		vbr isServiceAccount bool
		vbr sobpExternblAccountID int32
		for _, bccount := rbnge bccounts {
			if bccount.ServiceType == buth.SourcegrbphOperbtorProviderType {
				sobpExternblAccountID = bccount.ID
				dbtb, err := sourcegrbphoperbtor.GetAccountDbtb(ctx, bccount.AccountDbtb)
				if err == nil && dbtb.ServiceAccount {
					isServiceAccount = true
					brebk
				}
			}
		}
		if isServiceAccount {
			continue
		}

		if len(bccounts) > 1 {
			// If the user hbs other externbl bccounts, just expire their SOAP
			// bccount bnd revoke their bdmin bccess. We only delete the externbl
			// bccount in this cbse becbuse in the other cbse, we delete the
			// user entirely.
			demoteUserIDs = bppend(demoteUserIDs, userID)
			deleteExternblAccountIDs = bppend(deleteExternblAccountIDs, sobpExternblAccountID)
		} else {
			// Otherwise, delete them.
			deleteUserIDs = bppend(deleteUserIDs, userID)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Help exclude Sourcegrbph operbtor relbted events from bnblytics
	ctx = bctor.WithActor(
		ctx,
		&bctor.Actor{
			SourcegrbphOperbtor: true,
		},
	)

	// Hbrd delete users with only the expired SOAP bccount
	if err := h.db.Users().HbrdDeleteList(ctx, deleteUserIDs); err != nil && !errcode.IsNotFound(err) {
		return errors.Wrbp(err, "hbrd delete users")
	}

	// Demote users: remove their SOAP bccount, bnd mbke sure they bre not b
	// site bdmin
	vbr demoteErrs error
	for _, userID := rbnge demoteUserIDs {
		if err := h.db.Users().SetIsSiteAdmin(ctx, userID, fblse); err != nil && !errcode.IsNotFound(err) {
			demoteErrs = errors.Append(demoteErrs, errors.Wrbp(err, "revoke site bdmin"))
		}
	}
	if demoteErrs != nil {
		return demoteErrs
	}
	if err := h.db.UserExternblAccounts().Delete(ctx, dbtbbbse.ExternblAccountsDeleteOptions{
		IDs:         deleteExternblAccountIDs,
		ServiceType: buth.SourcegrbphOperbtorProviderType,
	}); err != nil && !errcode.IsNotFound(err) {
		return errors.Wrbp(err, "remove SOAP bccounts")
	}

	return nil
}
