package scip

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// TargetRangeFetcher returns the set of LSIF range identifiers that form the targets of the given result identifier.
//
// When reading processed LSIF data, this will be determined by checking if the range attached to the input range's
// definition or implementation result set is the same as the input range. When reading unprocessed LSIF data, this
// will be determined by traversing a state map of the read index.
type TargetRangeFetcher func(resultID precise.ID) []precise.ID

// ConvertLSIFDocument converts the given processed LSIF document into a SCIP document.
func ConvertLSIFDocument(
	uploadID int,
	targetRangeFetcher TargetRangeFetcher,
	indexerName string,
	path string,
	document precise.DocumentData,
) *scip.Document {
	var (
		n                         = len(document.Ranges)
		occurrences               = make([]*scip.Occurrence, 0, n)
		documentationBySymbolName = make(map[string]map[string]struct{}, n)
		interfacesBySymbolName    = make(map[string]map[string]struct{}, n)
	)

	// Convert each correlated/canonicalized LSIF range within a document to a set of SCIP occurrences.
	// We may produce more than one occurrence for each range as each occurrence is attached to a single
	// symbol name.
	//
	// As we loop through the LSIF ranges we'll also stash relevant documentation and implementation
	// relationship data that will need to be added to the SCIP document's symbol information slice.

	for id, r := range document.Ranges {
		rangeOccurrences, symbols := convertRange(
			uploadID,
			targetRangeFetcher,
			document,
			id,
			r,
		)

		occurrences = append(occurrences, rangeOccurrences...)

		for _, symbol := range symbols {
			if _, ok := documentationBySymbolName[symbol.name]; !ok {
				documentationBySymbolName[symbol.name] = map[string]struct{}{}
			}

			documentationBySymbolName[symbol.name][symbol.documentation] = struct{}{}

			for _, interfaceName := range symbol.implementationRelationships {
				if _, ok := interfacesBySymbolName[symbol.name]; !ok {
					interfacesBySymbolName[symbol.name] = map[string]struct{}{}
				}

				interfacesBySymbolName[symbol.name][interfaceName] = struct{}{}
			}
		}
	}

	// Convert each LSIF diagnostic within a document to a SCIP occurrence with an attached diagnostic
	for _, diagnostic := range document.Diagnostics {
		occurrences = append(occurrences, convertDiagnostic(diagnostic))
	}

	// Aggregate symbol information to store documentation

	symbolMap := map[string]*scip.SymbolInformation{}
	for symbolName, documentationSet := range documentationBySymbolName {
		var documentation []string
		for doc := range documentationSet {
			if doc != "" {
				documentation = append(documentation, doc)
			}
		}
		sort.Strings(documentation)

		symbolMap[symbolName] = &scip.SymbolInformation{
			Symbol:        symbolName,
			Documentation: documentation,
		}
	}

	// Add additional implements relationships to symbols
	for symbolName, interfaceNames := range interfacesBySymbolName {
		symbol, ok := symbolMap[symbolName]
		if !ok {
			symbol = &scip.SymbolInformation{Symbol: symbolName}
			symbolMap[symbolName] = symbol
		}

		for interfaceName := range interfaceNames {
			symbol.Relationships = append(symbol.Relationships, &scip.Relationship{
				Symbol:           interfaceName,
				IsImplementation: true,
			})

			if _, ok := symbolMap[interfaceName]; !ok {
				symbolMap[interfaceName] = &scip.SymbolInformation{Symbol: interfaceName}
			}
		}
	}

	symbols := make([]*scip.SymbolInformation, 0, len(symbolMap))
	for _, symbol := range symbolMap {
		symbols = append(symbols, symbol)
	}

	return &scip.Document{
		Language:     extractLanguageFromIndexerName(indexerName),
		RelativePath: path,
		Occurrences:  occurrences,
		Symbols:      symbols,
	}
}

type symbolMetadata struct {
	name                        string
	documentation               string
	implementationRelationships []string
}

const maxDefinitionsPerDefinitionResult = 16

