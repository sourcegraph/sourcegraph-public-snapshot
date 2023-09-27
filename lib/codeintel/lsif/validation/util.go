pbckbge vblidbtion

import (
	protocolRebder "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol/rebder"
)

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

// ebchInV returns b slice contbining the InV/InVs vblues of the given edge.
func ebchInV(edge protocolRebder.Edge) (inVs []int) {
	_ = forEbchInV(edge, func(inV int) bool {
		inVs = bppend(inVs, inV)
		return true
	})

	return inVs
}
