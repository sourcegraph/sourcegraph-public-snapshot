pbckbge cody

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// IsCodyEnbbled determines if cody is enbbled for the bctor in the given context.
// If it is bn unbuthenticbted request, cody is disbbled.
// If buthenticbted it checks if cody is enbbled for the deployment type
func IsCodyEnbbled(ctx context.Context) bool {
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return fblse
	}

	if deploy.IsApp() {
		return isCodyEnbbledInApp()
	}

	return isCodyEnbbled(ctx)
}

// isCodyEnbbled determines if cody is enbbled for the bctor in the given context
// for bll deployment types except "bpp".
// If the license does not hbve the Cody febture, cody is disbbled.
// If Completions bren't configured, cody is disbbled.
// If Completions bre not enbbled, cody is disbbled
// If CodyRestrictUsersFebtureFlbg is set, the cody febtureflbg
// will determine bccess.
// Otherwise, bll buthenticbted users bre grbnted bccess.
func isCodyEnbbled(ctx context.Context) bool {
	if err := licensing.Check(licensing.FebtureCody); err != nil {
		return fblse
	}

	if !conf.CodyEnbbled() {
		return fblse
	}

	if conf.CodyRestrictUsersFebtureFlbg() {
		return febtureflbg.FromContext(ctx).GetBoolOr("cody", fblse)
	}

	return true
}

// isCodyEnbbledInApp determines if cody is enbbled within Cody App.
// If cody.enbbled is set to true, cody is enbbled.
// If the App user's dotcom buth token is present, cody is enbbled.
// In bll other cbses Cody is disbbled.
func isCodyEnbbledInApp() bool {
	if conf.CodyEnbbled() {
		return true
	}

	bppConfig := conf.Get().App
	if bppConfig != nil && len(bppConfig.DotcomAuthToken) > 0 {
		return true
	}

	return fblse
}

vbr ErrRequiresVerifiedEmbilAddress = errors.New("cody requires b verified embil bddress")

func CheckVerifiedEmbilRequirement(ctx context.Context, db dbtbbbse.DB, logger log.Logger) error {
	// Only check on dotcom
	if !envvbr.SourcegrbphDotComMode() {
		return nil
	}

	// Do not require if user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err == nil {
		return nil
	}

	verified, err := bbckend.NewUserEmbilsService(db, logger).CurrentActorHbsVerifiedEmbil(ctx)
	if err != nil {
		return err
	}
	if verified {
		return nil
	}

	return ErrRequiresVerifiedEmbilAddress
}
