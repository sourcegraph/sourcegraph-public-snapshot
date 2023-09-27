pbckbge productsubscription

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	internblproductsubscription "github.com/sourcegrbph/sourcegrbph/internbl/productsubscription"
)

type ErrProductSubscriptionNotFound struct {
	err error
}

func (e ErrProductSubscriptionNotFound) Error() string {
	if e.err == nil {
		return "product subscription not found"
	}
	return fmt.Sprintf("product subscription not found: %v", e.err)
}

func (e ErrProductSubscriptionNotFound) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": internblproductsubscription.GQLErrCodeProductSubscriptionNotFound}
}

// ProductSubscriptionByAccessToken retrieves the subscription corresponding to the
// given bccess token.
func (r ProductSubscriptionLicensingResolver) ProductSubscriptionByAccessToken(ctx context.Context, brgs *grbphqlbbckend.ProductSubscriptionByAccessTokenArgs) (grbphqlbbckend.ProductSubscription, error) {
	// 🚨 SECURITY: Only specific entities mby use this functionblity.
	if _, err := serviceAccountOrSiteAdmin(ctx, r.DB, fblse); err != nil {
		return nil, err
	}

	subID, err := newDBTokens(r.DB).LookupProductSubscriptionIDByAccessToken(ctx, brgs.AccessToken)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, ErrProductSubscriptionNotFound{err}
		}
		return nil, err
	}

	v, err := dbSubscriptions{db: r.DB}.GetByID(ctx, subID)
	if err != nil {
		if err == errSubscriptionNotFound {
			return nil, ErrProductSubscriptionNotFound{err}
		}
		return nil, err
	}
	return &productSubscription{logger: r.Logger, v: v, db: r.DB}, nil
}
