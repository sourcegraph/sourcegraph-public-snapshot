package lsifstore

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

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

type extractedOccurrenceData struct {
	definitions     []scip.Range
	references      []scip.Range
	implementations []scip.Range
	prototypes      []scip.Range
	hoverText       []string
}

func extractDefinitionRanges(document *scip.Document, lookup *scip.Occurrence) []scip.Range {
	return extractOccurrenceData(document, lookup).definitions
}

func extractReferenceRanges(document *scip.Document, lookup *scip.Occurrence) []scip.Range {
	return extractOccurrenceData(document, lookup).references
}

func extractImplementationRanges(document *scip.Document, lookup *scip.Occurrence) []scip.Range {
	return extractOccurrenceData(document, lookup).implementations
}

func extractPrototypesRanges(document *scip.Document, lookup *scip.Occurrence) []scip.Range {
	return extractOccurrenceData(document, lookup).prototypes
}

func extractHoverData(document *scip.Document, lookup *scip.Occurrence) []string {
	return extractOccurrenceData(document, lookup).hoverText
}

// extractOccurrenceData identifies occurrences inside document that are related to
// lookupOccurrence in various ways (e.g. defs/refs/impls/supers etc.)
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

	definitions := []scip.Range{}
	references := []scip.Range{}
	implementations := []scip.Range{}
	prototypes := []scip.Range{}

	// Include original symbol names for reference search below
	referencesBySymbol.Add(lookupOccurrence.Symbol)

	// For each occurrence that references one of the definition, reference, or
	// implementation symbol names, extract and aggregate their source positions.

	for _, occ := range document.Occurrences {
		isDefinition := scip.SymbolRole_Definition.Matches(occ)

		// This occurrence defines this symbol
		if definitionSymbol == occ.Symbol && isDefinition {
			definitions = append(definitions, scip.NewRangeUnchecked(occ.Range))
		}

		// This occurrence references this symbol (or a sibling of it)
		if !isDefinition && referencesBySymbol.Has(occ.Symbol) {
			references = append(references, scip.NewRangeUnchecked(occ.Range))
		}

		// This occurrence is a definition of a symbol with an implementation relationship
		if isDefinition && implementationsBySymbol.Has(occ.Symbol) && definitionSymbol != occ.Symbol {
			implementations = append(implementations, scip.NewRangeUnchecked(occ.Range))
		}

		// This occurrence is a definition of a symbol with a prototype relationship
		if isDefinition && prototypeBySymbol.Has(occ.Symbol) {
			prototypes = append(prototypes, scip.NewRangeUnchecked(occ.Range))
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

func (s *store) ExtractDefinitionLocationsFromPosition(ctx context.Context, locationKey LocationKey) (_ []shared.Location, _ []string, err error) {
	return s.extractLocationsFromPosition(ctx, extractDefinitionRanges, symbolExtractDefault, s.operations.getDefinitionLocations, locationKey)
}

func (s *store) ExtractReferenceLocationsFromPosition(ctx context.Context, locationKey LocationKey) (_ []shared.Location, _ []string, err error) {
	return s.extractLocationsFromPosition(ctx, extractReferenceRanges, symbolExtractDefault, s.operations.getReferenceLocations, locationKey)
}

func (s *store) ExtractImplementationLocationsFromPosition(ctx context.Context, locationKey LocationKey) (_ []shared.Location, _ []string, err error) {
	return s.extractLocationsFromPosition(ctx, extractImplementationRanges, symbolExtractImplementations, s.operations.getImplementationLocations, locationKey)
}

func (s *store) ExtractPrototypeLocationsFromPosition(ctx context.Context, locationKey LocationKey) (_ []shared.Location, _ []string, err error) {
	return s.extractLocationsFromPosition(ctx, extractPrototypesRanges, symbolExtractPrototype, s.operations.getPrototypesLocations, locationKey)
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

func (s *store) extractLocationsFromPosition(
	ctx context.Context,
	extractLocations func(document *scip.Document, occurrence *scip.Occurrence) []scip.Range,
	extractRelatedSymbolNames func(document *scip.Document, symbolName string) []string,
	operation *observation.Operation,
	key LocationKey,
) (_ []shared.Location, _ []string, err error) {
	matcher := shared.NewStartPositionMatcher(scip.Position{Line: int32(key.Line), Character: int32(key.Character)})
	findUsagesKey := FindUsagesKey{UploadID: key.UploadID, Path: key.Path, Matcher: matcher}
	extractUsages := func(document *scip.Document, occurrence *scip.Occurrence) []shared.UsageBuilder {
		return genslices.Map(extractLocations(document, occurrence), func(range_ scip.Range) shared.UsageBuilder {
			return shared.UsageBuilder{
				Range:       range_,
				Symbol:      "",
				SymbolRoles: scip.SymbolRole_UnspecifiedSymbolRole,
				// We will discard Symbol and SymbolRoles below, so use zero values here for simplicity.
			}
		})
	}
	usageBuilders, relatedSymbols, err := s.extractRelatedUsagesAndSymbolNames(ctx, findUsagesKey, operation, extractUsages, extractRelatedSymbolNames)
	return genslices.Map(usageBuilders, func(ub shared.UsageBuilder) shared.Location {
		return shared.Location{
			UploadID: key.UploadID,
			Path:     key.Path,
			Range:    shared.TranslateRange(ub.Range),
		}
	}), relatedSymbols, err
}

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
	// TODO(id: doc-N-traversals): Since this API is used in a limited number of ways,
	// consider de-functionalizing this to take a 'strategy' enum/bitset
	// and handling extraction of all related symbols in one pass based on that.

	ctx, trace, endObservation := operation.With(ctx, &err, observation.Args{Attrs: append([]attribute.KeyValue{
		attribute.Int("uploadID", findUsagesKey.UploadID),
		attribute.String("path", findUsagesKey.Path.RawValue()),
	}, findUsagesKey.Matcher.Attrs()...)})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		locationsDocumentQuery,
		findUsagesKey.UploadID,
		findUsagesKey.Path.RawValue(),
	)))
	if err != nil || !exists {
		return nil, nil, err
	}

	occurrencesMatchingLookupKey, matchStrategy := findUsagesKey.IdentifyMatchingOccurrences(documentData.SCIPData.Occurrences)

	trace.AddEvent("IdentifyMatchingOccurrences",
		attribute.Int("numDocumentOccurrences", len(documentData.SCIPData.Occurrences)),
		attribute.Int("numMatchingOccurrences", len(occurrencesMatchingLookupKey)),
		attribute.String("matchStrategy", string(matchStrategy)))

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

	switch matchStrategy {
	case SinglePositionBasedMatchStrategy:
		// When using matching using a single position, we may get a set of overlapping
		// occurrences, all for the same source range. In that case, we don't care about
		// the symbol data, so we de-duplicate the objects purely based on source range.
		//
		// So if there are multiple symbols for the same range, then only one will be used.
		relatedUsages = collections.DeduplicateBy(relatedUsages, shared.UsageBuilder.RangeKey)
	case RangeBasedMatchStrategy:
		// When using range-based exact matching, we already know that the ranges for
		// all the occurrences must be equal. So we don't need to deduplicate based on
		// that. However, we need to maintain different objects for different symbol
		// names and roles.
		relatedUsages = collections.DeduplicateBy(relatedUsages, shared.UsageBuilder.SymbolAndRoleKey)
	case RangeAndSymbolBasedMatchStrategy:
		relatedUsages = collections.DeduplicateBy(relatedUsages, shared.UsageBuilder.SymbolRoleKey)
	}

	return relatedUsages, collections.SortedSetValues(relatedSymbols), nil
}

