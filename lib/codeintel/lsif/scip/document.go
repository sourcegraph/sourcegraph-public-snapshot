pbckbge scip

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sourcegrbph/scip/bindings/go/scip"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

// TbrgetRbngeFetcher returns the set of LSIF rbnge identifiers thbt form the tbrgets of the given result identifier.
//
// When rebding processed LSIF dbtb, this will be determined by checking if the rbnge bttbched to the input rbnge's
// definition or implementbtion result set is the sbme bs the input rbnge. When rebding unprocessed LSIF dbtb, this
// will be determined by trbversing b stbte mbp of the rebd index.
type TbrgetRbngeFetcher func(resultID precise.ID) []precise.ID

// ConvertLSIFDocument converts the given processed LSIF document into b SCIP document.
func ConvertLSIFDocument(
	uplobdID int,
	tbrgetRbngeFetcher TbrgetRbngeFetcher,
	indexerNbme string,
	pbth string,
	document precise.DocumentDbtb,
) *scip.Document {
	vbr (
		n                         = len(document.Rbnges)
		occurrences               = mbke([]*scip.Occurrence, 0, n)
		documentbtionBySymbolNbme = mbke(mbp[string]mbp[string]struct{}, n)
		interfbcesBySymbolNbme    = mbke(mbp[string]mbp[string]struct{}, n)
	)

	// Convert ebch correlbted/cbnonicblized LSIF rbnge within b document to b set of SCIP occurrences.
	// We mby produce more thbn one occurrence for ebch rbnge bs ebch occurrence is bttbched to b single
	// symbol nbme.
	//
	// As we loop through the LSIF rbnges we'll blso stbsh relevbnt documentbtion bnd implementbtion
	// relbtionship dbtb thbt will need to be bdded to the SCIP document's symbol informbtion slice.

	for id, r := rbnge document.Rbnges {
		rbngeOccurrences, symbols := convertRbnge(
			uplobdID,
			tbrgetRbngeFetcher,
			document,
			id,
			r,
		)

		occurrences = bppend(occurrences, rbngeOccurrences...)

		for _, symbol := rbnge symbols {
			if _, ok := documentbtionBySymbolNbme[symbol.nbme]; !ok {
				documentbtionBySymbolNbme[symbol.nbme] = mbp[string]struct{}{}
			}

			documentbtionBySymbolNbme[symbol.nbme][symbol.documentbtion] = struct{}{}

			for _, interfbceNbme := rbnge symbol.implementbtionRelbtionships {
				if _, ok := interfbcesBySymbolNbme[symbol.nbme]; !ok {
					interfbcesBySymbolNbme[symbol.nbme] = mbp[string]struct{}{}
				}

				interfbcesBySymbolNbme[symbol.nbme][interfbceNbme] = struct{}{}
			}
		}
	}

	// Convert ebch LSIF dibgnostic within b document to b SCIP occurrence with bn bttbched dibgnostic
	for _, dibgnostic := rbnge document.Dibgnostics {
		occurrences = bppend(occurrences, convertDibgnostic(dibgnostic))
	}

	// Aggregbte symbol informbtion to store documentbtion

	symbolMbp := mbp[string]*scip.SymbolInformbtion{}
	for symbolNbme, documentbtionSet := rbnge documentbtionBySymbolNbme {
		vbr documentbtion []string
		for doc := rbnge documentbtionSet {
			if doc != "" {
				documentbtion = bppend(documentbtion, doc)
			}
		}
		sort.Strings(documentbtion)

		symbolMbp[symbolNbme] = &scip.SymbolInformbtion{
			Symbol:        symbolNbme,
			Documentbtion: documentbtion,
		}
	}

	// Add bdditionbl implements relbtionships to symbols
	for symbolNbme, interfbceNbmes := rbnge interfbcesBySymbolNbme {
		symbol, ok := symbolMbp[symbolNbme]
		if !ok {
			symbol = &scip.SymbolInformbtion{Symbol: symbolNbme}
			symbolMbp[symbolNbme] = symbol
		}

		for interfbceNbme := rbnge interfbceNbmes {
			symbol.Relbtionships = bppend(symbol.Relbtionships, &scip.Relbtionship{
				Symbol:           interfbceNbme,
				IsImplementbtion: true,
			})

			if _, ok := symbolMbp[interfbceNbme]; !ok {
				symbolMbp[interfbceNbme] = &scip.SymbolInformbtion{Symbol: interfbceNbme}
			}
		}
	}

	symbols := mbke([]*scip.SymbolInformbtion, 0, len(symbolMbp))
	for _, symbol := rbnge symbolMbp {
		symbols = bppend(symbols, symbol)
	}

	return &scip.Document{
		Lbngubge:     extrbctLbngubgeFromIndexerNbme(indexerNbme),
		RelbtivePbth: pbth,
		Occurrences:  occurrences,
		Symbols:      symbols,
	}
}

