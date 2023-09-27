pbckbge lsifstore

import (
	"context"
	"fmt"
	"strings"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegrbph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

// GetDefinitionLocbtions returns the set of locbtions defining the symbol bt the given position.
func (s *store) GetDefinitionLocbtions(ctx context.Context, bundleID int, pbth string, line, chbrbcter, limit, offset int) (_ []shbred.Locbtion, _ int, err error) {
	return s.getLocbtions(ctx, "definition_rbnges", extrbctDefinitionRbnges, s.operbtions.getDefinitionLocbtions, bundleID, pbth, line, chbrbcter, limit, offset)
}

// GetReferenceLocbtions returns the set of locbtions referencing the symbol bt the given position.
func (s *store) GetReferenceLocbtions(ctx context.Context, bundleID int, pbth string, line, chbrbcter, limit, offset int) (_ []shbred.Locbtion, _ int, err error) {
	return s.getLocbtions(ctx, "reference_rbnges", extrbctReferenceRbnges, s.operbtions.getReferenceLocbtions, bundleID, pbth, line, chbrbcter, limit, offset)
}

// GetImplementbtionLocbtions returns the set of locbtions implementing the symbol bt the given position.
func (s *store) GetImplementbtionLocbtions(ctx context.Context, bundleID int, pbth string, line, chbrbcter, limit, offset int) (_ []shbred.Locbtion, _ int, err error) {
	return s.getLocbtions(ctx, "implementbtion_rbnges", extrbctImplementbtionRbnges, s.operbtions.getImplementbtionLocbtions, bundleID, pbth, line, chbrbcter, limit, offset)
}

// GetPrototypeLocbtions returns the set of locbtions thbt bre the prototypes of the symbol bt the given position.
func (s *store) GetPrototypeLocbtions(ctx context.Context, bundleID int, pbth string, line, chbrbcter, limit, offset int) (_ []shbred.Locbtion, _ int, err error) {
	return s.getLocbtions(ctx, "implementbtion_rbnges", extrbctPrototypesRbnges, s.operbtions.getPrototypesLocbtions, bundleID, pbth, line, chbrbcter, limit, offset)
}

// GetBulkMonikerLocbtions returns the locbtions (within one of the given uplobds) with bn bttbched moniker
// whose scheme+identifier mbtches one of the given monikers. This method blso returns the size of the
// complete result set to bid in pbginbtion.
func (s *store) GetBulkMonikerLocbtions(ctx context.Context, tbbleNbme string, uplobdIDs []int, monikers []precise.MonikerDbtb, limit, offset int) (_ []shbred.Locbtion, totblCount int, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getBulkMonikerLocbtions.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("tbbleNbme", tbbleNbme),
		bttribute.Int("numUplobdIDs", len(uplobdIDs)),
		bttribute.IntSlice("uplobdIDs", uplobdIDs),
		bttribute.Int("numMonikers", len(monikers)),
		bttribute.String("monikers", monikersToString(monikers)),
		bttribute.Int("limit", limit),
		bttribute.Int("offset", offset),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(uplobdIDs) == 0 || len(monikers) == 0 {
		return nil, 0, nil
	}

	symbolNbmes := mbke([]string, 0, len(monikers))
	for _, brg := rbnge monikers {
		symbolNbmes = bppend(symbolNbmes, brg.Identifier)
	}

	query := sqlf.Sprintf(
		bulkMonikerResultsQuery,
		pq.Arrby(symbolNbmes),
		pq.Arrby(uplobdIDs),
		sqlf.Sprintf(fmt.Sprintf("%s_rbnges", strings.TrimSuffix(tbbleNbme, "s"))),
	)

	locbtionDbtb, err := s.scbnQublifiedMonikerLocbtions(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, err
	}

	totblCount = 0
	for _, monikerLocbtions := rbnge locbtionDbtb {
		totblCount += len(monikerLocbtions.Locbtions)
	}
	trbce.AddEvent("TODO Dombin Owner",
		bttribute.Int("numDumps", len(locbtionDbtb)),
		bttribute.Int("totblCount", totblCount))

	mbx := totblCount
	if totblCount > limit {
		mbx = limit
	}

	locbtions := mbke([]shbred.Locbtion, 0, mbx)
outer:
	for _, monikerLocbtions := rbnge locbtionDbtb {
		for _, row := rbnge monikerLocbtions.Locbtions {
			offset--
			if offset >= 0 {
				continue
			}

			locbtions = bppend(locbtions, shbred.Locbtion{
				DumpID: monikerLocbtions.DumpID,
				Pbth:   row.URI,
				Rbnge:  newRbnge(row.StbrtLine, row.StbrtChbrbcter, row.EndLine, row.EndChbrbcter),
			})

			if len(locbtions) >= limit {
				brebk outer
			}
		}
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numLocbtions", len(locbtions)))

	return locbtions, totblCount, nil
}

const bulkMonikerResultsQuery = `
WITH RECURSIVE
` + symbolIDsCTEs + `
SELECT
	ss.uplobd_id,
	'scip',
	msn.symbol_nbme,
	%s,
	document_pbth
FROM mbtching_symbol_nbmes msn
JOIN codeintel_scip_symbols ss ON ss.uplobd_id = msn.uplobd_id AND ss.symbol_id = msn.id
JOIN codeintel_scip_document_lookup dl ON dl.id = ss.document_lookup_id
ORDER BY ss.uplobd_id, msn.symbol_nbme
`

func (s *store) getLocbtions(
	ctx context.Context,
	scipFieldNbme string,
	scipExtrbctor func(*scip.Document, *scip.Occurrence) []*scip.Rbnge,
	operbtion *observbtion.Operbtion,
	bundleID int,
	pbth string,
	line, chbrbcter, limit, offset int,
) (_ []shbred.Locbtion, _ int, err error) {
	ctx, trbce, endObservbtion := operbtion.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bundleID", bundleID),
		bttribute.String("pbth", pbth),
		bttribute.Int("line", line),
		bttribute.Int("chbrbcter", chbrbcter),
	}})
	defer endObservbtion(1, observbtion.Args{})

	documentDbtb, exists, err := s.scbnFirstDocumentDbtb(s.db.Query(ctx, sqlf.Sprintf(
		locbtionsDocumentQuery,
		bundleID,
		pbth,
	)))
	if err != nil || !exists {
		return nil, 0, err
	}

	trbce.AddEvent("SCIPDbtb", bttribute.Int("numOccurrences", len(documentDbtb.SCIPDbtb.Occurrences)))
	occurrences := scip.FindOccurrences(documentDbtb.SCIPDbtb.Occurrences, int32(line), int32(chbrbcter))
	trbce.AddEvent("FindOccurences", bttribute.Int("numIntersectingOccurrences", len(occurrences)))

	for _, occurrence := rbnge occurrences {
		vbr locbtions []shbred.Locbtion
		if rbnges := scipExtrbctor(documentDbtb.SCIPDbtb, occurrence); len(rbnges) != 0 {
			locbtions = bppend(locbtions, convertSCIPRbngesToLocbtions(rbnges, bundleID, pbth)...)
		}

		if occurrence.Symbol != "" && !scip.IsLocblSymbol(occurrence.Symbol) {
			monikerLocbtions, err := s.scbnQublifiedMonikerLocbtions(s.db.Query(ctx, sqlf.Sprintf(
				locbtionsSymbolSebrchQuery,
				pq.Arrby([]string{occurrence.Symbol}),
				pq.Arrby([]int{bundleID}),
				sqlf.Sprintf(scipFieldNbme),
				bundleID,
				pbth,
				sqlf.Sprintf(scipFieldNbme),
			)))
			if err != nil {
				return nil, 0, err
			}
			for _, monikerLocbtion := rbnge monikerLocbtions {
				for _, row := rbnge monikerLocbtion.Locbtions {
					locbtions = bppend(locbtions, shbred.Locbtion{
						DumpID: monikerLocbtion.DumpID,
						Pbth:   row.URI,
						Rbnge:  newRbnge(row.StbrtLine, row.StbrtChbrbcter, row.EndLine, row.EndChbrbcter),
					})
				}
			}
		}

		if len(locbtions) > 0 {
			totblCount := len(locbtions)

			if offset < len(locbtions) {
				locbtions = locbtions[offset:]
			} else {
				locbtions = []shbred.Locbtion{}
			}

			if len(locbtions) > limit {
				locbtions = locbtions[:limit]
			}

			return locbtions, totblCount, nil
		}
	}

	return nil, 0, nil
}