func (s *store) GetSymbolUsages(ctx context.Context, opts SymbolUsagesOptions) (_ []shared.Usage, totalCount int, err error) {
	ctx, trace, endObservation := s.operations.getSymbolUsages.With(ctx, &err, observation.Args{Attrs: opts.Attrs()})
	defer endObservation(1, observation.Args{})

	if len(opts.UploadIDs) == 0 || len(opts.LookupSymbols) == 0 {
		return nil, 0, nil
	}

	var skipConds []*sqlf.Query
	for _, id := range opts.UploadIDs {
		if path, ok := opts.SkipPathsByUploadID[id]; ok {
			skipConds = append(skipConds, sqlf.Sprintf("(%s, %s)", id, path))
		}
	}
	if len(skipConds) == 0 {
		skipConds = append(skipConds, sqlf.Sprintf("(%s, %s)", -1, ""))
	}

	rangesColumn := sqlf.Sprintf(opts.UsageKind.RangesColumnName())
	query := sqlf.Sprintf(
		symbolUsagesQuery,
		pq.Array(opts.LookupSymbols),
		pq.Array(opts.UploadIDs),
		rangesColumn,
		rangesColumn,
		sqlf.Join(skipConds, ", "),
	)

	usageData, err := s.scanUploadSymbolLoci(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, err
	}

	totalCount = 0
	for _, data := range usageData {
		totalCount += len(data.Loci)
	}
	trace.AddEvent("TODO Domain Owner",
		attribute.Int("numUniqueUploadIDSymbolPairs", len(usageData)),
		attribute.Int("totalCount", totalCount))

	usages := make([]shared.Usage, 0, min(totalCount, opts.Limit))
	offset := opts.Offset
outer:
	for _, uploadSymbolLoci := range usageData {
		for _, locus := range uploadSymbolLoci.Loci {
			offset--
			if offset >= 0 {
				continue
			}

			usages = append(usages, shared.Usage{
				UploadID: uploadSymbolLoci.UploadID,
				Path:     locus.Path,
				Range:    shared.TranslateRange(locus.Range),
				Symbol:   uploadSymbolLoci.Symbol,
				Kind:     opts.UsageKind,
			})

			if len(usages) >= opts.Limit {
				break outer
			}
		}
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numUsages", len(usages)))

	return usages, totalCount, nil
}

// symbolUsagesQuery gets ALL usages of a bunch of symbols across the ENTIRE instance
// (within the given set of uploadIDs). We need to do this because the ranges are
// stored using a custom binary encoding which means we can't use LIMIT+OFFSET at
// the level of locations.
const symbolUsagesQuery = `
WITH RECURSIVE
` + symbolIDsCTEs + `
SELECT
	ss.upload_id,
	msn.symbol_name,
	array_agg(%s ORDER BY dl.document_path),
	array_agg(document_path ORDER BY dl.document_path)
    -- ORDER BY ss.upload_id, msn.symbol_name, dl.document_path to maintain determinism for pagination
FROM codeintel_scip_symbols ss
JOIN codeintel_scip_document_lookup dl
     ON dl.id = ss.document_lookup_id
JOIN matching_symbol_names msn
     ON msn.upload_id = ss.upload_id AND msn.id = ss.symbol_id
WHERE
	ss.%s IS NOT NULL AND
	(ss.upload_id, dl.document_path) NOT IN (%s)
GROUP BY ss.upload_id, msn.symbol_name
ORDER BY ss.upload_id, msn.symbol_name
`