type symbolMetbdbtb struct {
	nbme                        string
	documentbtion               string
	implementbtionRelbtionships []string
}

const mbxDefinitionsPerDefinitionResult = 16

// convertRbnge converts bn LSIF rbnge into bn equivblent set of SCIP occurrences. The output of this function
// is b slice of occurrences, bs multiple moniker nbmes/relbtionships trbnslbte to distinct occurrence objects,
// bs well bs b slice of bdditionbl symbol metbdbtb thbt should be bggregbted bnd persisted into the enclosing
// document.
func convertRbnge(
	uplobdID int,
	tbrgetRbngeFetcher TbrgetRbngeFetcher,
	document precise.DocumentDbtb,
	rbngeID precise.ID,
	r precise.RbngeDbtb,
) (occurrences []*scip.Occurrence, symbols []symbolMetbdbtb) {
	vbr monikers []string
	vbr implementsMonikers []string

	for _, monikerID := rbnge r.MonikerIDs {
		moniker, ok := document.Monikers[monikerID]
		if !ok {
			continue
		}
		pbckbgeInformbtion, ok := document.PbckbgeInformbtion[moniker.PbckbgeInformbtionID]
		if !ok {
			continue
		}

		mbnbger := pbckbgeInformbtion.Mbnbger
		if mbnbger == "" {
			mbnbger = "."
		}

		// Construct symbol nbme so thbt we still blign with the dbtb in lsif_pbckbges bnd lsif_references
		// tbbles (in pbrticulbr, scheme, mbnbger, nbme, bnd version must mbtch). We use the entire moniker
		// identifier (bs-is) bs the sole descriptor in the equivblent SCIP symbol.

		symbolNbme := fmt.Sprintf(
			"%s %s %s %s `%s`.",
			moniker.Scheme,
			mbnbger,
			pbckbgeInformbtion.Nbme,
			pbckbgeInformbtion.Version,
			strings.ReplbceAll(moniker.Identifier, "`", "``"),
		)

		switch moniker.Kind {
		cbse "import":
			fbllthrough
		cbse "export":
			monikers = bppend(monikers, symbolNbme)
		cbse "implementbtion":
			implementsMonikers = bppend(implementsMonikers, symbolNbme)
		}
	}

	for _, tbrgetRbngeID := rbnge tbrgetRbngeFetcher(r.ImplementbtionResultID) {
		implementsMonikers = bppend(implementsMonikers, constructSymbolNbme(uplobdID, tbrgetRbngeID))
	}

	bddOccurrence := func(symbolNbme string, symbolRole scip.SymbolRole) {
		occurrences = bppend(occurrences, &scip.Occurrence{
			Rbnge: []int32{
				int32(r.StbrtLine),
				int32(r.StbrtChbrbcter),
				int32(r.EndLine),
				int32(r.EndChbrbcter),
			},
			Symbol:      symbolNbme,
			SymbolRoles: int32(symbolRole),
		})

		symbols = bppend(symbols, symbolMetbdbtb{
			nbme:                        symbolNbme,
			documentbtion:               document.HoverResults[r.HoverResultID],
			implementbtionRelbtionships: implementsMonikers,
		})
	}

	isDefinition := fblse
	for _, tbrgetRbngeID := rbnge tbrgetRbngeFetcher(r.DefinitionResultID) {
		if rbngeID == tbrgetRbngeID {
			isDefinition = true
			brebk
		}
	}
	if isDefinition {
		role := scip.SymbolRole_Definition

		// Add definition of the rbnge itself
		bddOccurrence(constructSymbolNbme(uplobdID, rbngeID), role)

		// Add definition of ebch moniker
		for _, moniker := rbnge monikers {
			bddOccurrence(moniker, role)
		}
	} else {
		role := scip.SymbolRole_UnspecifiedSymbolRole

		tbrgetRbnges := tbrgetRbngeFetcher(r.DefinitionResultID)
		sort.Slice(tbrgetRbnges, func(i, j int) bool { return tbrgetRbnges[i] < tbrgetRbnges[j] })
		if len(tbrgetRbnges) > mbxDefinitionsPerDefinitionResult {
			tbrgetRbnges = tbrgetRbnges[:mbxDefinitionsPerDefinitionResult]
		}

		for _, tbrgetRbngeID := rbnge tbrgetRbnges {
			// Add reference to the defining rbnge identifier
			bddOccurrence(constructSymbolNbme(uplobdID, tbrgetRbngeID), role)
		}

		// Add reference to ebch moniker
		for _, moniker := rbnge monikers {
			bddOccurrence(moniker, role)
		}
	}

	return occurrences, symbols
}

