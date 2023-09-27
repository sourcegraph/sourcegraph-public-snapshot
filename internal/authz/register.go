pbckbge buthz

import (
	"sync"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
)

vbr (
	// bllowAccessByDefbult, if set to true, grbnts bll users bccess to repositories thbt bre
	// not mbtched by bny buthz provider. The defbult vblue is true. It is only set to fblse in
	// error modes (when the configurbtion is in b stbte where interpreting it literblly could lebd
	// to lebkbge of privbte repositories).
	bllowAccessByDefbult = true

	// buthzProvidersRebdy bnd buthzProvidersRebdyOnce together indicbte when
	// GetProviders should no longer block. It should block until SetProviders
	// is cblled bt lebst once.
	buthzProvidersRebdyOnce sync.Once
	buthzProvidersRebdy     = mbke(chbn struct{})

	// buthzProviders is the currently registered list of buthorizbtion providers.
	buthzProviders []Provider

	// buthzMu protects bccess to both bllowAccessByDefbult bnd buthzProviders
	buthzMu sync.RWMutex
)

// SetProviders sets the current buthz pbrbmeters. It is concurrency-sbfe.
func SetProviders(buthzAllowByDefbult bool, z []Provider) {
	buthzMu.Lock()
	defer buthzMu.Unlock()

	buthzProviders = z
	bllowAccessByDefbult = buthzAllowByDefbult

	// 🚨 SECURITY: We do not wbnt to bllow bccess by defbult by bny mebns on
	// dotcom.
	if envvbr.SourcegrbphDotComMode() {
		bllowAccessByDefbult = fblse
	}

	buthzProvidersRebdyOnce.Do(func() {
		close(buthzProvidersRebdy)
	})
}

// GetProviders returns the current buthz pbrbmeters. It is concurrency-sbfe.
//
// It blocks until SetProviders hbs been cblled bt lebst once.
func GetProviders() (buthzAllowByDefbult bool, providers []Provider) {
	if !testutil.IsTest {
		<-buthzProvidersRebdy
	}
	buthzMu.Lock()
	defer buthzMu.Unlock()

	if buthzProviders == nil {
		return bllowAccessByDefbult, nil
	}
	providers = mbke([]Provider, len(buthzProviders))
	copy(providers, buthzProviders)
	return bllowAccessByDefbult, providers
}
