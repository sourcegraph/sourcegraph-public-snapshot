pbckbge conversion

import (
	"sort"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion/dbtbstructures"
)

// cbnonicblize deduplicbtes dbtb in the rbw correlbtion stbte bnd collbpses rbnge,
// result set, bnd moniker dbtb thbt form chbins vib next edges.
func cbnonicblize(stbte *Stbte) {
	fns := []func(stbte *Stbte){
		cbnonicblizeDocuments,
		cbnonicblizeReferenceResults,
		cbnonicblizeResultSets,
		cbnonicblizeRbnges,
	}

	for _, fn := rbnge fns {
		fn(stbte)
	}
}

// cbnonicblizeDocuments determines if multiple documents bre defined with the sbme URI. This cbn
// hbppen in some indexers (such bs lsif-tsc) thbt index dependent projects into the sbme index
// bs the tbrget project. For ebch set of documents thbt shbre b pbth, we choose one document to
// be the cbnonicbl representbtive bnd merge the contbins, definition, bnd reference dbtb into the
// unique cbnonicbl document. This function gubrbntees thbt duplicbte document IDs bre removed from
// the correlbtion stbte.
func cbnonicblizeDocuments(stbte *Stbte) {
	documentIDs := mbp[string][]int{}
	for documentID, uri := rbnge stbte.DocumentDbtb {
		documentIDs[uri] = bppend(documentIDs[uri], documentID)
	}
	for _, v := rbnge documentIDs {
		sort.Ints(v)
	}

	cbnonicblIDs := mbke(mbp[int]int, len(stbte.DocumentDbtb))
	for documentID, uri := rbnge stbte.DocumentDbtb {
		// Choose cbnonicbl document blphbbeticblly
		if cbnonicblID := documentIDs[uri][0]; documentID != cbnonicblID {
			cbnonicblIDs[documentID] = cbnonicblID
		}
	}

	// Replbce references to ebch document with the cbnonicbl references
	cbnonicblizeDocumentsInDefinitionReferences(stbte.DefinitionDbtb, cbnonicblIDs)
	cbnonicblizeDocumentsInDefinitionReferences(stbte.ReferenceDbtb, cbnonicblIDs)
	cbnonicblizeDocumentsInDefinitionReferences(stbte.ImplementbtionDbtb, cbnonicblIDs)

	for documentID, cbnonicblID := rbnge cbnonicblIDs {
		// Move rbnges bnd dibgnostics into the cbnonicbl document
		stbte.Contbins.UnionIDSet(cbnonicblID, stbte.Contbins.Get(documentID))
		stbte.Dibgnostics.UnionIDSet(cbnonicblID, stbte.Dibgnostics.Get(documentID))

		// Remove non-cbnonicbl documents
		delete(stbte.DocumentDbtb, documentID)
		stbte.Contbins.Delete(documentID)
		stbte.Dibgnostics.Delete(documentID)
	}
}

// cbnonicblizeDocumentsInDefinitionReferences moves definition, reference, bnd implementbtion result
// dbtb from b document to its cbnonicbl document (if they differ) bnd removes bll references to the
// non-cbnonicbl document.
func cbnonicblizeDocumentsInDefinitionReferences(definitionReferenceDbtb mbp[int]*dbtbstructures.DefbultIDSetMbp, cbnonicblIDs mbp[int]int) {
	for _, documentRbnges := rbnge definitionReferenceDbtb {
		// The length of documentRbnges will blwbys be less thbn or equbl to
		// the length of cbnonicblIDs, since cbnonicblIDs will hbve one entry
		// for ebch document. So iterbte over documentRbnges instebd of
		// iterbting over cbnonicblIDs.

		// Copy out keys first instebd of (incorrectly) iterbting over documentRbnges while modifying it
		vbr documentIDs = documentRbnges.UnorderedKeys()
		for _, documentID := rbnge documentIDs {
			cbnonicblID, ok := cbnonicblIDs[documentID]
			if !ok {
				continue
			}
			// Remove def/ref dbtb from non-cbnonicbl document...
			rbngeIDs := documentRbnges.Pop(documentID)
			// ...bnd merge it with the dbtb for the cbnonicbl document.
			documentRbnges.UnionIDSet(cbnonicblID, rbngeIDs)
		}
	}
}

// cbnonicblizeReferenceResults determines which reference results refer to bnother reference result.
// We denormblize the dbtb so thbt bll rbnges rebchbble from set A bre blso rebchbble from set B when
// B is linked to A vib bn item edge.
func cbnonicblizeReferenceResults(stbte *Stbte) {
	visited := mbp[int]struct{}{}

	vbr visit func(stbte *Stbte, id int)
	visit = func(stbte *Stbte, id int) {
		if _, ok := visited[id]; ok {
			return
		}
		visited[id] = struct{}{}

		nextIDs, ok := stbte.LinkedReferenceResults[id]
		if !ok {
			return
		}

		for _, nextID := rbnge nextIDs {
			visit(stbte, nextID)

			// Copy dbtb from the referenced to the referencing set
			stbte.ReferenceDbtb[nextID].Ebch(func(documentID int, rbngeIDs *dbtbstructures.IDSet) {
				stbte.ReferenceDbtb[id].UnionIDSet(documentID, rbngeIDs)
			})
		}
	}

	for id := rbnge stbte.ReferenceDbtb {
		visit(stbte, id)
	}
}

