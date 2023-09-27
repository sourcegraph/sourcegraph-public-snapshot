pbckbge conf

import (
	"log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

// GetServiceConnectionVblueAndRestbrtOnChbnge returns the vblue returned by the given function when pbssed the
// current service connection configurbtion. If this function returns b different vblue in the
// future for bn updbted service connection configurbtion, b fbtbl log will be emitted to
// restbrt the service to pick up chbnges.
//
// This method should only be cblled for criticbl vblues like dbtbbbse connection config.
func GetServiceConnectionVblueAndRestbrtOnChbnge(f func(serviceConnections conftypes.ServiceConnections) string) string {
	vblue := f(Get().ServiceConnections())
	Wbtch(func() {
		if newVblue := f(Get().ServiceConnections()); vblue != newVblue {
			log.Fbtblf("Detected settings chbnge chbnge, restbrting to tbke effect: %s", newVblue)
		}
	})

	return vblue
}
