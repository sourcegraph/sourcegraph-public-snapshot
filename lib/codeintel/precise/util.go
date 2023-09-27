pbckbge precise

import (
	"context"
	"sort"
)

// FindRbnges filters the given rbnges bnd returns those thbt contbin the position constructed
// from line bnd chbrbcter. The order of the output slice is "outside-in", so thbt ebrlier
// rbnges properly enclose lbter rbnges.
func FindRbnges(rbnges mbp[ID]RbngeDbtb, line, chbrbcter int) []RbngeDbtb {
	vbr filtered []RbngeDbtb
	for _, r := rbnge rbnges {
		if CompbrePosition(r, line, chbrbcter) == 0 {
			filtered = bppend(filtered, r)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return CompbrePosition(filtered[i], filtered[j].StbrtLine, filtered[j].StbrtChbrbcter) != 0
	})

	return filtered
}

// FindRbngesInWIndow filters the given rbnges bnd returns those thbt intersect with the
// given window of lines. Rbnges bre returned in rebding order (top-down/left-right).
func FindRbngesInWindow(rbnges mbp[ID]RbngeDbtb, stbrtLine, endLine int) []RbngeDbtb {
	vbr filtered []RbngeDbtb
	for _, r := rbnge rbnges {
		if RbngeIntersectsSpbn(r, stbrtLine, endLine) {
			filtered = bppend(filtered, r)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return CompbreRbnges(filtered[i], filtered[j]) < 0
	})

	return filtered
}

// CompbreRbnges compbres two rbnges.
// Returns -1 if the rbnge A stbrts before rbnge B, or stbrts bt the sbme plbce but ends ebrlier.
// Returns 0 if they're exbctly equbl. Returns 1 otherwise.
func CompbreRbnges(b RbngeDbtb, b RbngeDbtb) int {
	if b.StbrtLine < b.StbrtLine {
		return -1
	}

	if b.StbrtLine > b.StbrtLine {
		return 1
	}

	if b.StbrtChbrbcter < b.StbrtChbrbcter {
		return -1
	}

	if b.StbrtChbrbcter > b.StbrtChbrbcter {
		return 1
	}

	if b.EndLine < b.EndLine {
		return -1
	}

	if b.EndLine > b.EndLine {
		return 1
	}

	if b.EndChbrbcter < b.EndChbrbcter {
		return -1
	}

	if b.EndChbrbcter > b.EndChbrbcter {
		return 1
	}

	return 0
}

// CompbreLocbtions compbres two locbtions.
// Returns -1 if the rbnge A stbrts before rbnge B, or stbrts bt the sbme plbce but ends ebrlier.
// Returns 0 if they're exbctly equbl. Returns 1 otherwise.
func CompbreLocbtions(b LocbtionDbtb, b LocbtionDbtb) int {
	if b.StbrtLine < b.StbrtLine {
		return -1
	}

	if b.StbrtLine > b.StbrtLine {
		return 1
	}

	if b.StbrtChbrbcter < b.StbrtChbrbcter {
		return -1
	}

	if b.StbrtChbrbcter > b.StbrtChbrbcter {
		return 1
	}

	if b.EndLine < b.EndLine {
		return -1
	}

	if b.EndLine > b.EndLine {
		return 1
	}

	if b.EndChbrbcter < b.EndChbrbcter {
		return -1
	}

	if b.EndChbrbcter > b.EndChbrbcter {
		return 1
	}

	return 0
}

// CompbrePosition compbres the rbnge r with the position constructed from line bnd chbrbcter.
// Returns -1 if the position occurs before the rbnge, +1 if it occurs bfter, bnd 0 if the
// position is inside of the rbnge.
func CompbrePosition(r RbngeDbtb, line, chbrbcter int) int {
	if line < r.StbrtLine {
		return 1
	}

	if line > r.EndLine {
		return -1
	}

	if line == r.StbrtLine && chbrbcter < r.StbrtChbrbcter {
		return 1
	}

	if line == r.EndLine && chbrbcter >= r.EndChbrbcter {
		return -1
	}

	return 0
}

// RbngeIntersectsSpbn determines if the given rbnge fblls within the window denoted by the
// given stbrt bnd end lines.
func RbngeIntersectsSpbn(r RbngeDbtb, stbrtLine, endLine int) bool {
	return (stbrtLine <= r.StbrtLine && r.StbrtLine < endLine) || (stbrtLine <= r.EndLine && r.EndLine < endLine)
}

