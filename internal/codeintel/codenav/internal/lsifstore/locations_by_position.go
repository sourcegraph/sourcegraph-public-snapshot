package lsifstore

import (
	"context"
	"fmt"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// GetBulkMonikerLocations returns the locations (within one of the given uploads) with an attached moniker
// whose scheme+identifier matches one of the given monikers. This method also returns the size of the
// complete result set to aid in pagination.
func (s *store) GetBulkMonikerLocations(ctx context.Context, tableName string, uploadIDs []int, monikers []precise.MonikerData, limit, offset int) (_ []shared.Location, totalCount int, err error) {
	ctx, trace, endObservation := s.operations.getBulkMonikerLocations.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("tableName", tableName),
		attribute.Int("numUploadIDs", len(uploadIDs)),
		attribute.IntSlice("uploadIDs", uploadIDs),
		attribute.Int("numMonikers", len(monikers)),
		attribute.String("monikers", monikersToString(monikers)),
		attribute.Int("limit", limit),
		attribute.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	if len(uploadIDs) == 0 || len(monikers) == 0 {
		return nil, 0, nil
	}

	symbolNames := make([]string, 0, len(monikers))
	for _, arg := range monikers {
		symbolNames = append(symbolNames, arg.Identifier)
	}

	query := sqlf.Sprintf(
		bulkMonikerResultsQuery,
		pq.Array(symbolNames),
		pq.Array(uploadIDs),
		sqlf.Sprintf(fmt.Sprintf("%s_ranges", strings.TrimSuffix(tableName, "s"))),
	)

	locationData, err := s.scanQualifiedMonikerLocations(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, err
	}

	totalCount = 0
	for _, monikerLocations := range locationData {
		totalCount += len(monikerLocations.Locations)
	}
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("numUploads", len(locationData)),
		attribute.Int("totalCount", totalCount))

	max := totalCount
	if totalCount > limit {
		max = limit
	}

	locations := make([]shared.Location, 0, max)
outer:
	for _, monikerLocations := range locationData {
		for _, row := range monikerLocations.Locations {
			offset--
			if offset >= 0 {
				continue
			}

			locations = append(locations, shared.Location{
				UploadID: monikerLocations.UploadID,
				Path:     core.NewUploadRelPathUnchecked(row.DocumentPath),
				Range:    shared.NewRange(row.StartLine, row.StartCharacter, row.EndLine, row.EndCharacter),
			})

			if len(locations) >= limit {
				break outer
			}
		}
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numLocations", len(locations)))

	return locations, totalCount, nil
}

const bulkMonikerResultsQuery = `
WITH RECURSIVE
` + symbolIDsCTEs + `
SELECT
	ss.upload_id,
	'scip',
	msn.symbol_name,
	%s,
	document_path
FROM matching_symbol_names msn
JOIN codeintel_scip_symbols ss ON ss.upload_id = msn.upload_id AND ss.symbol_id = msn.id
JOIN codeintel_scip_document_lookup dl ON dl.id = ss.document_lookup_id
ORDER BY ss.upload_id, msn.symbol_name
`

const locationsDocumentQuery = `
SELECT
	sd.id,
	sid.document_path,
	sd.raw_scip_payload
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE
	sid.upload_id = %d AND
	sid.document_path = %s
LIMIT 1
`

type extractedOccurrenceData struct {
	definitions     []shared.UsageBuilder
	references      []shared.UsageBuilder
	implementations []shared.UsageBuilder
	prototypes      []shared.UsageBuilder
	hoverText       []string
}

func extractDefinitionRanges(document *scip.Document, occurrence *scip.Occurrence) []shared.UsageBuilder {
	return extractOccurrenceData(document, occurrence).definitions
}

func extractReferenceRanges(document *scip.Document, occurrence *scip.Occurrence) []shared.UsageBuilder {
	return extractOccurrenceData(document, occurrence).references
}

func extractImplementationRanges(document *scip.Document, occurrence *scip.Occurrence) []shared.UsageBuilder {
	return extractOccurrenceData(document, occurrence).implementations
}

func extractPrototypesRanges(document *scip.Document, occurrence *scip.Occurrence) []shared.UsageBuilder {
	return extractOccurrenceData(document, occurrence).prototypes
}

func extractHoverData(document *scip.Document, occurrence *scip.Occurrence) []string {
	return extractOccurrenceData(document, occurrence).hoverText
}

func extractOccurrenceData(document *scip.Document, lookupOccurrence *scip.Occurrence) extractedOccurrenceData {
	if lookupOccurrence.Symbol == "" {
		return extractedOccurrenceData{
			hoverText: lookupOccurrence.OverrideDocumentation,
		}
	}

	var (
		hoverText               []string
		definitionSymbol        = lookupOccurrence.Symbol
		referencesBySymbol      = collections.NewSet[string]()
		implementationsBySymbol = collections.NewSet[string]()
		prototypeBySymbol       = collections.NewSet[string]()
	)

	// Extract hover text and relationship data from the symbol information that
	// matches the given occurrence. This will give us additional symbol names that
	// we should include in reference and implementation searches.

	if lookupSymbolInfo := scip.FindSymbol(document, lookupOccurrence.Symbol); lookupSymbolInfo != nil {
		hoverText = symbolHoverText(lookupSymbolInfo)

		for _, rel := range lookupSymbolInfo.Relationships {
			if rel.IsDefinition {
				definitionSymbol = rel.Symbol
			}
			if rel.IsReference {
				referencesBySymbol.Add(rel.Symbol)
			}
			if rel.IsImplementation {
				prototypeBySymbol.Add(rel.Symbol)
			}
		}
	}

	for _, sym := range document.Symbols {
		for _, rel := range sym.Relationships {
			if rel.IsImplementation {
				if rel.Symbol == lookupOccurrence.Symbol {
					implementationsBySymbol.Add(sym.Symbol)
				}
			}
		}
	}

	definitions := []shared.UsageBuilder{}
	references := []shared.UsageBuilder{}
	implementations := []shared.UsageBuilder{}
	prototypes := []shared.UsageBuilder{}

	// Include original symbol names for reference search below
	referencesBySymbol.Add(lookupOccurrence.Symbol)

	// For each occurrence that references one of the definition, reference, or
	// implementation symbol names, extract and aggregate their source positions.

	for _, occ := range document.Occurrences {
		isDefinition := scip.SymbolRole_Definition.Matches(occ)

		// This occurrence defines this symbol
		if definitionSymbol == occ.Symbol && isDefinition {
			definitions = append(definitions, shared.NewUsageBuilder(occ))
		}

		// This occurrence references this symbol (or a sibling of it)
		if !isDefinition && referencesBySymbol.Has(occ.Symbol) {
			references = append(references, shared.NewUsageBuilder(occ))
		}

		// This occurrence is a definition of a symbol with an implementation relationship
		if isDefinition && implementationsBySymbol.Has(occ.Symbol) && definitionSymbol != occ.Symbol {
			implementations = append(implementations, shared.NewUsageBuilder(occ))
		}

		// This occurrence is a definition of a symbol with a prototype relationship
		if isDefinition && prototypeBySymbol.Has(occ.Symbol) {
			prototypes = append(prototypes, shared.NewUsageBuilder(occ))
		}
	}

	// Override symbol documentation with occurrence documentation, if it exists
	if len(lookupOccurrence.OverrideDocumentation) != 0 {
		hoverText = lookupOccurrence.OverrideDocumentation
	}

	return extractedOccurrenceData{
		definitions:     definitions,
		references:      references,
		implementations: implementations,
		hoverText:       hoverText,
		prototypes:      prototypes,
	}
}

func monikersToString(vs []precise.MonikerData) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, fmt.Sprintf("%s:%s:%s", v.Kind, v.Scheme, v.Identifier))
	}

	return strings.Join(strs, ", ")
}

