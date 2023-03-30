package lsifstore

import (
	"context"
	"strconv"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetDefinitionLocations returns the set of locations defining the symbol at the given position.
func (s *store) GetDefinitionLocations(ctx context.Context, bundleID int, path string, line, character, limit, offset int) (_ []shared.Location, _ int, err error) {
	return s.getLocations(ctx, "definition_ranges", extractDefinitionRanges, s.operations.getDefinitions, bundleID, path, line, character, limit, offset)
}

// GetReferenceLocations returns the set of locations referencing the symbol at the given position.
func (s *store) GetReferenceLocations(ctx context.Context, bundleID int, path string, line, character, limit, offset int) (_ []shared.Location, _ int, err error) {
	return s.getLocations(ctx, "reference_ranges", extractReferenceRanges, s.operations.getReferences, bundleID, path, line, character, limit, offset)
}

// GetImplementationLocations returns the set of locations implementing the symbol at the given position.
func (s *store) GetImplementationLocations(ctx context.Context, bundleID int, path string, line, character, limit, offset int) (_ []shared.Location, _ int, err error) {
	return s.getLocations(ctx, "implementation_ranges", extractImplementationRanges, s.operations.getImplementations, bundleID, path, line, character, limit, offset)
}

func (s *store) getLocations(
	ctx context.Context,
	scipFieldName string,
	scipExtractor func(*scip.Document, *scip.Occurrence) []*scip.Range,
	operation *observation.Operation,
	bundleID int,
	path string,
	line, character, limit, offset int,
) (_ []shared.Location, _ int, err error) {
	ctx, trace, endObservation := operation.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.Int("line", line),
		log.Int("character", character),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		locationsDocumentQuery,
		bundleID,
		path,
	)))
	if err != nil || !exists {
		return nil, 0, err
	}

	trace.AddEvent("SCIPData", attribute.Int("numOccurrences", len(documentData.SCIPData.Occurrences)))
	occurrences := types.FindOccurrences(documentData.SCIPData.Occurrences, int32(line), int32(character))
	trace.AddEvent("FindOccurences", attribute.Int("numIntersectingOccurrences", len(occurrences)))

	for _, occurrence := range occurrences {
		var locations []shared.Location
		if ranges := scipExtractor(documentData.SCIPData, occurrence); len(ranges) != 0 {
			locations = append(locations, convertSCIPRangesToLocations(ranges, bundleID, path)...)
		}

		if occurrence.Symbol != "" && !scip.IsLocalSymbol(occurrence.Symbol) {
			monikerLocations, err := s.scanQualifiedMonikerLocations(s.db.Query(ctx, sqlf.Sprintf(
				locationsSymbolSearchQuery,
				pq.Array([]string{occurrence.Symbol}),
				pq.Array([]int{bundleID}),
				sqlf.Sprintf(scipFieldName),
				bundleID,
				path,
				sqlf.Sprintf(scipFieldName),
			)))
			if err != nil {
				return nil, 0, err
			}
			for _, monikerLocation := range monikerLocations {
				for _, row := range monikerLocation.Locations {
					locations = append(locations, shared.Location{
						DumpID: monikerLocation.DumpID,
						Path:   row.URI,
						Range:  newRange(row.StartLine, row.StartCharacter, row.EndLine, row.EndCharacter),
					})
				}
			}
		}

		if len(locations) > 0 {
			totalCount := len(locations)

			if offset < len(locations) {
				locations = locations[offset:]
			} else {
				locations = []shared.Location{}
			}

			if len(locations) > limit {
				locations = locations[:limit]
			}

			return locations, totalCount, nil
		}
	}

	return nil, 0, nil
}

const locationsDocumentQuery = `
SELECT
	sd.id,
	sid.document_path,
	sd.raw_scip_payload
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE
	sid.upload_id = %s AND
	sid.document_path = %s
LIMIT 1
`

