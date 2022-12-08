package lsifstore

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// GetHover returns the hover text of the symbol at the given position.
func (s *store) GetHover(ctx context.Context, bundleID int, path string, line, character int) (_ string, _ types.Range, _ bool, err error) {
	ctx, trace, endObservation := s.operations.getHover.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		hoverDocumentQuery,
		bundleID,
		path,
		bundleID,
		path,
	)))
	if err != nil || !exists {
		return "", types.Range{}, false, err
	}

	if documentData.SCIPData != nil {
		trace.Log(log.Int("numOccurrences", len(documentData.SCIPData.Occurrences)))
		occurrences := types.FindOccurrences(documentData.SCIPData.Occurrences, int32(line), int32(character))
		trace.Log(log.Int("numIntersectingOccurrences", len(occurrences)))

		for _, occurrence := range occurrences {
			// Return the hover data we can extract from the most specific occurrence
			if hoverText := extractHoverData(documentData.SCIPData, occurrence); len(hoverText) != 0 {
				return strings.Join(hoverText, "\n"), translateRange(scip.NewRange(occurrence.Range)), true, nil
			}
		}

		// We don't have any in-document symbol information with hover data, so we'll now attempt to
		// find the symbol information in the text document that defines a symbol attached to the target
		// occurrence.

		// First, we extract the symbol names and the range of the most specific occurrence associated
		// with it. We construct a map and a slice in parallel as we want to retain the ordering of
		// symbols when processing the documents below.

		symbolNames := make([]string, 0, len(occurrences))
		rangeBySymbol := make(map[string]types.Range, len(occurrences))

		for _, occurrence := range occurrences {
			if occurrence.Symbol == "" || scip.IsLocalSymbol(occurrence.Symbol) {
				continue
			}

			if _, ok := rangeBySymbol[occurrence.Symbol]; !ok {
				symbolNames = append(symbolNames, occurrence.Symbol)
				rangeBySymbol[occurrence.Symbol] = translateRange(scip.NewRange(occurrence.Range))
			}
		}

		// Open documents from the same index that define one of the symbols. We return documents ordered
		// by path, which is arbitrary but deterministic in the case that multiple files mark a defining
		// occurrence of a symbol.

		documents, err := s.scanDocumentData(s.db.Query(ctx, sqlf.Sprintf(
			hoverSymbolsQuery,
			bundleID,
			pq.Array(symbolNames),
		)))
		if err != nil {
			return "", types.Range{}, false, err
		}

		// Re-perform the symbol information search. This loop is constructed to prefer matches for symbols
		// associated with the most specific occurrences over less specific occurrences. We also make the
		// observation that processing will inline equivalent symbol information nodes into multiple documents
		// in the persistence layer, so we return the first match rather than aggregating and de-duplicating
		// documentation over all matching documents.

		for _, symbolName := range symbolNames {
			for _, document := range documents {
				for _, symbol := range document.SCIPData.Symbols {
					if symbol.Symbol != symbolName {
						continue
					}

					// Return first match
					return strings.Join(symbol.Documentation, "\n"), rangeBySymbol[symbolName], true, nil
				}
			}
		}

		return "", types.Range{}, false, nil
	}

	trace.Log(log.Int("numRanges", len(documentData.LSIFData.Ranges)))
	ranges := precise.FindRanges(documentData.LSIFData.Ranges, line, character)
	trace.Log(log.Int("numIntersectingRanges", len(ranges)))

	for _, r := range ranges {
		if text, ok := documentData.LSIFData.HoverResults[r.HoverResultID]; ok {
			return text, newRange(r.StartLine, r.StartCharacter, r.EndLine, r.EndCharacter), true, nil
		}
	}

	return "", types.Range{}, false, nil
}

const hoverDocumentQuery = `
(
	SELECT
		sd.id,
		sid.document_path,
		NULL AS data,
		NULL AS ranges,
		NULL AS hovers,
		NULL AS monikers,
		NULL AS packages,
		NULL AS diagnostics,
		sd.raw_scip_payload AS scip_document
	FROM codeintel_scip_document_lookup sid
	JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
	WHERE
		sid.upload_id = %s AND
		sid.document_path = %s
	LIMIT 1
) UNION (
	SELECT
		dump_id,
		path,
		data,
		ranges,
		hovers,
		NULL AS monikers,
		NULL AS packages,
		NULL AS diagnostics,
		NULL AS scip_document
	FROM
		lsif_data_documents
	WHERE
		dump_id = %s AND
		path = %s
	LIMIT 1
)
`

const hoverSymbolsQuery = `
SELECT
	sd.id,
	sid.document_path,
	NULL AS data,
	NULL AS ranges,
	NULL AS hovers,
	NULL AS monikers,
	NULL AS packages,
	NULL AS diagnostics,
	sd.raw_scip_payload AS scip_document
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE EXISTS (
	SELECT 1
	FROM codeintel_scip_symbols ss
	WHERE
		ss.upload_id = %s AND
		ss.symbol_name = ANY(%s) AND
		ss.document_lookup_id = sid.id AND
		ss.definition_ranges IS NOT NULL
)
`
