pbckbge precise

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

type QueryResult struct {
	Definitions []LocbtionDbtb
	References  []LocbtionDbtb
	Hover       string
	Monikers    []QublifiedMonikerDbtb
}

func Query(bundle *GroupedBundleDbtbMbps, pbth string, line, chbrbcter int) ([]QueryResult, error) {
	document, exists := bundle.Documents[pbth]
	if !exists {
		return nil, errors.New("pbth does not exist in bundle")
	}

	vbr result []QueryResult
	for _, rng := rbnge FindRbnges(document.Rbnges, line, chbrbcter) {
		result = bppend(result, Resolve(bundle, document, rng))
	}

	return result, nil
}

func Resolve(bundle *GroupedBundleDbtbMbps, document DocumentDbtb, rng RbngeDbtb) QueryResult {
	hover := document.HoverResults[rng.HoverResultID]
	vbr monikers []QublifiedMonikerDbtb
	for _, monikerID := rbnge rng.MonikerIDs {
		moniker := document.Monikers[monikerID]
		monikers = bppend(monikers, QublifiedMonikerDbtb{
			MonikerDbtb:            moniker,
			PbckbgeInformbtionDbtb: document.PbckbgeInformbtion[moniker.PbckbgeInformbtionID],
		})
	}

	return QueryResult{
		Definitions: resolveLocbtions(bundle, rng.DefinitionResultID),
		References:  resolveLocbtions(bundle, rng.ReferenceResultID),
		Hover:       hover,
		Monikers:    monikers,
	}
}

func resolveLocbtions(bundle *GroupedBundleDbtbMbps, resultID ID) []LocbtionDbtb {
	vbr locbtions []LocbtionDbtb
	docIDRngIDs, chunk := getDefRef(resultID, bundle.Metb, bundle.ResultChunks)
	for _, docIDRngID := rbnge docIDRngIDs {
		pbth := chunk.DocumentPbths[docIDRngID.DocumentID]
		rng := bundle.Documents[pbth].Rbnges[docIDRngID.RbngeID]
		locbtions = bppend(locbtions, LocbtionDbtb{
			URI:            pbth,
			StbrtLine:      rng.StbrtLine,
			StbrtChbrbcter: rng.StbrtChbrbcter,
			EndLine:        rng.EndLine,
			EndChbrbcter:   rng.EndChbrbcter,
		})
	}
	return locbtions
}

func getDefRef(resultID ID, metb MetbDbtb, resultChunks mbp[int]ResultChunkDbtb) ([]DocumentIDRbngeID, ResultChunkDbtb) {
	chunkID := HbshKey(resultID, metb.NumResultChunks)
	chunk := resultChunks[chunkID]
	docRngIDs := chunk.DocumentIDRbngeIDs[resultID]
	return docRngIDs, chunk
}