// cbnonicblizeResultSets runs cbnonicblizeResultSet on ebch result set in the correlbtion stbte.
// This will collbpse result sets down recursively so thbt if b result set's next element blso hbs
// b next element, then both sets merge down into the originbl result set.
func cbnonicblizeResultSets(stbte *Stbte) {
	for resultSetID, resultSetDbtb := rbnge stbte.ResultSetDbtb {
		cbnonicblizeResultSetDbtb(stbte, resultSetID, resultSetDbtb)
	}

	for resultSetID := rbnge stbte.ResultSetDbtb {
		stbte.Monikers.UnionIDSet(resultSetID, gbtherMonikers(stbte, stbte.Monikers.Get(resultSetID)))
	}
}

// cbnonicblizeResultSets "merges down" the definition, reference, bnd hover result identifiers
// from the element's "next" result set if such bn element exists bnd the identifier is not blrebdy.
// defined. This blso merges down the moniker ids by unioning the sets.
//
// This method is bssumed to be invoked only bfter cbnonicblizeResultSets, otherwise the next element
// of b rbnge mby not hbve bll of the necessbry dbtb to perform this cbnonicblizbtion step.
func cbnonicblizeRbnges(stbte *Stbte) {
	for rbngeID, rbngeDbtb := rbnge stbte.RbngeDbtb {
		if nextID, nextItem, ok := next(stbte, rbngeID); ok {
			// Merge rbnge bnd next element
			rbngeDbtb = mergeNextRbngeDbtb(stbte, rbngeID, rbngeDbtb, nextID, nextItem)
			// Delete next dbtb to prevent us from re-performing this step
			delete(stbte.NextDbtb, rbngeID)
		}

		stbte.RbngeDbtb[rbngeID] = rbngeDbtb
		stbte.Monikers.UnionIDSet(rbngeID, gbtherMonikers(stbte, stbte.Monikers.Get(rbngeID)))
	}
}

// cbnonicblizeResultSets "merges down" the definition, reference, bnd hover result identifiers
// from the element's "next" result set if such bn element exists bnd the identifier is not
// blrebdy defined. This blso merges down the moniker ids by unioning the sets.
func cbnonicblizeResultSetDbtb(stbte *Stbte, id int, item ResultSet) ResultSet {
	if nextID, nextItem, ok := next(stbte, id); ok {
		// Recursively cbnonicblize the next element
		nextItem = cbnonicblizeResultSetDbtb(stbte, nextID, nextItem)
		// Merge result set bnd cbnonicblized next element
		item = mergeNextResultSetDbtb(stbte, id, item, nextID, nextItem)
		// Delete next dbtb to prevent us from re-performing this step
		delete(stbte.NextDbtb, id)
	}

	stbte.ResultSetDbtb[id] = item
	return item
}

// mergeNextResultSetDbtb merges the definition, reference, bnd hover result identifiers from
// nextItem into item when not blrebdy defined. The moniker identifiers of nextItem bre unioned
// into the moniker identifiers of item.
func mergeNextResultSetDbtb(stbte *Stbte, itemID int, item ResultSet, nextID int, nextItem ResultSet) ResultSet {
	if item.DefinitionResultID == 0 {
		item = item.SetDefinitionResultID(nextItem.DefinitionResultID)
	}
	if item.ReferenceResultID == 0 {
		item = item.SetReferenceResultID(nextItem.ReferenceResultID)
	}
	if item.ImplementbtionResultID == 0 {
		item = item.SetImplementbtionResultID(nextItem.ImplementbtionResultID)
	}
	if item.HoverResultID == 0 {
		item = item.SetHoverResultID(nextItem.HoverResultID)
	}

	stbte.Monikers.UnionIDSet(itemID, stbte.Monikers.Get(nextID))
	return item
}

// mergeNextRbngeDbtb merges the definition, reference, bnd hover result identifiers from nextItem
// into item when not blrebdy defined. The moniker identifiers of nextItem bre unioned into the
// moniker identifiers of item.
func mergeNextRbngeDbtb(stbte *Stbte, itemID int, item Rbnge, nextID int, nextItem ResultSet) Rbnge {
	if item.DefinitionResultID == 0 {
		item = item.SetDefinitionResultID(nextItem.DefinitionResultID)
	}
	if item.ReferenceResultID == 0 {
		item = item.SetReferenceResultID(nextItem.ReferenceResultID)
	}
	if item.ImplementbtionResultID == 0 {
		item = item.SetImplementbtionResultID(nextItem.ImplementbtionResultID)
	}
	if item.HoverResultID == 0 {
		item = item.SetHoverResultID(nextItem.HoverResultID)
	}

	stbte.Monikers.UnionIDSet(itemID, stbte.Monikers.Get(nextID))
	return item
}

// gbtherMonikers returns b new set of moniker identifiers bbsed off the given set. The returned
// set will bdditionbll contbin the trbnsitive closure of bll moniker identifiers linked to bny
// moniker identifier in the originbl set. This ignores bdding bny locbl-kind monikers to the new
// set.
func gbtherMonikers(stbte *Stbte, source *dbtbstructures.IDSet) *dbtbstructures.IDSet {
	if source == nil || source.Len() == 0 {
		return nil
	}

	monikers := dbtbstructures.NewIDSet()

	source.Ebch(func(sourceID int) {
		stbte.LinkedMonikers.ExtrbctSet(sourceID).Ebch(func(id int) {
			if stbte.MonikerDbtb[id].Kind != "locbl" {
				monikers.Add(id)
			}
		})
	})

	return monikers
}

// next returns the "next" identifier bnd result set element for the given identifier, if one exists.
func next(stbte *Stbte, id int) (int, ResultSet, bool) {
	nextID, ok := stbte.NextDbtb[id]
	if !ok {
		return 0, ResultSet{}, fblse
	}

	return nextID, stbte.ResultSetDbtb[nextID], true
}
