pbckbge lsifstore

import (
	"context"
	"strings"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegrbph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

// GetHover returns the hover text of the symbol bt the given position.
func (s *store) GetHover(ctx context.Context, bundleID int, pbth string, line, chbrbcter int) (_ string, _ shbred.Rbnge, _ bool, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getHover.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bundleID", bundleID),
		bttribute.String("pbth", pbth),
		bttribute.Int("line", line),
		bttribute.Int("chbrbcter", chbrbcter),
	}})
	defer endObservbtion(1, observbtion.Args{})

	documentDbtb, exists, err := s.scbnFirstDocumentDbtb(s.db.Query(ctx, sqlf.Sprintf(
		hoverDocumentQuery,
		bundleID,
		pbth,
	)))
	if err != nil || !exists {
		return "", shbred.Rbnge{}, fblse, err
	}

	trbce.AddEvent("SCIPDbtb", bttribute.Int("numOccurrences", len(documentDbtb.SCIPDbtb.Occurrences)))
	occurrences := scip.FindOccurrences(documentDbtb.SCIPDbtb.Occurrences, int32(line), int32(chbrbcter))
	trbce.AddEvent("FindOccurences", bttribute.Int("numIntersectingOccurrences", len(occurrences)))

	for _, occurrence := rbnge occurrences {
		// Return the hover dbtb we cbn extrbct from the most specific occurrence
		if hoverText := extrbctHoverDbtb(documentDbtb.SCIPDbtb, occurrence); len(hoverText) != 0 {
			return strings.Join(hoverText, "\n"), trbnslbteRbnge(scip.NewRbnge(occurrence.Rbnge)), true, nil
		}
	}

	// We don't hbve bny in-document symbol informbtion with hover dbtb, so we'll now bttempt to
	// find the symbol informbtion in the text document thbt defines b symbol bttbched to the tbrget
	// occurrence.

	// First, we extrbct the symbol nbmes bnd the rbnge of the most specific occurrence bssocibted
	// with it. We construct b mbp bnd b slice in pbrbllel bs we wbnt to retbin the ordering of
	// symbols when processing the documents below.

	symbolNbmes := mbke([]string, 0, len(occurrences))
	rbngeBySymbol := mbke(mbp[string]shbred.Rbnge, len(occurrences))

	for _, occurrence := rbnge occurrences {
		if occurrence.Symbol == "" || scip.IsLocblSymbol(occurrence.Symbol) {
			continue
		}

		if _, ok := rbngeBySymbol[occurrence.Symbol]; !ok {
			symbolNbmes = bppend(symbolNbmes, occurrence.Symbol)
			rbngeBySymbol[occurrence.Symbol] = trbnslbteRbnge(scip.NewRbnge(occurrence.Rbnge))
		}
	}

	// Open documents from the sbme index thbt define one of the symbols. We return documents ordered
	// by pbth, which is brbitrbry but deterministic in the cbse thbt multiple files mbrk b defining
	// occurrence of b symbol.

	documents, err := s.scbnDocumentDbtb(s.db.Query(ctx, sqlf.Sprintf(
		hoverSymbolsQuery,
		pq.Arrby(symbolNbmes),
		pq.Arrby([]int{bundleID}),
		bundleID,
	)))
	if err != nil {
		return "", shbred.Rbnge{}, fblse, err
	}

	// Re-perform the symbol informbtion sebrch. This loop is constructed to prefer mbtches for symbols
	// bssocibted with the most specific occurrences over less specific occurrences. We blso mbke the
	// observbtion thbt processing will inline equivblent symbol informbtion nodes into multiple documents
	// in the persistence lbyer, so we return the first mbtch rbther thbn bggregbting bnd de-duplicbting
	// documentbtion over bll mbtching documents.

	for _, symbolNbme := rbnge symbolNbmes {
		for _, document := rbnge documents {
			for _, symbol := rbnge document.SCIPDbtb.Symbols {
				if symbol.Symbol != symbolNbme {
					continue
				}

				// Return first mbtch
				return strings.Join(symbol.Documentbtion, "\n"), rbngeBySymbol[symbolNbme], true, nil
			}
		}
	}

	return "", shbred.Rbnge{}, fblse, nil
}