// convertRange converts an LSIF range into an equivalent set of SCIP occurrences. The output of this function
// is a slice of occurrences, as multiple moniker names/relationships translate to distinct occurrence objects,
// as well as a slice of additional symbol metadata that should be aggregated and persisted into the enclosing
// document.
func convertRange(
	uploadID int,
	targetRangeFetcher TargetRangeFetcher,
	document precise.DocumentData,
	rangeID precise.ID,
	r precise.RangeData,
) (occurrences []*scip.Occurrence, symbols []symbolMetadata) {
	var monikers []string
	var implementsMonikers []string

	for _, monikerID := range r.MonikerIDs {
		moniker, ok := document.Monikers[monikerID]
		if !ok {
			continue
		}
		packageInformation, ok := document.PackageInformation[moniker.PackageInformationID]
		if !ok {
			continue
		}

		manager := packageInformation.Manager
		if manager == "" {
			manager = "."
		}

		// Construct symbol name so that we still align with the data in lsif_packages and lsif_references
		// tables (in particular, scheme, manager, name, and version must match). We use the entire moniker
		// identifier (as-is) as the sole descriptor in the equivalent SCIP symbol.

		symbolName := fmt.Sprintf(
			"%s %s %s %s `%s`.",
			moniker.Scheme,
			manager,
			packageInformation.Name,
			packageInformation.Version,
			strings.ReplaceAll(moniker.Identifier, "`", "``"),
		)

		switch moniker.Kind {
		case "import":
			fallthrough
		case "export":
			monikers = append(monikers, symbolName)
		case "implementation":
			implementsMonikers = append(implementsMonikers, symbolName)
		}
	}

	for _, targetRangeID := range targetRangeFetcher(r.ImplementationResultID) {
		implementsMonikers = append(implementsMonikers, constructSymbolName(uploadID, targetRangeID))
	}

	addOccurrence := func(symbolName string, symbolRole scip.SymbolRole) {
		occurrences = append(occurrences, &scip.Occurrence{
			Range: []int32{
				int32(r.StartLine),
				int32(r.StartCharacter),
				int32(r.EndLine),
				int32(r.EndCharacter),
			},
			Symbol:      symbolName,
			SymbolRoles: int32(symbolRole),
		})

		symbols = append(symbols, symbolMetadata{
			name:                        symbolName,
			documentation:               document.HoverResults[r.HoverResultID],
			implementationRelationships: implementsMonikers,
		})
	}

	isDefinition := false
	for _, targetRangeID := range targetRangeFetcher(r.DefinitionResultID) {
		if rangeID == targetRangeID {
			isDefinition = true
			break
		}
	}
	if isDefinition {
		role := scip.SymbolRole_Definition

		// Add definition of the range itself
		addOccurrence(constructSymbolName(uploadID, rangeID), role)

		// Add definition of each moniker
		for _, moniker := range monikers {
			addOccurrence(moniker, role)
		}
	} else {
		role := scip.SymbolRole_UnspecifiedSymbolRole

		targetRanges := targetRangeFetcher(r.DefinitionResultID)
		sort.Slice(targetRanges, func(i, j int) bool { return targetRanges[i] < targetRanges[j] })
		if len(targetRanges) > maxDefinitionsPerDefinitionResult {
			targetRanges = targetRanges[:maxDefinitionsPerDefinitionResult]
		}

		for _, targetRangeID := range targetRanges {
			// Add reference to the defining range identifier
			addOccurrence(constructSymbolName(uploadID, targetRangeID), role)
		}

		// Add reference to each moniker
		for _, moniker := range monikers {
			addOccurrence(moniker, role)
		}
	}

	return occurrences, symbols
}

// convertDiagnostic converts an LSIF diagnostic into an equivalent SCIP diagnostic.
func convertDiagnostic(diagnostic precise.DiagnosticData) *scip.Occurrence {
	return &scip.Occurrence{
		Range: []int32{
			int32(diagnostic.StartLine),
			int32(diagnostic.StartCharacter),
			int32(diagnostic.EndLine),
			int32(diagnostic.EndCharacter),
		},
		Diagnostics: []*scip.Diagnostic{
			{
				Severity: scip.Severity(diagnostic.Severity),
				Code:     diagnostic.Code,
				Message:  diagnostic.Message,
				Source:   diagnostic.Source,
				Tags:     nil,
			},
		},
	}
}

// constructSymbolName returns a synthetic SCIP symbol name from the given LSIF identifiers. This is meant
// to be a way to retain behavior of existing indexes, but not necessarily take advantage of things like
// canonical symbol names or non-position-centric queries. For that we rely on the code being re-indexed
// and re-processed as SCIP in the future.
func constructSymbolName(uploadID int, resultID precise.ID) string {
	if resultID == "" {
		return ""
	}

	// scheme = lsif
	// package manager = <empty>
	// package name = upload identifier
	// package version = <empty>
	// descriptor = result identifier (unique within upload)

	return fmt.Sprintf("lsif . %d . `%s`.", uploadID, resultID)
}

// extractLanguageFromIndexerName attempts to extract the SCIP language name from the name of the LSIF
// indexer. If the language name is not recognized an empty string is returned. The returned language
// name will be formatted as defined in the SCIP repository.
func extractLanguageFromIndexerName(indexerName string) string {
	for _, prefix := range []string{"scip-", "lsif-"} {
		if !strings.HasPrefix(indexerName, prefix) {
			continue
		}

		needle := strings.ToLower(strings.TrimPrefix(indexerName, prefix))

		for candidate := range scip.Language_value {
			if needle == strings.ToLower(candidate) {
				return candidate
			}
		}
	}

	return ""
}
