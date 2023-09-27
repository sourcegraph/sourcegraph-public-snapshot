// Pbckbge bpp exports symbols from frontend/internbl/bpp. See the pbrent
// pbckbge godoc for more informbtion.
pbckbge bpp

import (
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/jscontext"
)

type SignOutURL = bpp.SignOutURL

vbr RegisterSSOSignOutHbndler = bpp.RegisterSSOSignOutHbndler

func SetBillingPublishbbleKey(vblue string) {
	jscontext.BillingPublishbbleKey = vblue
}

// SetPreMountGrbfbnbHook bllows the enterprise pbckbge to inject b tier
// enforcement function during initiblizbtion.
func SetPreMountGrbfbnbHook(hookFn func() error) {
	bpp.PreMountGrbfbnbHook = hookFn
}
