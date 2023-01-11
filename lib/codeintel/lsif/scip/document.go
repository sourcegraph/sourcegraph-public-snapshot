package scip

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// DefinitionMatcher returns true if the given definition result ID has the given target range ID within
// the given document path in the list of its definition ranges.
//
// When reading processed LSIF data, this will be determined by checking if the range attached to the input
// range's definition result set is the same as the input range. When reading unprocessed LSIF data, this
// will be determined by traversing a state map of the read index.
type DefinitionMatcher func(targetPath string, targetRangeID precise.ID, definitionResultID precise.ID) bool

// ConvertLSIFDocument converts the given processed LSIF document into a SCIP document.
func ConvertLSIFDocument(
	uploadID int,
	definitionMatcher DefinitionMatcher,
	indexerName string,
	path string,
	document precise.DocumentData,
) *scip.Document {
	var (
		n                           = len(document.Ranges)
		occurrences                 = make([]*scip.Occurrence, 0, n)
		documentationBySymbolName   = make(map[string]map[string]struct{}, n)
		implementationsBySymbolName = make(map[string]map[string]struct{}, n)
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
			definitionMatcher,
			path,
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

			for _, other := range symbol.implementationRelationships {
				if _, ok := implementationsBySymbolName[symbol.name]; !ok {
					implementationsBySymbolName[symbol.name] = map[string]struct{}{}
				}

				implementationsBySymbolName[symbol.name][other] = struct{}{}
			}
		}
	}

	// Convert each LSIF diagnostic within a document to a SCIP occurrence with an attached diagnostic
	for _, diagnostic := range document.Diagnostics {
		occurrences = append(occurrences, convertDiagnostic(diagnostic))
	}

	// Aggregate symbol information to store documentation and implementation relationships
	symbols := make([]*scip.SymbolInformation, 0, len(documentationBySymbolName))
	for symbolName, documentationSet := range documentationBySymbolName {
		var documentation []string
		for doc := range documentationSet {
			if doc != "" {
				documentation = append(documentation, doc)
			}
		}
		sort.Strings(documentation)

		symbols = append(symbols, &scip.SymbolInformation{
			Symbol:        symbolName,
			Documentation: documentation,
		})
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

// convertRange converts an LSIF range into an equivalent set of SCIP occurrences. The output of this function
// is a slice of occurrences, as multiple moniker names/relationships translate to distinct occurrence objects,
// as well as a slice of additional symbol metadata that should be aggregated and persisted into the enclosing
// document.
func convertRange(
	uploadID int,
	definitionMatcher DefinitionMatcher,
	path string,
	document precise.DocumentData,
	rangeID precise.ID,
	r precise.RangeData,
) ([]*scip.Occurrence, []symbolMetadata) {
	symbolNames, implementsSymbolNames := constructSymbolNames(uploadID, document, r)

	var (
		n           = len(symbolNames)
		occurrences = make([]*scip.Occurrence, 0, n)
		symbols     = make([]symbolMetadata, 0, n)
	)

	symbolRoles := scip.SymbolRole_UnspecifiedSymbolRole
	if r.DefinitionResultID != "" && definitionMatcher(path, rangeID, r.DefinitionResultID) {
		symbolRoles = symbolRoles | scip.SymbolRole_Definition
	}

	for _, symbolName := range symbolNames {
		occurrences = append(occurrences, &scip.Occurrence{
			Range: []int32{
				int32(r.StartLine),
				int32(r.StartCharacter),
				int32(r.EndLine),
				int32(r.EndCharacter),
			},
			Symbol:      symbolName,
			SymbolRoles: int32(symbolRoles),
		})

		symbols = append(symbols, symbolMetadata{
			name:                        symbolName,
			documentation:               document.HoverResults[r.HoverResultID],
			implementationRelationships: implementsSymbolNames,
		})
	}

	return occurrences, symbols
}

// constructSymbolNames returns a slice of symbol names to be associated with the given range, as well as
// a slice of symbol names that are related to this range via an implements relationship.
//
// Symbol names associated with a range include a synthetic symbol name constructed from the upload ID and
// the definition/reference/implementation result identifiers. These are not deterministic across uploads
// (as LSIF graph object interning rewrote them to arbitrary values) and won't allow us to take immediate
// benefit of SCIP data sharing. Symbol names will also represent moniker/package information pairs attached
// to the range.
func constructSymbolNames(uploadID int, document precise.DocumentData, r precise.RangeData) (symbolNames, implementsSymbolNames []string) {
	var (
		symbolNameMap           = make(map[string]struct{}, 4)
		implementsSymbolNameMap = make(map[string]struct{}, 4)
	)

	symbolNameMap[constructSymbolName(uploadID, r.DefinitionResultID)] = struct{}{}
	symbolNameMap[constructSymbolName(uploadID, r.ReferenceResultID)] = struct{}{}
	implementsSymbolNameMap[constructSymbolName(uploadID, r.ImplementationResultID)] = struct{}{}

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
			symbolNameMap[symbolName] = struct{}{}
		case "export":
			symbolNameMap[symbolName] = struct{}{}
		case "implementation":
			implementsSymbolNameMap[symbolName] = struct{}{}
		}
	}

	return flattenNonEmptyMap(symbolNameMap), flattenNonEmptyMap(implementsSymbolNameMap)
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

// flattenNonEmptyMap returns an ordered slice of the keys from the given map, excluding the empty string.
func flattenNonEmptyMap(m map[string]struct{}) []string {
	s := make([]string, 0, len(m))
	for k := range m {
		if k != "" {
			s = append(s, k)
		}
	}
	sort.Strings(s)

	return s
}