func symbolHoverText(symbol *scip.SymbolInformation) []string {
	if sigdoc := symbol.SignatureDocumentation; sigdoc != nil && sigdoc.Text != "" && sigdoc.Language != "" {
		signature := []string{fmt.Sprintf("```%s\n%s\n```", sigdoc.Language, sigdoc.Text)}
		return append(signature, symbol.Documentation...)
	}
	return symbol.Documentation
}

// TODO(id: doc-N-traversals): Internally, these four methods all compute the same
// exact raw data, and then they throw away most of the data. For example, the definition
// extraction logic will waste cycles by getting information about implementations.
//
// Additionally, AFAICT, each function will do a separate read of the document
// from the database and unmarshal it. This means that for the ref panel,
// we will unmarshal the same Protobuf document at least four times. :facepalm:

func (s *store) ExtractDefinitionLocationsFromPosition(ctx context.Context, key FindUsagesKey) (_ []shared.UsageBuilder, _ []string, err error) {
	return s.extractRelatedUsagesAndSymbolNames(ctx, key, s.operations.getDefinitionLocations, extractDefinitionRanges, symbolExtractDefault)
}

func (s *store) ExtractReferenceLocationsFromPosition(ctx context.Context, key FindUsagesKey) (_ []shared.UsageBuilder, _ []string, err error) {
	return s.extractRelatedUsagesAndSymbolNames(ctx, key, s.operations.getReferenceLocations, extractReferenceRanges, symbolExtractDefault)
}

func (s *store) ExtractImplementationLocationsFromPosition(ctx context.Context, key FindUsagesKey) (_ []shared.UsageBuilder, _ []string, err error) {
	return s.extractRelatedUsagesAndSymbolNames(ctx, key, s.operations.getImplementationLocations, extractImplementationRanges, symbolExtractImplementations)
}

