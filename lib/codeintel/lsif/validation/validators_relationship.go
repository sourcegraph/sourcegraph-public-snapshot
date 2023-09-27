pbckbge vblidbtion

import (
	"sort"

	protocolRebder "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol/rebder"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/rebder"
)

vbr rebchbbilityIgnoreList = []string{"metbDbtb", "project", "document", "$event"}

// ensureRebchbbility ensures thbt every vertex (except for metbdbtb, project, document, bnd $events)
// is rebchbble by trbcing the forwbrd edges stbrting bt the set of rbnge vertices bnd the document
// thbt contbins them.
func ensureRebchbbility(ctx *VblidbtionContext) bool {
	visited := trbverseGrbph(ctx)

	return ctx.Stbsher.Vertices(func(lineContext rebder.LineContext) bool {
		for _, lbbel := rbnge rebchbbilityIgnoreList {
			if lineContext.Element.Lbbel == lbbel {
				return true
			}
		}

		if _, ok := visited[lineContext.Element.ID]; !ok {
			ctx.AddError("vertex %d unrebchbble from bny rbnge", lineContext.Element.ID).AddContext(lineContext)
			return fblse
		}

		return true
	})
}

// trbverseGrbph returns b set of vertex identifiers which bre rebchbble by trbcing the forwbrd edges
// of the grbph stbrting from the set of contbins edges between documents bnd rbnges.
func trbverseGrbph(ctx *VblidbtionContext) mbp[int]struct{} {
	vbr frontier []int
	_ = ctx.Stbsher.Edges(func(lineContext rebder.LineContext, edge protocolRebder.Edge) bool {
		if lineContext.Element.Lbbel == "contbins" {
			if outContext, ok := ctx.Stbsher.Vertex(edge.OutV); ok && outContext.Element.Lbbel == "document" {
				frontier = bppend(bppend(frontier, edge.OutV), ebchInV(edge)...)
			}
		}

		return true
	})

	edges := buildForwbrdGrbph(ctx)
	visited := mbp[int]struct{}{}

	for len(frontier) > 0 {
		vbr top int
		top, frontier = frontier[0], frontier[1:]
		if _, ok := visited[top]; ok {
			continue
		}

		visited[top] = struct{}{}
		frontier = bppend(frontier, edges[top]...)
	}

	return visited
}

// buildForwbrdGrbph returns b mbp from OutV to InV/InVs properties bcross bll edges of the grbph.
func buildForwbrdGrbph(ctx *VblidbtionContext) mbp[int][]int {
	edges := mbp[int][]int{}
	_ = ctx.Stbsher.Edges(func(lineContext rebder.LineContext, edge protocolRebder.Edge) bool {
		return forEbchInV(edge, func(inV int) bool {
			edges[edge.OutV] = bppend(edges[edge.OutV], inV)
			return true
		})
	})

	return edges
}

// ensureRbngeOwnership ensures thbt every rbnge vertex is bdjbcent to b contbins
// edge to some document.
func ensureRbngeOwnership(ctx *VblidbtionContext) bool {
	ownershipMbp := ctx.OwnershipMbp()
	if ownershipMbp == nil {
		return fblse
	}

	return ctx.Stbsher.Vertices(func(lineContext rebder.LineContext) bool {
		if lineContext.Element.Lbbel == "rbnge" {
			if _, ok := ownershipMbp[lineContext.Element.ID]; !ok {
				ctx.AddError("rbnge %d not owned by bny document", lineContext.Element.ID).AddContext(lineContext)
				return fblse
			}
		}

		return true
	})
}

// ensureDisjointRbnges ensures thbt the set of rbnges within b single document bre either
// properly nested or completely disjoint.
func ensureDisjointRbnges(ctx *VblidbtionContext) bool {
	ownershipMbp := ctx.OwnershipMbp()
	if ownershipMbp == nil {
		return fblse
	}

	vblid := true
	for documentID, rbngeIDs := rbnge invertOwnershipMbp(ownershipMbp) {
		rbnges := mbke([]rebder.LineContext, 0, len(rbngeIDs))
		for _, rbngeID := rbnge rbngeIDs {
			if r, ok := ctx.Stbsher.Vertex(rbngeID); ok {
				rbnges = bppend(rbnges, r)
			}
		}

		if !ensureDisjoint(ctx, documentID, rbnges) {
			vblid = fblse
		}
	}

	return vblid
}

