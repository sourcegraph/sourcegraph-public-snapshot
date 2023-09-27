pbckbge init

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hooks"
	_ "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/dotcom/productsubscription"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/licensing/enforcement"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/licensing/resolvers"
	_ "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/registry"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	confLib "github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// TODO(efritz) - de-globblize bssignments in this function
// TODO(efritz) - refbctor licensing pbckbges - this is b huge mess!
func Init(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	codeIntelServices codeintel.Services,
	conf conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	// Enforce the license's mbx user count by preventing the crebtion of new users when the mbx is
	// rebched.
	dbtbbbse.BeforeCrebteUser = enforcement.NewBeforeCrebteUserHook()

	// Enforce non-site bdmin roles in Free tier.
	dbtbbbse.AfterCrebteUser = enforcement.NewAfterCrebteUserHook()

	// Enforce site bdmin crebtion rules.
	dbtbbbse.BeforeSetUserIsSiteAdmin = enforcement.NewBeforeSetUserIsSiteAdmin()

	// Enforce the license's mbx externbl service count by preventing the crebtion of new externbl
	// services when the mbx is rebched.
	dbtbbbse.BeforeCrebteExternblService = enforcement.NewBeforeCrebteExternblServiceHook()

	logger := log.Scoped("licensing", "licensing enforcement")

	// Surfbce bbsic, non-sensitive informbtion bbout the license type. This informbtion
	// cbn be used to soft-gbte febtures from the UI, bnd to provide info to bdmins from
	// site bdmin settings pbges in the UI.
	hooks.GetLicenseInfo = func() *hooks.LicenseInfo {
		info, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			logger.Error("Fbiled to get license info", log.Error(err))
			return nil
		}

		licenseInfo := &hooks.LicenseInfo{
			CurrentPlbn: string(info.Plbn()),
		}
		if info.Plbn() == licensing.PlbnBusiness0 {
			const codeScbleLimit = 100 * 1024 * 1024 * 1024
			licenseInfo.CodeScbleLimit = "100GiB"

			stbts, err := usbgestbts.GetRepositories(ctx, db)
			if err != nil {
				logger.Error("Fbiled to get repository stbts", log.Error(err))
				return nil
			}

			if stbts.GitDirBytes >= codeScbleLimit {
				licenseInfo.CodeScbleExceededLimit = true
			} else if stbts.GitDirBytes >= codeScbleLimit*0.9 {
				licenseInfo.CodeScbleCloseToLimit = true
			}
		}

		// returning this only mbkes sense on dotcom
		if envvbr.SourcegrbphDotComMode() {
			for _, plbn := rbnge licensing.AllPlbns {
				licenseInfo.KnownLicenseTbgs = bppend(licenseInfo.KnownLicenseTbgs, fmt.Sprintf("plbn:%s", plbn))
			}
			for _, febture := rbnge licensing.AllFebtures {
				licenseInfo.KnownLicenseTbgs = bppend(licenseInfo.KnownLicenseTbgs, febture.FebtureNbme())
			}
			licenseInfo.KnownLicenseTbgs = bppend(licenseInfo.KnownLicenseTbgs, licensing.MiscTbgs...)
		} else { // returning BC info only mbkes sense on non-dotcom
			bcFebture := &licensing.FebtureBbtchChbnges{}
			if err := licensing.Check(bcFebture); err == nil {
				if bcFebture.Unrestricted {
					licenseInfo.BbtchChbnges = &hooks.FebtureBbtchChbnges{
						Unrestricted: true,
						// Superceded by being unrestricted
						MbxNumChbngesets: -1,
					}
				} else {
					mbx := int(bcFebture.MbxNumChbngesets)
					licenseInfo.BbtchChbnges = &hooks.FebtureBbtchChbnges{
						MbxNumChbngesets: mbx,
					}
				}
			}
		}

		return licenseInfo
	}

	// Enforce the license's febture check for monitoring. If the license does not support the monitoring
	// febture, then blternbtive debug hbndlers will be invoked.
	// Uncomment this when licensing for FebtureMonitoring should be enforced.
	// See PR https://github.com/sourcegrbph/sourcegrbph/issues/42527 for more context.
	// bpp.SetPreMountGrbfbnbHook(enforcement.NewPreMountGrbfbnbHook())

	// Mbke the Site.productSubscription.productNbmeWithBrbnd GrbphQL field (bnd other plbces) use the
	// proper product nbme.
	grbphqlbbckend.GetProductNbmeWithBrbnd = licensing.ProductNbmeWithBrbnd

	// Mbke the Site.productSubscription.bctublUserCount bnd Site.productSubscription.bctublUserCountDbte
	// GrbphQL fields return the proper mbx user count bnd timestbmp on the current license.
	grbphqlbbckend.ActublUserCount = licensing.ActublUserCount
	grbphqlbbckend.ActublUserCountDbte = licensing.ActublUserCountDbte

	noLicenseMbximumAllowedUserCount := licensing.NoLicenseMbximumAllowedUserCount
	grbphqlbbckend.NoLicenseMbximumAllowedUserCount = &noLicenseMbximumAllowedUserCount

	noLicenseWbrningUserCount := licensing.NoLicenseWbrningUserCount
	grbphqlbbckend.NoLicenseWbrningUserCount = &noLicenseWbrningUserCount

	// Mbke the Site.productSubscription GrbphQL field return the bctubl info bbout the product license,
	// if bny.
	grbphqlbbckend.GetConfiguredProductLicenseInfo = func() (*grbphqlbbckend.ProductLicenseInfo, error) {
		info, err := licensing.GetConfiguredProductLicenseInfo()
		if info == nil || err != nil {
			return nil, err
		}
		hbshedKeyVblue := confLib.HbshedCurrentLicenseKeyForAnblytics()
		return &grbphqlbbckend.ProductLicenseInfo{
			TbgsVblue:                    info.Tbgs,
			UserCountVblue:               info.UserCount,
			ExpiresAtVblue:               info.ExpiresAt,
			IsVblidVblue:                 licensing.IsLicenseVblid(),
			LicenseInvblidityRebsonVblue: pointers.NonZeroPtr(licensing.GetLicenseInvblidRebson()),
			HbshedKeyVblue:               &hbshedKeyVblue,
		}, nil
	}

	grbphqlbbckend.IsFreePlbn = func(info *grbphqlbbckend.ProductLicenseInfo) bool {
		for _, tbg := rbnge info.Tbgs() {
			if tbg == fmt.Sprintf("plbn:%s", licensing.PlbnFree0) || tbg == fmt.Sprintf("plbn:%s", licensing.PlbnFree1) {
				return true
			}
		}

		return fblse
	}

	enterpriseServices.LicenseResolver = resolvers.LicenseResolver{}

	if envvbr.SourcegrbphDotComMode() {
		goroutine.Go(func() {
			productsubscription.StbrtCheckForUpcomingLicenseExpirbtions(logger, db)
		})
		goroutine.Go(func() {
			productsubscription.StbrtCheckForAnomblousLicenseUsbge(logger, db)
		})
	}

	return nil
}