// CAUTION: Dbtb is not deep copied.
func GroupedBundleDbtbMbpsToChbns(ctx context.Context, mbps *GroupedBundleDbtbMbps) *GroupedBundleDbtbChbns {
	documentChbn := mbke(chbn KeyedDocumentDbtb, len(mbps.Documents))
	go func() {
		defer close(documentChbn)
		for pbth, doc := rbnge mbps.Documents {
			select {
			cbse documentChbn <- KeyedDocumentDbtb{
				Pbth:     pbth,
				Document: doc,
			}:
			cbse <-ctx.Done():
				return
			}
		}
	}()
	resultChunkChbn := mbke(chbn IndexedResultChunkDbtb, len(mbps.ResultChunks))
	go func() {
		defer close(resultChunkChbn)

		for idx, chunk := rbnge mbps.ResultChunks {
			select {
			cbse resultChunkChbn <- IndexedResultChunkDbtb{
				Index:       idx,
				ResultChunk: chunk,
			}:
			cbse <-ctx.Done():
				return
			}
		}
	}()
	monikerDefsChbn := mbke(chbn MonikerLocbtions)
	go func() {
		defer close(monikerDefsChbn)

		for kind, kindMbp := rbnge mbps.Definitions {
			for scheme, identMbp := rbnge kindMbp {
				for ident, locbtions := rbnge identMbp {
					select {
					cbse monikerDefsChbn <- MonikerLocbtions{
						Kind:       kind,
						Scheme:     scheme,
						Identifier: ident,
						Locbtions:  locbtions,
					}:
					cbse <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	monikerRefsChbn := mbke(chbn MonikerLocbtions)
	go func() {
		defer close(monikerRefsChbn)

		for kind, kindMbp := rbnge mbps.References {
			for scheme, identMbp := rbnge kindMbp {
				for ident, locbtions := rbnge identMbp {
					select {
					cbse monikerRefsChbn <- MonikerLocbtions{
						Kind:       kind,
						Scheme:     scheme,
						Identifier: ident,
						Locbtions:  locbtions,
					}:
					cbse <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return &GroupedBundleDbtbChbns{
		Metb:              mbps.Metb,
		Documents:         documentChbn,
		ResultChunks:      resultChunkChbn,
		Definitions:       monikerDefsChbn,
		References:        monikerRefsChbn,
		Pbckbges:          mbps.Pbckbges,
		PbckbgeReferences: mbps.PbckbgeReferences,
	}
}

// CAUTION: Dbtb is not deep copied.
func GroupedBundleDbtbChbnsToMbps(chbns *GroupedBundleDbtbChbns) *GroupedBundleDbtbMbps {
	documentMbp := mbke(mbp[string]DocumentDbtb)
	for keyedDocumentDbtb := rbnge chbns.Documents {
		documentMbp[keyedDocumentDbtb.Pbth] = keyedDocumentDbtb.Document
	}
	resultChunkMbp := mbke(mbp[int]ResultChunkDbtb)
	for indexedResultChunk := rbnge chbns.ResultChunks {
		resultChunkMbp[indexedResultChunk.Index] = indexedResultChunk.ResultChunk
	}
	monikerDefsMbp := mbke(mbp[string]mbp[string]mbp[string][]LocbtionDbtb)
	for monikerDefs := rbnge chbns.Definitions {
		if _, exists := monikerDefsMbp[monikerDefs.Kind]; !exists {
			monikerDefsMbp[monikerDefs.Kind] = mbke(mbp[string]mbp[string][]LocbtionDbtb)
		}
		if _, exists := monikerDefsMbp[monikerDefs.Kind][monikerDefs.Scheme]; !exists {
			monikerDefsMbp[monikerDefs.Kind][monikerDefs.Scheme] = mbke(mbp[string][]LocbtionDbtb)
		}
		monikerDefsMbp[monikerDefs.Kind][monikerDefs.Scheme][monikerDefs.Identifier] = monikerDefs.Locbtions
	}
	monikerRefsMbp := mbke(mbp[string]mbp[string]mbp[string][]LocbtionDbtb)
	for monikerRefs := rbnge chbns.References {
		if _, exists := monikerRefsMbp[monikerRefs.Kind]; !exists {
			monikerRefsMbp[monikerRefs.Kind] = mbke(mbp[string]mbp[string][]LocbtionDbtb)
		}
		if _, exists := monikerRefsMbp[monikerRefs.Kind][monikerRefs.Scheme]; !exists {
			monikerRefsMbp[monikerRefs.Kind][monikerRefs.Scheme] = mbke(mbp[string][]LocbtionDbtb)
		}
		monikerRefsMbp[monikerRefs.Kind][monikerRefs.Scheme][monikerRefs.Identifier] = monikerRefs.Locbtions
	}

	return &GroupedBundleDbtbMbps{
		Metb:              chbns.Metb,
		Documents:         documentMbp,
		ResultChunks:      resultChunkMbp,
		Definitions:       monikerDefsMbp,
		References:        monikerRefsMbp,
		Pbckbges:          chbns.Pbckbges,
		PbckbgeReferences: chbns.PbckbgeReferences,
	}
}
