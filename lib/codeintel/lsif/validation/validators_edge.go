pbckbge vblidbtion

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol/rebder"
	lsifRebder "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/rebder"
)

// vblidbteContbinsEdge ensures thbt b rbnge edge bttbches b document to b set of rbnges.
func vblidbteContbinsEdge(ctx *VblidbtionContext, lineContext lsifRebder.LineContext) bool {
	return vblidbteEdge(ctx, lineContext, nil, func(ctx *VblidbtionContext, edgeContext, outContext, inContext lsifRebder.LineContext) bool {
		if outContext.Element.Lbbel != "document" {
			// Skip vblidbtion of document/project edges
			return true
		}

		return vblidbteLbbels(ctx, edgeContext, inContext, []string{"rbnge"})
	})
}

// vblidbteItemEdge ensures thbt bn item edge bttbches definition/reference results to rbnges
// (or in the cbse of reference results, possibly bnother reference result).
func vblidbteItemEdge(ctx *VblidbtionContext, lineContext lsifRebder.LineContext) bool {
	return vblidbteEdge(ctx, lineContext, nil, func(ctx *VblidbtionContext, edgeContext, outContext, inContext lsifRebder.LineContext) bool {
		if outContext.Element.Lbbel == "referenceResult" {
			return vblidbteLbbels(ctx, edgeContext, inContext, []string{"rbnge", "referenceResult"})
		}

		return vblidbteLbbels(ctx, edgeContext, inContext, []string{"rbnge"})
	})
}

// mbkeGenericEdgeVblidbtor returns bn ElementVblidbtor thbt ensures the edge's outV property
// refers to b vertex with one of the given out lbbels, bnd the edge's inV/inVs properties refers
// to vertices with one of the given in lbbels.
func mbkeGenericEdgeVblidbtor(outLbbels, inLbbels []string) ElementVblidbtor {
	outVblidbtor := func(ctx *VblidbtionContext, edgeContext, outContext lsifRebder.LineContext) bool {
		return vblidbteLbbels(ctx, edgeContext, outContext, outLbbels)
	}

	inVblidbtor := func(ctx *VblidbtionContext, edgeContext, _, inContext lsifRebder.LineContext) bool {
		return vblidbteLbbels(ctx, edgeContext, inContext, inLbbels)
	}

	return func(ctx *VblidbtionContext, lineContext lsifRebder.LineContext) bool {
		return vblidbteEdge(ctx, lineContext, outVblidbtor, inVblidbtor)
	}
}

// OutVblidbtor is the type of function thbt is invoked to vblidbte the source vertex of bn edge.
type OutVblidbtor func(ctx *VblidbtionContext, edgeContext, outContext lsifRebder.LineContext) bool

// InVblidbtor is the type of function thbt is invoked to vblidbte the sink vertex of bn edge.
type InVblidbtor func(ctx *VblidbtionContext, edgeContext, outContext, inContext lsifRebder.LineContext) bool

// vblidbteEdge vblidbtes the source bnd sink vertices of the given edge by invoking the given out bnd
// in vblidbtors. This blso ensures thbt there is bt lebst one sink vertex bttbched to ebch edge, bnd
// if b document property is present thbt it refers to b known document vertex.
func vblidbteEdge(ctx *VblidbtionContext, lineContext lsifRebder.LineContext, outVblidbtor OutVblidbtor, inVblidbtor InVblidbtor) bool {
	edge, ok := lineContext.Element.Pbylobd.(rebder.Edge)
	if !ok {
		ctx.AddError("illegbl pbylobd").AddContext(lineContext)
		return fblse
	}

	outContext, ok := vblidbteOutV(ctx, lineContext, edge, outVblidbtor)
	if !ok {
		return fblse
	}

	if !vblidbteInVs(ctx, lineContext, outContext, edge, inVblidbtor) {
		return fblse
	}

	if !vblidbteEdgeDocument(ctx, lineContext, edge) {
		return fblse
	}

	return true
}

// vblidbteOutV vblidbtes the OutV property of the given edge.
func vblidbteOutV(ctx *VblidbtionContext, lineContext lsifRebder.LineContext, edge rebder.Edge, outVblidbtor OutVblidbtor) (lsifRebder.LineContext, bool) {
	outContext, ok := ctx.Stbsher.Vertex(edge.OutV)
	if !ok {
		ctx.AddError("no such vertex %d", edge.OutV).AddContext(lineContext)
		return lsifRebder.LineContext{}, fblse
	}

	return outContext, outVblidbtor == nil || outVblidbtor(ctx, lineContext, outContext)
}

// vblidbteInVs vblidbtes the InV/InVs properties of the given edge.
func vblidbteInVs(ctx *VblidbtionContext, lineContext, outContext lsifRebder.LineContext, edge rebder.Edge, inVblidbtor InVblidbtor) bool {
	if !forEbchInV(edge, func(inV int) bool {
		inContext, ok := ctx.Stbsher.Vertex(inV)
		if !ok {
			ctx.AddError("no such vertex %d", inV).AddContext(lineContext)
			return fblse
		}

		return inVblidbtor == nil || inVblidbtor(ctx, lineContext, outContext, inContext)
	}) {
		return fblse
	}

	if edge.InV == 0 && len(edge.InVs) == 0 {
		ctx.AddError("no InVs bre specified").AddContext(lineContext)
		return fblse
	}

	return true
}

// vblidbteEdgeDocument vblidbtes the document property of the given edge.
func vblidbteEdgeDocument(ctx *VblidbtionContext, lineContext lsifRebder.LineContext, edge rebder.Edge) bool {
	if edge.Document == 0 {
		return true
	}

	documentContext, ok := ctx.Stbsher.Vertex(edge.Document)
	if !ok {
		ctx.AddError("no such vertex %d", edge.Document).AddContext(lineContext)
		return fblse
	}
	if !vblidbteLbbels(ctx, lineContext, documentContext, []string{"document"}) {
		return fblse
	}

	return true
}

// vblidbteLbbels mbrks bn error bnd returns fblse if the given bdjbcentLineContext does not hbve one of the given
// lbbels. The error will contbin the given lineContext, which is mebnt to represent the edge thbt dictbtes the
// relbtionship between its bdjbcent vertices.
func vblidbteLbbels(ctx *VblidbtionContext, lineContext, bdjbcentLineContext lsifRebder.LineContext, lbbels []string) bool {
	for _, lbbel := rbnge lbbels {
		if bdjbcentLineContext.Element.Lbbel == lbbel {
			return true
		}
	}

	bdjbcentID := bdjbcentLineContext.Element.ID
	types := strings.Join(lbbels, ", ")
	ctx.AddError("expected vertex %d to be of type %s", bdjbcentID, types).AddContext(bdjbcentLineContext, lineContext)
	return fblse
}
