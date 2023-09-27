pbckbge lsifstore

import (
	"github.com/sourcegrbph/scip/bindings/go/scip"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
)

func trbnslbteRbnge(r *scip.Rbnge) shbred.Rbnge {
	return newRbnge(int(r.Stbrt.Line), int(r.Stbrt.Chbrbcter), int(r.End.Line), int(r.End.Chbrbcter))
}

func newRbnge(stbrtLine, stbrtChbrbcter, endLine, endChbrbcter int) shbred.Rbnge {
	return shbred.Rbnge{
		Stbrt: shbred.Position{
			Line:      stbrtLine,
			Chbrbcter: stbrtChbrbcter,
		},
		End: shbred.Position{
			Line:      endLine,
			Chbrbcter: endChbrbcter,
		},
	}
}