func (s *store) ExtractPrototypeLocationsFromPosition(ctx context.Context, key FindUsagesKey) (_ []shared.UsageBuilder, _ []string, err error) {
	return s.extractRelatedUsagesAndSymbolNames(ctx, key, s.operations.getPrototypesLocations, extractPrototypesRanges, symbolExtractPrototype)
}

func symbolExtractDefault(document *scip.Document, symbolName string) (symbols []string) {
	if symbol := scip.FindSymbol(document, symbolName); symbol != nil {
		for _, rel := range symbol.Relationships {
			if rel.IsReference {
				symbols = append(symbols, rel.Symbol)
			}
		}
	}

	return append(symbols, symbolName)
}

func symbolExtractImplementations(document *scip.Document, symbolName string) (symbols []string) {
	for _, sym := range document.Symbols {
		for _, rel := range sym.Relationships {
			if rel.IsImplementation {
				if rel.Symbol == symbolName {
					symbols = append(symbols, sym.Symbol)
				}
			}
		}
	}

	return append(symbols, symbolName)
}

func symbolExtractPrototype(document *scip.Document, symbolName string) (symbols []string) {
	if symbol := scip.FindSymbol(document, symbolName); symbol != nil {
		for _, rel := range symbol.Relationships {
			if rel.IsImplementation {
				symbols = append(symbols, rel.Symbol)
			}
		}
	}

	return symbols
}

//
//

// TODO(id: doc-N-traversals): Since this API is used in a limited number of ways,
// take some basic 'strategy' enums and implement the logic for extraction here
// so we can avoid multiple document traversals.

// extractRelatedUsagesAndSymbolNames uses findUsagesKey to identify a
// position/range/symbol within a single SCIP Document and returns the usages
// and a set of related symbols in that document associated with the findUsagesKey,
// based on the extraction functions.
func (s *store) extractRelatedUsagesAndSymbolNames(
	ctx context.Context,
	findUsagesKey FindUsagesKey,
	operation *observation.Operation,
	extractUsages func(document *scip.Document, occurrence *scip.Occurrence) []shared.UsageBuilder,
	extractRelatedSymbolNames func(document *scip.Document, symbolName string) []string,
) (_ []shared.UsageBuilder, _ []string, err error) {
	ctx, trace, endObservation := operation.With(ctx, &err, observation.Args{Attrs: append([]attribute.KeyValue{
		attribute.Int("uploadID", findUsagesKey.UploadID),
		attribute.String("path", findUsagesKey.Path.RawValue()),
	}, findUsagesKey.Matcher.Attrs()...)})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		locationsDocumentQuery,
		findUsagesKey.UploadID,
		findUsagesKey.Path,
	)))
	if err != nil || !exists {
		return nil, nil, err
	}

	occurrencesMatchingLookupKey, matchKind := findUsagesKey.IdentifyMatchingOccurrences(documentData.SCIPData.Occurrences)

	trace.AddEvent("IdentifyMatchingOccurrences",
		attribute.Int("numDocumentOccurrences", len(documentData.SCIPData.Occurrences)),
		attribute.Int("numMatchingOccurrences", len(occurrencesMatchingLookupKey)),
		attribute.String("matchingKind", string(matchKind)))

	if len(occurrencesMatchingLookupKey) == 0 {
		return nil, nil, nil
	}

	// relatedUsages may contain different kinds of usages depending
	// on the extraction functions
	var relatedUsages []shared.UsageBuilder
	relatedSymbols := collections.NewSet[string]()

	for _, matchingOccurrence := range occurrencesMatchingLookupKey {
		// TODO(id: doc-N-traversals): Optimize this to do a single pass instead of
		// one pass per matching occurrence. Also, we shouldn't need one traversal
		// for an occurrence and one for symbol names, just zero-or-one traversal for the
		// occurrences and zero-or-one traversal over the symbol information.
		//
		// In practice, this loop will only go through 1 iteration in the vast majority
		// of cases, since one source range will generally have a def/ref for a single symbol,
		// so this doesn't need to be fixed urgently.
		relatedUsages = append(relatedUsages,
			extractUsages(documentData.SCIPData, matchingOccurrence)...)

		// QUESTION(id: stronger-doc-canonicalization): Should we strip out occurrences
		// with empty symbol names during canonicalization? Such occurrences will
		// not be targetable by code navigation. This will require a DB migration.
		//
		// NOTE: For local symbols, we know that we will not need to perform any
		// lookups in other documents. So skip the symbol extraction logic instead
		// of having each caller do the skipping in extractRelatedSymbolNames.
		if matchingOccurrence.Symbol != "" && !scip.IsLocalSymbol(matchingOccurrence.Symbol) {
			relatedSymbols.Add(extractRelatedSymbolNames(documentData.SCIPData, matchingOccurrence.Symbol)...)
		}
	}

	switch matchKind {
	case SinglePositionBasedMatching:
		// When using matching using a single position, we may get a set of overlapping
		// occurrences, all for the same source range. In that case, we don't care about
		// the symbol data, so we de-duplicate the objects purely based on source range.
		//
		// So if there are multiple symbols for the same range, then only one will be used.
		relatedUsages = collections.DeduplicateBy(relatedUsages, shared.UsageBuilder.RangeKey)
	case RangeBasedMatching:
		// When using range-based exact matching, we already know that the ranges for
		// all the occurrences must be equal. So we don't need to deduplicate based on
		// that. However, we need to maintain different objects for different symbol
		// names and roles.
		relatedUsages = collections.DeduplicateBy(relatedUsages, shared.UsageBuilder.SymbolAndRoleKey)
		//return collections.DeduplicateBy(relatedOccurrences, uniqueBySymbolNameAndRole), collections.Deduplicate(relatedSymbols), nil
	case RangeAndSymbolBasedMatching:
		relatedUsages = collections.DeduplicateBy(relatedUsages, shared.UsageBuilder.SymbolRoleKey)
	}

	return relatedUsages, collections.SortedSetValues(relatedSymbols), nil
}