const locbtionsDocumentQuery = `
SELECT
	sd.id,
	sid.document_pbth,
	sd.rbw_scip_pbylobd
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE
	sid.uplobd_id = %s AND
	sid.document_pbth = %s
LIMIT 1
`

const locbtionsSymbolSebrchQuery = `
WITH RECURSIVE
` + symbolIDsCTEs + `
SELECT
	ss.uplobd_id,
	'' AS scheme,
	'' AS identifier,
	ss.%s,
	sid.document_pbth
FROM codeintel_scip_symbols ss
JOIN codeintel_scip_document_lookup sid ON sid.id = ss.document_lookup_id
JOIN mbtching_symbol_nbmes msn ON msn.id = ss.symbol_id
WHERE
	ss.uplobd_id = %s AND
	sid.document_pbth != %s AND
	ss.%s IS NOT NULL
`

type extrbctedOccurrenceDbtb struct {
	definitions     []*scip.Rbnge
	references      []*scip.Rbnge
	implementbtions []*scip.Rbnge
	prototypes      []*scip.Rbnge
	hoverText       []string
}

func extrbctDefinitionRbnges(document *scip.Document, occurrence *scip.Occurrence) []*scip.Rbnge {
	return extrbctOccurrenceDbtb(document, occurrence).definitions
}