const locationsSymbolSearchQuery = `
WITH RECURSIVE
` + symbolIDsCTEs + `
SELECT
	ss.upload_id,
	'' AS scheme,
	'' AS identifier,
	ss.%s,
	sid.document_path
FROM codeintel_scip_symbols ss
JOIN codeintel_scip_document_lookup sid ON sid.id = ss.document_lookup_id
JOIN matching_symbol_names msn ON msn.id = ss.symbol_id
WHERE
	ss.upload_id = %s AND
	sid.document_path != %s AND
	ss.%s IS NOT NULL
`

func newRange(startLine, startCharacter, endLine, endCharacter int) types.Range {
	return types.Range{
		Start: types.Position{
			Line:      startLine,
			Character: startCharacter,
		},
		End: types.Position{
			Line:      endLine,
			Character: endCharacter,
		},
	}
}

func intsToString(vs []int) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.Itoa(v))
	}

	return strings.Join(strs, ", ")
}

type extractedOccurrenceData struct {
	definitions     []*scip.Range
	references      []*scip.Range
	implementations []*scip.Range
	hoverText       []string
}

func extractDefinitionRanges(document *scip.Document, occurrence *scip.Occurrence) []*scip.Range {
	return extractOccurrenceData(document, occurrence).definitions
}

func extractReferenceRanges(document *scip.Document, occurrence *scip.Occurrence) []*scip.Range {
	return append(extractOccurrenceData(document, occurrence).definitions, extractOccurrenceData(document, occurrence).references...)
}

func extractImplementationRanges(document *scip.Document, occurrence *scip.Occurrence) []*scip.Range {
	return extractOccurrenceData(document, occurrence).implementations
}

func extractHoverData(document *scip.Document, occurrence *scip.Occurrence) []string {
	return extractOccurrenceData(document, occurrence).hoverText
}

func extractOccurrenceData(document *scip.Document, occurrence *scip.Occurrence) extractedOccurrenceData {
	if occurrence.Symbol == "" {
		return extractedOccurrenceData{
			hoverText: occurrence.OverrideDocumentation,
		}
	}

	var (
		hoverText               []string
		definitionSymbol        = occurrence.Symbol
		referencesBySymbol      = map[string]struct{}{}
		implementationsBySymbol = map[string]struct{}{}
	)

	// Extract hover text and relationship data from the symbol information that
	// matches the given occurrence. This will give us additional symbol names that
	// we should include in reference and implementation searches.

	if symbol := types.FindSymbol(document, occurrence.Symbol); symbol != nil {
		hoverText = symbol.Documentation

		for _, rel := range symbol.Relationships {
			if rel.IsDefinition {
				definitionSymbol = rel.Symbol
			}
			if rel.IsReference {
				referencesBySymbol[rel.Symbol] = struct{}{}
			}
			if rel.IsImplementation {
				implementationsBySymbol[rel.Symbol] = struct{}{}
			}
		}
	}

	definitions := []*scip.Range{}
	references := []*scip.Range{}
	implementations := []*scip.Range{}

	// Include original symbol names for reference search below
	referencesBySymbol[occurrence.Symbol] = struct{}{}

	// For each occurrence that references one of the definition, reference, or
	// implementation symbol names, extract and aggregate their source positions.

	for _, occ := range document.Occurrences {
		isDefinition := scip.SymbolRole_Definition.Matches(occ)

		// This occurrence defines this symbol
		if definitionSymbol == occ.Symbol && isDefinition {
			definitions = append(definitions, scip.NewRange(occ.Range))
		}

		// This occurrence references this symbol (or a sibling of it)
		if _, ok := referencesBySymbol[occ.Symbol]; ok && !isDefinition {
			references = append(references, scip.NewRange(occ.Range))
		}

		// This occurrence is a definition of a symbol with an implementation relationship
		if _, ok := implementationsBySymbol[occ.Symbol]; ok && isDefinition {
			implementations = append(implementations, scip.NewRange(occ.Range))
		}
	}

	// Override symbol documentation with occurrence documentation, if it exists
	if len(occurrence.OverrideDocumentation) != 0 {
		hoverText = occurrence.OverrideDocumentation
	}

	return extractedOccurrenceData{
		definitions:     definitions,
		references:      references,
		implementations: implementations,
		hoverText:       hoverText,
	}
}
