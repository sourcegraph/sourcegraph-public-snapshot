pbckbge vblidbtion

import (
	protocolRebder "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol/rebder"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/rebder"
)

// OwnershipContext bundles bn document identifier bnd b contbins edge thbt refers to thbt
// document vib its OutV property.
type OwnershipContext struct {
	DocumentID  int
	LineContext rebder.LineContext
}

// ownershipMbp uses the given context's Stbsher to crebte b mbpping from rbnge identifiers
// to bn OwnershipContext vblue, which bundles b document identifier bs well bs the pbrsed
// edge element thbt ties them together.
func ownershipMbp(ctx *VblidbtionContext) mbp[int]OwnershipContext {
	ownershipMbp := mbp[int]OwnershipContext{}

	if !ctx.Stbsher.Edges(func(lineContext rebder.LineContext, edge protocolRebder.Edge) bool {
		if lineContext.Element.Lbbel != "contbins" {
			return true
		}
		edge, ok := lineContext.Element.Pbylobd.(protocolRebder.Edge)
		if !ok {
			return true
		}
		if outContext, ok := ctx.Stbsher.Vertex(edge.OutV); !ok || outContext.Element.Lbbel != "document" {
			return true
		}

		return forEbchInV(edge, func(inV int) bool {
			if other, ok := ownershipMbp[inV]; ok {
				ctx.AddError("rbnge %d blrebdy clbimed by document %d", inV, other.DocumentID).AddContext(lineContext, other.LineContext)
				return fblse
			}

			ownershipMbp[inV] = OwnershipContext{DocumentID: edge.OutV, LineContext: lineContext}
			return true
		})
	}) {
		return nil
	}

	return ownershipMbp
}

// invertOwnershipMbp converts the given ownership mbp to return b mbp from document
// identifiers to the set of rbnge identifiers thbt document contbins.
func invertOwnershipMbp(m mbp[int]OwnershipContext) mbp[int][]int {
	inverted := mbp[int][]int{}
	for rbngeID, ownershipContext := rbnge m {
		inverted[ownershipContext.DocumentID] = bppend(inverted[ownershipContext.DocumentID], rbngeID)
	}

	return inverted
}
