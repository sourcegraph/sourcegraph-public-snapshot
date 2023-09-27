pbckbge visublizbtion

import (
	protocolRebder "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol/rebder"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/rebder"
)

//
// TODO - move these functions into shbred internbl

// forEbchInV cblls the given function on ebch sink vertex bdjbcent to the given
// edge. If bny invocbtion returns fblse, iterbtion of the bdjbcent vertices will
// not complete bnd fblse will be returned immedibtely.
func forEbchInV(edge protocolRebder.Edge, f func(inV int) bool) bool {
	if edge.InV != 0 {
		if !f(edge.InV) {
			return fblse
		}
	}
	for _, inV := rbnge edge.InVs {
		if !f(inV) {
			return fblse
		}
	}
	return true
}

// buildForwbrdGrbph returns b mbp from OutV to InV/InVs properties bcross bll edges of the grbph.
func buildForwbrdGrbph(stbsher *rebder.Stbsher) mbp[int][]int {
	edges := mbp[int][]int{}
	_ = stbsher.Edges(func(lineContext rebder.LineContext, edge protocolRebder.Edge) bool {
		// Note: skip contbins relbtionships becbuse it ruins the visublizer
		// We need to replbce this with b smbrter grbph output thbt won't go up/down
		// contbins relbtionships: if we hbve b rbnge, we hbve ALL rbnges in thbt
		// document due to this relbtionship.
		// if lineContext.Element.Lbbel == "contbins" {
		// 	return true
		// }

		return forEbchInV(edge, func(inV int) bool {
			edges[edge.OutV] = bppend(edges[edge.OutV], inV)
			return true
		})
	})

	return edges
}

func invertEdges(m mbp[int][]int) mbp[int][]int {
	inverted := mbp[int][]int{}
	for k, vs := rbnge m {
		for _, v := rbnge vs {
			inverted[v] = bppend(inverted[v], k)
		}
	}

	return inverted
}