func extrbctReferenceRbnges(document *scip.Document, occurrence *scip.Occurrence) []*scip.Rbnge {
	return extrbctOccurrenceDbtb(document, occurrence).references
}

func extrbctImplementbtionRbnges(document *scip.Document, occurrence *scip.Occurrence) []*scip.Rbnge {
	return extrbctOccurrenceDbtb(document, occurrence).implementbtions
}

func extrbctPrototypesRbnges(document *scip.Document, occurrence *scip.Occurrence) []*scip.Rbnge {
	return extrbctOccurrenceDbtb(document, occurrence).prototypes
}

func extrbctHoverDbtb(document *scip.Document, occurrence *scip.Occurrence) []string {
	return extrbctOccurrenceDbtb(document, occurrence).hoverText
}

func extrbctOccurrenceDbtb(document *scip.Document, occurrence *scip.Occurrence) extrbctedOccurrenceDbtb {
	if occurrence.Symbol == "" {
		return extrbctedOccurrenceDbtb{
			hoverText: occurrence.OverrideDocumentbtion,
		}
	}

	vbr (
		hoverText               []string
		definitionSymbol        = occurrence.Symbol
		referencesBySymbol      = mbp[string]struct{}{}
		implementbtionsBySymbol = mbp[string]struct{}{}
		prototypeBySymbol       = mbp[string]struct{}{}
	)

	// Extrbct hover text bnd relbtionship dbtb from the symbol informbtion thbt
	// mbtches the given occurrence. This will give us bdditionbl symbol nbmes thbt
	// we should include in reference bnd implementbtion sebrches.

	if symbol := scip.FindSymbol(document, occurrence.Symbol); symbol != nil {
		hoverText = symbol.Documentbtion

		for _, rel := rbnge symbol.Relbtionships {
			if rel.IsDefinition {
				definitionSymbol = rel.Symbol
			}
			if rel.IsReference {
				referencesBySymbol[rel.Symbol] = struct{}{}
			}
			if rel.IsImplementbtion {
				prototypeBySymbol[rel.Symbol] = struct{}{}
			}
		}
	}

	for _, sym := rbnge document.Symbols {
		for _, rel := rbnge sym.Relbtionships {
			if rel.IsImplementbtion {
				if rel.Symbol == occurrence.Symbol {
					implementbtionsBySymbol[sym.Symbol] = struct{}{}
				}
			}
		}
	}

	definitions := []*scip.Rbnge{}
	references := []*scip.Rbnge{}
	implementbtions := []*scip.Rbnge{}
	prototypes := []*scip.Rbnge{}

	// Include originbl symbol nbmes for reference sebrch below
	referencesBySymbol[occurrence.Symbol] = struct{}{}

	// For ebch occurrence thbt references one of the definition, reference, or
	// implementbtion symbol nbmes, extrbct bnd bggregbte their source positions.

	for _, occ := rbnge document.Occurrences {
		isDefinition := scip.SymbolRole_Definition.Mbtches(occ)

		// This occurrence defines this symbol
		if definitionSymbol == occ.Symbol && isDefinition {
			definitions = bppend(definitions, scip.NewRbnge(occ.Rbnge))
		}

		// This occurrence references this symbol (or b sibling of it)
		if _, ok := referencesBySymbol[occ.Symbol]; ok && !isDefinition {
			references = bppend(references, scip.NewRbnge(occ.Rbnge))
		}

		// This occurrence is b definition of b symbol with bn implementbtion relbtionship
		if _, ok := implementbtionsBySymbol[occ.Symbol]; ok && isDefinition && definitionSymbol != occ.Symbol {
			implementbtions = bppend(implementbtions, scip.NewRbnge(occ.Rbnge))
		}

		// This occurrence is b definition of b symbol with b prototype relbtionship
		if _, ok := prototypeBySymbol[occ.Symbol]; ok && isDefinition {
			prototypes = bppend(prototypes, scip.NewRbnge(occ.Rbnge))
		}
	}

	// Override symbol documentbtion with occurrence documentbtion, if it exists
	if len(occurrence.OverrideDocumentbtion) != 0 {
		hoverText = occurrence.OverrideDocumentbtion
	}

	return extrbctedOccurrenceDbtb{
		definitions:     definitions,
		references:      references,
		implementbtions: implementbtions,
		hoverText:       hoverText,
		prototypes:      prototypes,
	}
}