//
//

func (s *store) GetMinimalBulkMonikerLocations(ctx context.Context, usageKind shared.UsageKind, uploadIDs []int, skipPaths map[int]string, monikers []precise.MonikerData, limit, offset int) (_ []shared.Usage, totalCount int, err error) {
	ctx, trace, endObservation := s.operations.getBulkMonikerLocations.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("tableName", usageKind.TableName()),
		attribute.Int("numUploadIDs", len(uploadIDs)),
		attribute.IntSlice("uploadIDs", uploadIDs),
		attribute.Int("numMonikers", len(monikers)),
		attribute.String("monikers", monikersToString(monikers)),
		attribute.Int("limit", limit),
		attribute.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	if len(uploadIDs) == 0 || len(monikers) == 0 {
		return nil, 0, nil
	}

	symbolNames := make([]string, 0, len(monikers))
	for _, arg := range monikers {
		symbolNames = append(symbolNames, arg.Identifier)
	}

	var skipConds []*sqlf.Query
	for _, id := range uploadIDs {
		if path, ok := skipPaths[id]; ok {
			skipConds = append(skipConds, sqlf.Sprintf("(%s, %s)", id, path))
		}
	}
	if len(skipConds) == 0 {
		skipConds = append(skipConds, sqlf.Sprintf("(%s, %s)", -1, ""))
	}

	fieldName := fmt.Sprintf("%s_ranges", strings.TrimSuffix(usageKind.TableName(), "s"))
	query := sqlf.Sprintf(
		minimalBulkMonikerResultsQuery,
		pq.Array(symbolNames),
		pq.Array(uploadIDs),
		sqlf.Sprintf(fieldName),
		sqlf.Sprintf(fieldName),
		sqlf.Join(skipConds, ", "),
	)

	locationData, err := s.scanDeduplicatedQualifiedMonikerLocations(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, err
	}

	totalCount = 0
	for _, monikerLocations := range locationData {
		totalCount += len(monikerLocations.Locations)
	}
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("numUploads", len(locationData)),
		attribute.Int("totalCount", totalCount))

	max := totalCount
	if totalCount > limit {
		max = limit
	}

	locations := make([]shared.Usage, 0, max)
outer:
	for _, monikerLocations := range locationData {
		for _, row := range monikerLocations.Locations {
			offset--
			if offset >= 0 {
				continue
			}

			locations = append(locations, shared.Usage{
				UploadID: monikerLocations.UploadID,
				Path:     core.NewUploadRelPathUnchecked(row.DocumentPath),
				Range:    shared.NewRange(row.StartLine, row.StartCharacter, row.EndLine, row.EndCharacter),
				Symbol:   row.Symbol,
				Kind:     usageKind,
			})

			if len(locations) >= limit {
				break outer
			}
		}
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numLocations", len(locations)))

	return locations, totalCount, nil
}

const minimalBulkMonikerResultsQuery = `
WITH RECURSIVE
` + symbolIDsCTEs + `
SELECT
	ss.upload_id,
	%s,
	document_path,
	msn.symbol_name
FROM codeintel_scip_symbols ss
JOIN codeintel_scip_document_lookup dl ON dl.id = ss.document_lookup_id
JOIN matching_symbol_names msn ON msn.upload_id = ss.upload_id AND msn.id = ss.symbol_id
WHERE
	ss.%s IS NOT NULL AND
	(ss.upload_id, dl.document_path) NOT IN (%s)
ORDER BY ss.upload_id, dl.document_path
`
