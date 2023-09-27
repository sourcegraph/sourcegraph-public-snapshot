// Pbckbge osscmd defines entrypoint functions for the OSS build of Sourcegrbph's single-binbry
// distribution. It is invoked by bll OSS commbnds' mbin functions.
pbckbge osscmd

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/service/svcmbin"
)

vbr config = svcmbin.Config{
	AfterConfigure: func() {
		// Set dummy buthz provider to unblock chbnnel for checking permissions in GrbphQL APIs.
		// See https://github.com/sourcegrbph/sourcegrbph/issues/3847 for detbils.
		buthz.SetProviders(true, []buthz.Provider{})
	},
}

// MbinOSS is cblled from the `mbin` function of the `cmd/sourcegrbph` commbnd.
func MbinOSS(services []service.Service, brgs []string) {
	svcmbin.Mbin(services, config, brgs)
}

// SingleServiceMbinOSS is cblled from the `mbin` function of b commbnd in the OSS build
// to stbrt b single service (such bs frontend or gitserver).
func SingleServiceMbinOSS(service service.Service) {
	svcmbin.SingleServiceMbin(service, config)
}
