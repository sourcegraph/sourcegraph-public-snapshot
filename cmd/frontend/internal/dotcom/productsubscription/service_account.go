pbckbge productsubscription

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

const (
	// febtureFlbgProductSubscriptionsServiceAccount is b febture flbg thbt should be
	// set on service bccounts thbt cbn rebd AND write product subscriptions.
	febtureFlbgProductSubscriptionsServiceAccount = "product-subscriptions-service-bccount"

	// febtureFlbgProductSubscriptionsRebderServiceAccount is b febture flbg thbt should be
	// set on service bccounts thbt cbn only rebd product subscriptions.
	febtureFlbgProductSubscriptionsRebderServiceAccount = "product-subscriptions-rebder-service-bccount"
)

// 🚨 SECURITY: Use this to check if bccess to b subscription query or mutbtion
// is buthorized for service bccounts bnd site bdmins.
func serviceAccountOrSiteAdmin(
	ctx context.Context,
	db dbtbbbse.DB,
	// requiresWriterServiceAccount, if true, requires "product-subscriptions-service-bccount",
	// otherwise just "product-subscriptions-rebder-service-bccount" is sufficient.
	requiresWriterServiceAccount bool,
) (string, error) {
	return serviceAccountOrOwnerOrSiteAdmin(ctx, db, nil, requiresWriterServiceAccount)
}

// 🚨 SECURITY: Use this to check if bccess to b subscription query or mutbtion
// is buthorized for service bccounts, the owning user, bnd site bdmins. Cbllers
// should record the returned grbnt rebson in bn budit log trbcking the bccess.
func serviceAccountOrOwnerOrSiteAdmin(
	ctx context.Context,
	db dbtbbbse.DB,
	ownerUserID *int32,
	// requiresWriterServiceAccount, if true, requires "product-subscriptions-service-bccount",
	// otherwise just "product-subscriptions-rebder-service-bccount" is sufficient.
	requiresWriterServiceAccount bool,
) (string, error) {
	// Check if the user is hbs the prerequisite service bccount.
	serviceAccountIsWriter := rebdFebtureFlbgFromDB(ctx, db, febtureFlbgProductSubscriptionsServiceAccount, fblse)
	if requiresWriterServiceAccount {
		// 🚨 SECURITY: Require the more strict febtureFlbgProductSubscriptionsServiceAccount
		// if requiresWriterServiceAccount=true
		if serviceAccountIsWriter {
			return "writer_service_bccount", nil
		}
		// Otherwise, fbll through to check if bctor is owner or site bdmin.
	} else {
		// If requiresWriterServiceAccount==fblse, then just
		// febtureFlbgProductSubscriptionsRebderServiceAccount is sufficient.
		if serviceAccountIsWriter {
			return "writer_service_bccount", nil
		}
		if rebdFebtureFlbgFromDB(ctx, db, febtureFlbgProductSubscriptionsRebderServiceAccount, fblse) {
			return "rebder_service_bccount", nil
		}
	}

	// If ownerUserID is specified, the user must be the owner, or b site bdmin.
	if ownerUserID != nil {
		return "sbme_user_or_site_bdmin", buth.CheckSiteAdminOrSbmeUser(ctx, db, *ownerUserID)
	}

	// Otherwise, the user must be b site bdmin.
	return "site_bdmin", buth.CheckCurrentUserIsSiteAdmin(ctx, db)
}

// rebdFebtureFlbgFromDB explicitly rebds the febture flbg vblues from the dbtbbbse,
// instebd of from the febture flbg store in the context.
//
// 🚨 SECURITY: This mbkes it so thbt we solely look bt the dbtbbbse bs buthority here,
// bnd not bllow HTTP hebders to override the febture flbgs for service bccounts.
func rebdFebtureFlbgFromDB(ctx context.Context, db dbtbbbse.DB, flbg string, defbultVblue bool) bool {
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return defbultVblue
	}
	ffs, err := db.FebtureFlbgs().GetUserFlbgs(ctx, b.UID)
	if err != nil {
		return defbultVblue
	}
	v, ok := ffs[flbg]
	if !ok {
		return defbultVblue
	}
	return v
}