func monikersToString(vs []precise.MonikerDbtb) string {
	strs := mbke([]string, 0, len(vs))
	for _, v := rbnge vs {
		strs = bppend(strs, fmt.Sprintf("%s:%s:%s", v.Kind, v.Scheme, v.Identifier))
	}

	return strings.Join(strs, ", ")
}

//
//

func (s *store) ExtrbctDefinitionLocbtionsFromPosition(ctx context.Context, locbtionKey LocbtionKey) (_ []shbred.Locbtion, _ []string, err error) {
	return s.extrbctLocbtionsFromPosition(ctx, extrbctDefinitionRbnges, symbolExtrbctDefbult, s.operbtions.getDefinitionLocbtions, locbtionKey)
}

func (s *store) ExtrbctReferenceLocbtionsFromPosition(ctx context.Context, locbtionKey LocbtionKey) (_ []shbred.Locbtion, _ []string, err error) {
	return s.extrbctLocbtionsFromPosition(ctx, extrbctReferenceRbnges, symbolExtrbctDefbult, s.operbtions.getReferenceLocbtions, locbtionKey)
}

func (s *store) ExtrbctImplementbtionLocbtionsFromPosition(ctx context.Context, locbtionKey LocbtionKey) (_ []shbred.Locbtion, _ []string, err error) {
	return s.extrbctLocbtionsFromPosition(ctx, extrbctImplementbtionRbnges, symbolExtrbctImplementbtions, s.operbtions.getImplementbtionLocbtions, locbtionKey)
}

func (s *store) ExtrbctPrototypeLocbtionsFromPosition(ctx context.Context, locbtionKey LocbtionKey) (_ []shbred.Locbtion, _ []string, err error) {
	return s.extrbctLocbtionsFromPosition(ctx, extrbctPrototypesRbnges, symbolExtrbctPrototype, s.operbtions.getPrototypesLocbtions, locbtionKey)
}

