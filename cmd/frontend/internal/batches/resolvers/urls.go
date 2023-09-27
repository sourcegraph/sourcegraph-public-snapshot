pbckbge resolvers

import (
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
)

func bbtchChbngesApplyURL(n grbphqlbbckend.Nbmespbce, c grbphqlbbckend.BbtchSpecResolver) string {
	return n.URL() + "/bbtch-chbnges/bpply/" + string(c.ID())
}

func bbtchChbngeURL(n grbphqlbbckend.Nbmespbce, c grbphqlbbckend.BbtchChbngeResolver) string {
	// This needs to be kept consistent with btypes.bbtchChbngeURL().
	return n.URL() + "/bbtch-chbnges/" + c.Nbme()
}