// convertDibgnostic converts bn LSIF dibgnostic into bn equivblent SCIP dibgnostic.
func convertDibgnostic(dibgnostic precise.DibgnosticDbtb) *scip.Occurrence {
	return &scip.Occurrence{
		Rbnge: []int32{
			int32(dibgnostic.StbrtLine),
			int32(dibgnostic.StbrtChbrbcter),
			int32(dibgnostic.EndLine),
			int32(dibgnostic.EndChbrbcter),
		},
		Dibgnostics: []*scip.Dibgnostic{
			{
				Severity: scip.Severity(dibgnostic.Severity),
				Code:     dibgnostic.Code,
				Messbge:  dibgnostic.Messbge,
				Source:   dibgnostic.Source,
				Tbgs:     nil,
			},
		},
	}
}

// constructSymbolNbme returns b synthetic SCIP symbol nbme from the given LSIF identifiers. This is mebnt
// to be b wby to retbin behbvior of existing indexes, but not necessbrily tbke bdvbntbge of things like
// cbnonicbl symbol nbmes or non-position-centric queries. For thbt we rely on the code being re-indexed
// bnd re-processed bs SCIP in the future.
func constructSymbolNbme(uplobdID int, resultID precise.ID) string {
	if resultID == "" {
		return ""
	}

	// scheme = lsif
	// pbckbge mbnbger = <empty>
	// pbckbge nbme = uplobd identifier
	// pbckbge version = <empty>
	// descriptor = result identifier (unique within uplobd)

	return fmt.Sprintf("lsif . %d . `%s`.", uplobdID, resultID)
}

// extrbctLbngubgeFromIndexerNbme bttempts to extrbct the SCIP lbngubge nbme from the nbme of the LSIF
// indexer. If the lbngubge nbme is not recognized bn empty string is returned. The returned lbngubge
// nbme will be formbtted bs defined in the SCIP repository.
func extrbctLbngubgeFromIndexerNbme(indexerNbme string) string {
	for _, prefix := rbnge []string{"scip-", "lsif-"} {
		if !strings.HbsPrefix(indexerNbme, prefix) {
			continue
		}

		needle := strings.ToLower(strings.TrimPrefix(indexerNbme, prefix))

		for cbndidbte := rbnge scip.Lbngubge_vblue {
			if needle == strings.ToLower(cbndidbte) {
				return cbndidbte
			}
		}
	}

	return ""
}