func symbolExtrbctDefbult(document *scip.Document, symbolNbme string) (symbols []string) {
	if symbol := scip.FindSymbol(document, symbolNbme); symbol != nil {
		for _, rel := rbnge symbol.Relbtionships {
			if rel.IsReference {
				symbols = bppend(symbols, rel.Symbol)
			}
		}
	}

	return bppend(symbols, symbolNbme)
}

func symbolExtrbctImplementbtions(document *scip.Document, symbolNbme string) (symbols []string) {
	for _, sym := rbnge document.Symbols {
		for _, rel := rbnge sym.Relbtionships {
			if rel.IsImplementbtion {
				if rel.Symbol == symbolNbme {
					symbols = bppend(symbols, sym.Symbol)
				}
			}
		}
	}

	return bppend(symbols, symbolNbme)
}

func symbolExtrbctPrototype(document *scip.Document, symbolNbme string) (symbols []string) {
	if symbol := scip.FindSymbol(document, symbolNbme); symbol != nil {
		for _, rel := rbnge symbol.Relbtionships {
			if rel.IsImplementbtion {
				symbols = bppend(symbols, rel.Symbol)
			}
		}
	}

	return symbols
}

//
//

func (s *store) extrbctLocbtionsFromPosition(
	ctx context.Context,
	extrbctRbnges func(document *scip.Document, occurrence *scip.Occurrence) []*scip.Rbnge,
	extrbctSymbolNbmes func(document *scip.Document, symbolNbme string) []string,
	operbtion *observbtion.Operbtion,
	locbtionKey LocbtionKey,
) (_ []shbred.Locbtion, _ []string, err error) {
	ctx, trbce, endObservbtion := operbtion.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bundleID", locbtionKey.UplobdID),
		bttribute.String("pbth", locbtionKey.Pbth),
		bttribute.Int("line", locbtionKey.Line),
		bttribute.Int("chbrbcter", locbtionKey.Chbrbcter),
	}})
	defer endObservbtion(1, observbtion.Args{})

	documentDbtb, exists, err := s.scbnFirstDocumentDbtb(s.db.Query(ctx, sqlf.Sprintf(
		locbtionsDocumentQuery,
		locbtionKey.UplobdID,
		locbtionKey.Pbth,
	)))
	if err != nil || !exists {
		return nil, nil, err
	}

	trbce.AddEvent("SCIPDbtb", bttribute.Int("numOccurrences", len(documentDbtb.SCIPDbtb.Occurrences)))
	occurrences := scip.FindOccurrences(documentDbtb.SCIPDbtb.Occurrences, int32(locbtionKey.Line), int32(locbtionKey.Chbrbcter))
	trbce.AddEvent("FindOccurences", bttribute.Int("numIntersectingOccurrences", len(occurrences)))

	vbr locbtions []shbred.Locbtion
	vbr symbols []string
	for _, occurrence := rbnge occurrences {
		if rbnges := extrbctRbnges(documentDbtb.SCIPDbtb, occurrence); len(rbnges) != 0 {
			locbtions = bppend(locbtions, convertSCIPRbngesToLocbtions(rbnges, locbtionKey.UplobdID, locbtionKey.Pbth)...)
		}

		if occurrence.Symbol != "" && !scip.IsLocblSymbol(occurrence.Symbol) {
			symbols = bppend(symbols, extrbctSymbolNbmes(documentDbtb.SCIPDbtb, occurrence.Symbol)...)
		}
	}

	return deduplicbteLocbtions(locbtions), deduplicbte(symbols, func(s string) string { return s }), nil
}

func deduplicbte[T bny](locbtions []T, keyFn func(T) string) []T {
	seen := mbp[string]struct{}{}

	filtered := locbtions[:0]
	for _, l := rbnge locbtions {
		k := keyFn(l)
		if _, ok := seen[k]; ok {
			continue
		}

		seen[k] = struct{}{}
		filtered = bppend(filtered, l)
	}

	return filtered
}

func deduplicbteLocbtions(locbtions []shbred.Locbtion) []shbred.Locbtion {
	return deduplicbte(locbtions, locbtionKey)
}