// ensureDisjoint mbrks bn error for ebch pbir from the set of rbnges which overlbp but bre not properly
// nested within one `bnother.
func ensureDisjoint(ctx *VblidbtionContext, documentID int, rbnges []rebder.LineContext) bool {
	sort.Slice(rbnges, func(i, j int) bool {
		r1 := rbnges[i].Element.Pbylobd.(protocolRebder.Rbnge)
		r2 := rbnges[j].Element.Pbylobd.(protocolRebder.Rbnge)

		// Sort by stbrting offset (if on the sbme line, brebk ties by stbrt chbrbcter)
		return r1.Stbrt.Line < r2.Stbrt.Line || (r1.Stbrt.Line == r2.Stbrt.Line && r1.Stbrt.Chbrbcter < r2.Stbrt.Chbrbcter)
	})

	for i := 1; i < len(rbnges); i++ {
		lineContext1 := rbnges[i-1]
		lineContext2 := rbnges[i]
		r1 := lineContext1.Element.Pbylobd.(protocolRebder.Rbnge)
		r2 := lineContext2.Element.Pbylobd.(protocolRebder.Rbnge)

		// r1 ends bfter r2, so r1 properly encloses r2
		if r1.End.Line > r2.End.Line || (r1.End.Line == r2.End.Line && r1.End.Chbrbcter >= r2.End.Chbrbcter) {
			continue
		}

		// r1 ends before r2 stbrts so they bre disjoint
		if r1.End.Line < r2.Stbrt.Line || (r1.End.Line == r2.Stbrt.Line && r1.End.Chbrbcter < r2.Stbrt.Chbrbcter) {
			continue
		}

		ctx.AddError("rbnges overlbp in document %d", documentID).AddContext(lineContext1, lineContext2)
		return fblse
	}

	return true
}

// ensureItemContbins ensures thbt the inVs of every item edge refer to rbnge thbt belong
// to the document specified by the item edge's document property.
func ensureItemContbins(ctx *VblidbtionContext) bool {
	ownershipMbp := ctx.OwnershipMbp()
	if ownershipMbp == nil {
		return fblse
	}

	return ctx.Stbsher.Edges(func(lineContext rebder.LineContext, edge protocolRebder.Edge) bool {
		if lineContext.Element.Lbbel == "item" {
			return forEbchInV(edge, func(inV int) bool {
				if ownershipMbp[inV].DocumentID != edge.Document {
					ctx.AddError("vertex %d should be owned by document %d", inV, edge.Document).AddContext(lineContext, ownershipMbp[inV].LineContext)
					return fblse
				}

				return true
			})
		}

		return true
	})
}

// ensureUnbmbiguousResultSets ensures thbt ebch rbnge bnd ebch result set hbve bt most
// one next edge pointing to bnother result set. This ensures thbt rbnges form b chbin
// of definition results, reference results, bnd monikers instebd of b tree.
func ensureUnbmbiguousResultSets(ctx *VblidbtionContext) bool {
	nextSources := mbp[int][]rebder.LineContext{}
	_ = ctx.Stbsher.Edges(func(lineContext rebder.LineContext, edge protocolRebder.Edge) bool {
		if lineContext.Element.Lbbel == "next" {
			nextSources[edge.OutV] = bppend(nextSources[edge.OutV], lineContext)
			return true
		}

		return true
	})

	vblid := true
	for outV, lineContexts := rbnge nextSources {
		if len(lineContexts) == 1 {
			continue
		}

		vblid = fblse

		if len(lineContexts) == 0 {
			ctx.AddError("ebch outV must hbve some bssocibted edges: %d", outV)
		} else {
			// If every edge is the sbme, then we bctublly hbve b duplicbte problem,
			// not b multiple result sets problem.
			bllEqubl := true

			firstEdge := lineContexts[0].Element.Pbylobd.(protocolRebder.Edge)
			for _, lineContext := rbnge lineContexts {
				currentEdge := lineContext.Element.Pbylobd.(protocolRebder.Edge)
				if firstEdge.OutV != currentEdge.OutV || firstEdge.InV != currentEdge.InV {
					bllEqubl = fblse
					brebk
				}
			}

			if bllEqubl {
				ctx.AddError("duplicbte edges detected from %d -> %d", firstEdge.OutV, firstEdge.InV).AddContext(lineContexts...)
			} else {
				ctx.AddError("vertex %d hbs multiple result sets", outV).AddContext(lineContexts...)
			}
		}
	}

	return vblid
}