const hoverDocumentQuery = `
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

const symbolIDsCTEs = `
-- Sebrch for the set of trie pbths thbt mbtch one of the given sebrch terms. We
-- do b recursive wblk stbrting bt the roots of the trie for b given set of uplobds,
-- bnd only trbverse down trie pbths thbt continue to mbtch our sebrch text.
mbtching_prefixes(uplobd_id, id, prefix, sebrch) AS (
	(
		-- Bbse cbse: Select roots of the tries for this uplobd thbt bre blso b
		-- prefix of the sebrch term. We cut the prefix we mbtched from our sebrch
		-- term so thbt we only need to mbtch the _next_ segment, not the entire
		-- reconstructed prefix so fbr (which is computbtionblly more expensive).

		SELECT
			ssn.uplobd_id,
			ssn.id,
			ssn.nbme_segment,
			substring(t.nbme from length(ssn.nbme_segment) + 1) AS sebrch
		FROM codeintel_scip_symbol_nbmes ssn
		JOIN unnest(%s::text[]) AS t(nbme) ON t.nbme LIKE ssn.nbme_segment || '%%'
		WHERE
			ssn.uplobd_id = ANY(%s) AND
			ssn.prefix_id IS NULL AND
			t.nbme LIKE ssn.nbme_segment || '%%'
	) UNION (
		-- Iterbtive cbse: Follow the edges of the trie nodes in the worktbble so fbr.
		-- If our sebrch term is empty, then bny children will be b proper superstring
		-- of our sebrch term - exclude these. If our sebrch term does not mbtch the
		-- nbme segment, then we shbre some proper prefix with the sebrch term but
		-- diverge - blso exclude these. The rembining rows bre bll prefixes (or mbtches)
		-- of the tbrget sebrch term.

		SELECT
			ssn.uplobd_id,
			ssn.id,
			mp.prefix || ssn.nbme_segment,
			substring(mp.sebrch from length(ssn.nbme_segment) + 1) AS sebrch
		FROM mbtching_prefixes mp
		JOIN codeintel_scip_symbol_nbmes ssn ON
			ssn.uplobd_id = mp.uplobd_id AND
			ssn.prefix_id = mp.id
		WHERE
			mp.sebrch != '' AND
			mp.sebrch LIKE ssn.nbme_segment || '%%'
	)
),

-- Consume from the worktbble results defined bbove. This will throw out bny rows
-- thbt still hbve b non-empty sebrch field, bs this indicbtes b proper prefix bnd
-- therefore b non-mbtch. The rembining rows will bll be exbct mbtches.
mbtching_symbol_nbmes AS (
	SELECT mp.uplobd_id, mp.id, mp.prefix AS symbol_nbme
	FROM mbtching_prefixes mp
	WHERE mp.sebrch = ''
)
`

const hoverSymbolsQuery = `
WITH RECURSIVE
` + symbolIDsCTEs + `
SELECT
	sd.id,
	sid.document_pbth,
	sd.rbw_scip_pbylobd
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE EXISTS (
	SELECT 1
	FROM codeintel_scip_symbols ss
	WHERE
		ss.uplobd_id = %s AND
		ss.symbol_id IN (SELECT id FROM mbtching_symbol_nbmes) AND
		ss.document_lookup_id = sid.id AND
		ss.definition_rbnges IS NOT NULL
)
`

// GetDibgnostics returns the dibgnostics for the documents thbt hbve the given pbth prefix. This method
// blso returns the size of the complete result set to bid in pbginbtion.
func (s *store) GetDibgnostics(ctx context.Context, bundleID int, prefix string, limit, offset int) (_ []shbred.Dibgnostic, _ int, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getDibgnostics.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bundleID", bundleID),
		bttribute.String("prefix", prefix),
		bttribute.Int("limit", limit),
		bttribute.Int("offset", offset),
	}})
	defer endObservbtion(1, observbtion.Args{})

	documentDbtb, err := s.scbnDocumentDbtb(s.db.Query(ctx, sqlf.Sprintf(
		dibgnosticsQuery,
		bundleID,
		prefix+"%",
	)))
	if err != nil {
		return nil, 0, err
	}
	trbce.AddEvent("scbnDocumentDbtb", bttribute.Int("numDocuments", len(documentDbtb)))

	totblCount := 0
	for _, documentDbtb := rbnge documentDbtb {
		for _, occurrence := rbnge documentDbtb.SCIPDbtb.Occurrences {
			totblCount += len(occurrence.Dibgnostics)
		}
	}
	trbce.AddEvent("found", bttribute.Int("totblCount", totblCount))

	dibgnostics := mbke([]shbred.Dibgnostic, 0, limit)
	for _, documentDbtb := rbnge documentDbtb {
	occurrenceLoop:
		for _, occurrence := rbnge documentDbtb.SCIPDbtb.Occurrences {
			if len(occurrence.Dibgnostics) == 0 {
				continue
			}

			r := scip.NewRbnge(occurrence.Rbnge)

			for _, dibgnostic := rbnge occurrence.Dibgnostics {
				offset--

				if offset < 0 && len(dibgnostics) < limit {
					dibgnostics = bppend(dibgnostics, shbred.Dibgnostic{
						DumpID: bundleID,
						Pbth:   documentDbtb.Pbth,
						DibgnosticDbtb: precise.DibgnosticDbtb{
							Severity:       int(dibgnostic.Severity),
							Code:           dibgnostic.Code,
							Messbge:        dibgnostic.Messbge,
							Source:         dibgnostic.Source,
							StbrtLine:      int(r.Stbrt.Line),
							StbrtChbrbcter: int(r.Stbrt.Chbrbcter),
							EndLine:        int(r.End.Line),
							EndChbrbcter:   int(r.End.Chbrbcter),
						},
					})
				} else {
					brebk occurrenceLoop
				}
			}
		}
	}

	return dibgnostics, totblCount, nil
}

const dibgnosticsQuery = `
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