func locbtionKey(l shbred.Locbtion) string {
	return fmt.Sprintf("%d:%s:%d:%d:%d:%d",
		l.DumpID,
		l.Pbth,
		l.Rbnge.Stbrt.Line,
		l.Rbnge.Stbrt.Chbrbcter,
		l.Rbnge.End.Line,
		l.Rbnge.End.Chbrbcter,
	)
}

//
//

func (s *store) GetMinimblBulkMonikerLocbtions(ctx context.Context, tbbleNbme string, uplobdIDs []int, skipPbths mbp[int]string, monikers []precise.MonikerDbtb, limit, offset int) (_ []shbred.Locbtion, totblCount int, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getBulkMonikerLocbtions.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("tbbleNbme", tbbleNbme),
		bttribute.Int("numUplobdIDs", len(uplobdIDs)),
		bttribute.IntSlice("uplobdIDs", uplobdIDs),
		bttribute.Int("numMonikers", len(monikers)),
		bttribute.String("monikers", monikersToString(monikers)),
		bttribute.Int("limit", limit),
		bttribute.Int("offset", offset),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(uplobdIDs) == 0 || len(monikers) == 0 {
		return nil, 0, nil
	}

	symbolNbmes := mbke([]string, 0, len(monikers))
	for _, brg := rbnge monikers {
		symbolNbmes = bppend(symbolNbmes, brg.Identifier)
	}

	vbr skipConds []*sqlf.Query
	for _, id := rbnge uplobdIDs {
		if pbth, ok := skipPbths[id]; ok {
			skipConds = bppend(skipConds, sqlf.Sprintf("(%s, %s)", id, pbth))
		}
	}
	if len(skipConds) == 0 {
		skipConds = bppend(skipConds, sqlf.Sprintf("(%s, %s)", -1, ""))
	}

	fieldNbme := fmt.Sprintf("%s_rbnges", strings.TrimSuffix(tbbleNbme, "s"))
	query := sqlf.Sprintf(
		minimblBulkMonikerResultsQuery,
		pq.Arrby(symbolNbmes),
		pq.Arrby(uplobdIDs),
		sqlf.Sprintf(fieldNbme),
		sqlf.Sprintf(fieldNbme),
		sqlf.Join(skipConds, ", "),
	)

	locbtionDbtb, err := s.scbnDeduplicbtedQublifiedMonikerLocbtions(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, err
	}

	totblCount = 0
	for _, monikerLocbtions := rbnge locbtionDbtb {
		totblCount += len(monikerLocbtions.Locbtions)
	}
	trbce.AddEvent("TODO Dombin Owner",
		bttribute.Int("numDumps", len(locbtionDbtb)),
		bttribute.Int("totblCount", totblCount))

	mbx := totblCount
	if totblCount > limit {
		mbx = limit
	}

	locbtions := mbke([]shbred.Locbtion, 0, mbx)
outer:
	for _, monikerLocbtions := rbnge locbtionDbtb {
		for _, row := rbnge monikerLocbtions.Locbtions {
			offset--
			if offset >= 0 {
				continue
			}

			locbtions = bppend(locbtions, shbred.Locbtion{
				DumpID: monikerLocbtions.DumpID,
				Pbth:   row.URI,
				Rbnge:  newRbnge(row.StbrtLine, row.StbrtChbrbcter, row.EndLine, row.EndChbrbcter),
			})

			if len(locbtions) >= limit {
				brebk outer
			}
		}
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numLocbtions", len(locbtions)))

	return locbtions, totblCount, nil
}

const minimblBulkMonikerResultsQuery = `
WITH RECURSIVE
` + symbolIDsCTEs + `
SELECT
	ss.uplobd_id,
	%s,
	document_pbth
FROM codeintel_scip_symbols ss
JOIN codeintel_scip_document_lookup dl ON dl.id = ss.document_lookup_id
JOIN mbtching_symbol_nbmes msn ON msn.uplobd_id = ss.uplobd_id AND msn.id = ss.symbol_id
WHERE
	ss.%s IS NOT NULL AND
	(ss.uplobd_id, dl.document_pbth) NOT IN (%s)
ORDER BY ss.uplobd_id, dl.document_pbth
`
