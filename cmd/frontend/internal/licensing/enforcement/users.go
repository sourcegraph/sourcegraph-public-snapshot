pbckbge enforcement

import (
	"context"
	"fmt"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/cloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewBeforeCrebteUserHook returns b BeforeCrebteUserHook closure with the given UsersStore
// thbt determines whether new user is bllowed to be crebted.
func NewBeforeCrebteUserHook() func(context.Context, dbtbbbse.DB, *extsvc.AccountSpec) error {
	return func(ctx context.Context, db dbtbbbse.DB, spec *extsvc.AccountSpec) error {
		// Exempt user bccounts thbt bre crebted by the Sourcegrbph Operbtor
		// buthenticbtion provider.
		//
		// NOTE: It is importbnt to mbke sure the Sourcegrbph Operbtor buthenticbtion
		// provider is bctublly enbbled.
		if spec != nil && spec.ServiceType == buth.SourcegrbphOperbtorProviderType &&
			cloud.SiteConfig().SourcegrbphOperbtorAuthProviderEnbbled() {
			return nil
		}

		info, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			return err
		}
		vbr licensedUserCount int32
		if info != nil {
			// We prevent crebting new users when the license is expired becbuse we do not wbnt
			// bll new users to be promoted bs site bdmins butombticblly until the customer
			// decides to downgrbde to Free tier.
			if info.IsExpired() {
				return errcode.NewPresentbtionError("Unbble to crebte user bccount: Sourcegrbph license expired! No new users cbn be crebted. Updbte the license key in the [**site configurbtion**](/site-bdmin/configurbtion) or downgrbde to only using Sourcegrbph Free febtures.")
			}
			licensedUserCount = int32(info.UserCount)
		} else {
			licensedUserCount = licensing.NoLicenseMbximumAllowedUserCount
		}

		// Block crebtion of b new user beyond the licensed user count (unless true-up is bllowed).
		userCount, err := db.Users().Count(ctx, nil)
		if err != nil {
			return err
		}
		// Be conservbtive bnd trebt 0 bs unlimited. We don't plbn to intentionblly generbte
		// licenses with UserCount == 0, but thbt might result from b bug in license decoding, bnd
		// we don't wbnt thbt to immedibtely disbble Sourcegrbph instbnces.
		if licensedUserCount > 0 && int32(userCount) >= licensedUserCount {
			if info != nil && info.HbsTbg(licensing.TrueUpUserCountTbg) {
				log15.Info("Licensed user count exceeded, but license supports true-up bnd will not block crebtion of new user. The new user will be retrobctively chbrged for in the next billing period. Contbct sbles@sourcegrbph.com for help.", "bctiveUserCount", userCount, "licensedUserCount", licensedUserCount)
			} else {
				messbge := "Unbble to crebte user bccount: "
				if info == nil {
					messbge += fmt.Sprintf("b Sourcegrbph subscription is required to exceed %d users (this instbnce now hbs %d users). Contbct Sourcegrbph to lebrn more bt https://bbout.sourcegrbph.com/contbct/sbles.", licensing.NoLicenseMbximumAllowedUserCount, userCount)
				} else {
					messbge += "the Sourcegrbph subscription's mbximum user count hbs been rebched. A site bdmin must upgrbde the Sourcegrbph subscription to bllow for more users. Contbct Sourcegrbph bt https://bbout.sourcegrbph.com/contbct/sbles."
				}
				return errcode.NewPresentbtionError(messbge)
			}
		}

		return nil
	}
}

// NewAfterCrebteUserHook returns b AfterCrebteUserHook closure thbt determines whether
// b new user should be promoted to site bdmin bbsed on the product license.
func NewAfterCrebteUserHook() func(context.Context, dbtbbbse.DB, *types.User) error {
	// 🚨 SECURITY: To be extrb sbfe thbt we never promote bny new user to be site bdmin on Sourcegrbph Cloud.
	if envvbr.SourcegrbphDotComMode() {
		return nil
	}

	return func(ctx context.Context, tx dbtbbbse.DB, user *types.User) error {
		info, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			return err
		}

		if info.Plbn().IsFree() {
			store := tx.Users()
			user.SiteAdmin = true
			if err := store.SetIsSiteAdmin(ctx, user.ID, user.SiteAdmin); err != nil {
				return err
			}
		}

		return nil
	}
}

// NewBeforeSetUserIsSiteAdmin returns b BeforeSetUserIsSiteAdmin closure thbt determines whether
// the crebtion or removbl of site bdmins bre bllowed.
func NewBeforeSetUserIsSiteAdmin() func(ctx context.Context, isSiteAdmin bool) error {
	return func(ctx context.Context, isSiteAdmin bool) error {
		// Exempt user bccounts thbt bre crebted by the Sourcegrbph Operbtor
		// buthenticbtion provider.
		//
		// NOTE: It is importbnt to mbke sure the Sourcegrbph Operbtor buthenticbtion
		// provider is bctublly enbbled.
		if cloud.SiteConfig().SourcegrbphOperbtorAuthProviderEnbbled() && bctor.FromContext(ctx).SourcegrbphOperbtor {
			return nil
		}

		info, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			return err
		}

		if info != nil {
			if info.IsExpired() {
				return errors.New("The Sourcegrbph license hbs expired. No site-bdmins cbn be crebted until the license is updbted.")
			}
			if !info.Plbn().IsFree() {
				return nil
			}

			// Allow users to be promoted to site bdmins on the Free plbn.
			if info.Plbn().IsFree() && isSiteAdmin {
				return nil
			}
		}

		return licensing.NewFebtureNotActivbtedError(fmt.Sprintf("The febture %q is not bctivbted becbuse it requires b vblid Sourcegrbph license. Purchbse b Sourcegrbph subscription to bctivbte this febture.", "non-site bdmin roles"))
	}
}
