package reader

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/writer"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ConvertTypedIndexToGraphIndex takes an LSIF Typed index and returns the equivalent LSIF Graph index.
// There doesn't exist a reliable bijection between LSIF Typed and LSIF Typed.
// This conversion is lossy because LSIF Typed includes metadata that has no equivalent encoding in
// LSIF Graph, such as lsiftyped.SymbolRole beyond the definition role.
// Also, LSIF Graph allows encoding certain behaviors that LSIF Typed current doesn't support,
// such as asymmetric references/definitions.
func ConvertTypedIndexToGraphIndex(index *lsiftyped.Index) ([]Element, error) {
	g := newGraph()

	if index.Metadata == nil {
		return nil, errors.New(".Metadata is nil")
	}
	if index.Metadata.ToolInfo == nil {
		return nil, errors.New(".Metadata.ToolInfo is nil")
	}

	positionEncoding := ""
	switch index.Metadata.TextDocumentEncoding {
	case lsiftyped.TextEncoding_UTF8:
		positionEncoding = "utf-8"
	case lsiftyped.TextEncoding_UTF16:
		positionEncoding = "utf-16"
	default:
		return nil, errors.New(".Metadata.TextDocumentEncoding does not have value utf-8 or utf-16")
	}

	g.emitVertex(
		"metaData",
		MetaData{
			Version:          "0.4.3", // Hardcoded LSIF Graph version.
			ProjectRoot:      index.Metadata.ProjectRoot,
			PositionEncoding: positionEncoding,
			ToolInfo: ToolInfo{
				Name:    index.Metadata.ToolInfo.Name,
				Version: index.Metadata.ToolInfo.Version,
			},
		},
	)

	// Pass 1: create result sets for global symbols.
	for _, importedSymbol := range index.ExternalSymbols {
		g.symbolToResultSet[importedSymbol.Symbol] = g.emitResultSet(importedSymbol, "import")
	}
	for _, document := range index.Documents {
		for _, exportedSymbol := range document.Symbols {
			g.registerInverseRelationships(exportedSymbol)
			if lsiftyped.IsGlobalSymbol(exportedSymbol.Symbol) {
				// Local symbols are skipped here because we handle them in the
				// second pass when processing individual documents.
				g.symbolToResultSet[exportedSymbol.Symbol] = g.emitResultSet(exportedSymbol, "export")
			}
		}
	}

	// Pass 2: emit ranges for all documents.
	for _, document := range index.Documents {
		g.emitDocument(index, document)
	}

	return g.Elements, nil
}

// graph is a helper struct to emit an LSIF Graph.
type graph struct {
	ID                   int
	Elements             []Element
	symbolToResultSet    map[string]symbolInformationIDs
	inverseRelationships map[string][]*lsiftyped.Relationship
	packageToGraphID     map[string]int
}

// symbolInformationIDs is a container for LSIF Graph IDs corresponding to an lsiftyped.SymbolInformation.
type symbolInformationIDs struct {
	ResultSet            int
	DefinitionResult     int
	ReferenceResult      int
	ImplementationResult int
	HoverResult          int
}

func newGraph() graph {
	return graph{
		ID:                   0,
		Elements:             []Element{},
		symbolToResultSet:    map[string]symbolInformationIDs{},
		packageToGraphID:     map[string]int{},
		inverseRelationships: map[string][]*lsiftyped.Relationship{},
	}
}

func (g *graph) emitPackage(pkg *lsiftyped.Package) int {
	id := pkg.ID()
	graphID, ok := g.packageToGraphID[id]
	if ok {
		return graphID
	}

	graphID = g.emitVertex("packageInformation", PackageInformation{
		Name:    pkg.Name,
		Version: pkg.Version,
		Manager: pkg.Manager,
	})
	g.packageToGraphID[pkg.ID()] = graphID
	return graphID
}

// emitResultSet emits the associated resultSet, definitionResult, referenceResult, implementationResult and hoverResult
// for the provided lsiftyped.SymbolInformation.
func (g *graph) emitResultSet(info *lsiftyped.SymbolInformation, monikerKind string) symbolInformationIDs {
	if ids, ok := g.symbolToResultSet[info.Symbol]; ok {
		return ids
	}
	// NOTE: merge separate documentation sections with a horizontal Markdown rule. Indexers that emit LSIF graph
	// directly need to emit this separator directly while with LSIF Typed we render the horizontal rule here.
	hover := strings.Join(info.Documentation, "\n\n---\n\n")
	definitionResult := -1
	hasDefinition := monikerKind == "export" || monikerKind == "local"
	if hasDefinition {
		definitionResult = g.emitVertex("definitionResult", nil)
	}
	ids := symbolInformationIDs{
		ResultSet:            g.emitVertex("resultSet", ResultSet{}),
		DefinitionResult:     definitionResult,
		ReferenceResult:      g.emitVertex("referenceResult", nil),
		ImplementationResult: -1,
		HoverResult:          g.emitVertex("hoverResult", hover),
	}
	if hasDefinition {
		g.emitEdge("textDocument/definition", Edge{OutV: ids.ResultSet, InV: ids.DefinitionResult})
	}
	g.emitEdge("textDocument/references", Edge{OutV: ids.ResultSet, InV: ids.ReferenceResult})
	g.emitEdge("textDocument/hover", Edge{OutV: ids.ResultSet, InV: ids.HoverResult})
	if monikerKind == "export" || monikerKind == "import" {
		g.emitMonikerVertex(info.Symbol, monikerKind, ids.ResultSet)
	}
	return ids
}

// emitDocument emits all range vertices for the `lsiftyped.Occurrence` in the provided document, along with
// associated `item` edges to link ranges with result sets.
func (g *graph) emitDocument(index *lsiftyped.Index, doc *lsiftyped.Document) {
	uri := filepath.Join(index.Metadata.ProjectRoot, doc.RelativePath)
	documentID := g.emitVertex("document", uri)

	documentSymbolTable := map[string]*lsiftyped.SymbolInformation{}
	localSymbolInformationTable := map[string]symbolInformationIDs{}
	for _, info := range doc.Symbols {
		documentSymbolTable[info.Symbol] = info

		// Build symbol information table for Document-local symbols only.
		if lsiftyped.IsLocalSymbol(info.Symbol) {
			localSymbolInformationTable[info.Symbol] = g.emitResultSet(info, "local")
		}

		// Emit "implementation" monikers for external symbols (monikers with kind "import")
		for _, relationship := range info.Relationships {
			if relationship.IsImplementation {
				relationshipIDs := g.getOrInsertSymbolInformationIDs(relationship.Symbol, localSymbolInformationTable)
				if relationshipIDs.DefinitionResult > 0 {
					// Not an imported symbol
					continue
				}
				infoIDs := g.getOrInsertSymbolInformationIDs(info.Symbol, localSymbolInformationTable)
				g.emitMonikerVertex(relationship.Symbol, "implementation", infoIDs.ResultSet)
			}
		}
	}

	var rangeIDs []int
	for _, occ := range doc.Occurrences {
		rangeID, err := g.emitRange(occ.Range)
		if err != nil {
			// Silently skip invalid ranges.
			// TODO: add option to print a warning or fail fast here https://github.com/sourcegraph/sourcegraph/issues/31415
			continue
		}
		rangeIDs = append(rangeIDs, rangeID)
		resultIDs := g.getOrInsertSymbolInformationIDs(occ.Symbol, localSymbolInformationTable)
		g.emitEdge("next", Edge{OutV: rangeID, InV: resultIDs.ResultSet})
		isDefinition := occ.SymbolRoles&int32(lsiftyped.SymbolRole_Definition) != 0
		if isDefinition && resultIDs.DefinitionResult > 0 {
			g.emitEdge("item", Edge{OutV: resultIDs.DefinitionResult, InVs: []int{rangeID}, Document: documentID})
			symbolInfo, ok := documentSymbolTable[occ.Symbol]
			if ok {
				g.emitRelationships(rangeID, documentID, resultIDs, localSymbolInformationTable, symbolInfo)
			}
		}
		// reference
		g.emitEdge("item", Edge{OutV: resultIDs.ReferenceResult, InVs: []int{rangeID}, Document: documentID})
	}
	g.emitEdge("contains", Edge{OutV: documentID, InVs: rangeIDs})
}

// emitRelationships emits "referenceResults" and "implementationResult" based on the value of lsiftyped.SymbolInformation.Relationships
func (g *graph) emitRelationships(rangeID, documentID int, resultIDs symbolInformationIDs, localResultIDs map[string]symbolInformationIDs, info *lsiftyped.SymbolInformation) {
	var allReferenceResultIds []int
	relationships := g.inverseRelationships[info.Symbol]
	for _, relationship := range relationships {
		allReferenceResultIds = append(allReferenceResultIds, g.emitRelationship(relationship, rangeID, documentID, localResultIDs)...)
	}
	for _, relationship := range info.Relationships {
		allReferenceResultIds = append(allReferenceResultIds, g.emitRelationship(relationship, rangeID, documentID, localResultIDs)...)
	}
	if len(allReferenceResultIds) > 0 {
		g.emitEdge("item", Edge{
			OutV:     resultIDs.ReferenceResult,
			InVs:     allReferenceResultIds,
			Document: documentID,
			// According to the LSIF Graph spec, the 'property' field is required but it's not present in the reader.Element struct.
			// Property: "referenceResults",
		})
	}
}

func (g *graph) emitRelationship(relationship *lsiftyped.Relationship, rangeID, documentID int, localResultIDs map[string]symbolInformationIDs) []int {
	relationshipIDs := g.getOrInsertSymbolInformationIDs(relationship.Symbol, localResultIDs)

	if relationship.IsImplementation {
		if relationshipIDs.ImplementationResult < 0 {
			relationshipIDs.ImplementationResult = g.emitVertex("implementationResult", nil)
			g.emitEdge("textDocument/implementation", Edge{OutV: relationshipIDs.ResultSet, InV: relationshipIDs.ImplementationResult})
		}
		g.emitEdge("item", Edge{OutV: relationshipIDs.ImplementationResult, InVs: []int{rangeID}, Document: documentID})
	}

	if relationship.IsReference {
		g.emitEdge("item", Edge{
			OutV:     relationshipIDs.ReferenceResult,
			InVs:     []int{rangeID},
			Document: documentID,
			// The 'property' field is included in the LSIF Graph JSON but it's not present in reader.Element
			// Property: "referenceResults",
		})
		return []int{relationshipIDs.ReferenceResult}
	}

	return nil
}

// emitMonikerVertex emits the "moniker" vertex and optionally the accompanying "packageInformation" vertex.
func (g *graph) emitMonikerVertex(symbolID string, kind string, resultSetID int) {
	symbol, err := lsiftyped.ParsePartialSymbol(symbolID, false)
	if err != nil || symbol == nil || symbol.Scheme == "" {
		// Silently ignore symbols that are missing the scheme. The entire symbol does not have to be formatted
		// according to the BNF grammar in lsiftyped.Symbol, we only reject symbols that are missing the scheme.
		// TODO: add option to print a warning or fail fast here https://github.com/sourcegraph/sourcegraph/issues/31415
		return
	}
	// Accept the symbol as long as it has a non-empty scheme. We ignore
	// parse errors because we can still provide accurate
	// definition/references/hover within a repo.
	scheme := symbol.Scheme
	if symbol.Package != nil {
		// NOTE: these special cases are needed since the Sourcegraph backend uses the "scheme" field of monikers where
		// it should use the "manager" field of packageInformation instead.
		switch symbol.Scheme {
		case "lsif-java", "scip-java":
			scheme = "semanticdb"
		case "lsif-typescript":
			scheme = "npm"
		}
	}
	monikerID := g.emitVertex("moniker", Moniker{
		Kind:       kind,
		Scheme:     scheme,
		Identifier: symbolID,
	})
	g.emitEdge("moniker", Edge{OutV: resultSetID, InV: monikerID})
	if symbol.Package != nil &&
		symbol.Package.Manager != "" &&
		symbol.Package.Name != "" &&
		symbol.Package.Version != "" {
		packageID := g.emitPackage(symbol.Package)
		g.emitEdge("packageInformation", Edge{OutV: monikerID, InV: packageID})
	}
}

func (g *graph) emitRange(lsifRange []int32) (int, error) {
	startLine, startCharacter, endLine, endCharacter, err := interpretLsifRange(lsifRange)
	if err != nil {
		return 0, err
	}
	return g.emit("vertex", "range", Range{
		RangeData: protocol.RangeData{
			Start: protocol.Pos{
				Line:      int(startLine),
				Character: int(startCharacter),
			},
			End: protocol.Pos{
				Line:      int(endLine),
				Character: int(endCharacter),
			},
		},
	}), nil
}

func (g *graph) emitVertex(label string, payload any) int {
	return g.emit("vertex", label, payload)
}

func (g *graph) emitEdge(label string, payload Edge) {
	if payload.InV == 0 && len(payload.InVs) == 0 {
		panic("no inVs")
	}
	g.emit("edge", label, payload)
}

func (g *graph) emit(ty, label string, payload any) int {
	g.ID++
	g.Elements = append(g.Elements, Element{
		ID:      g.ID,
		Type:    ty,
		Label:   label,
		Payload: payload,
	})
	return g.ID
}

// registerInverseRelationships records symbol relationships from parent symbols to children symbols.
// For example, a struct (child) that implements an interface A (parent) encodes that child->parent
// relationship with LSIF Typed via the field `SymbolInformation.Relationships`.
// registerInverseRelationships method records the relationship in the opposite direction: parent->child.
func (g *graph) registerInverseRelationships(info *lsiftyped.SymbolInformation) {
	for _, relationship := range info.Relationships {
		inverseRelationships := g.inverseRelationships[relationship.Symbol]
		g.inverseRelationships[relationship.Symbol] = append(inverseRelationships, &lsiftyped.Relationship{
			Symbol:           info.Symbol,
			IsReference:      relationship.IsReference,
			IsImplementation: relationship.IsImplementation,
			IsTypeDefinition: relationship.IsTypeDefinition,
		})
	}
}

// interpretLsifRange handles the difference between single-line and multi-line encoding of range positions.
func interpretLsifRange(lsifRange []int32) (startLine, startCharacter, endLine, endCharacter int32, err error) {
	if len(lsifRange) == 3 {
		return lsifRange[0], lsifRange[1], lsifRange[0], lsifRange[2], nil
	}
	if len(lsifRange) == 4 {
		return lsifRange[0], lsifRange[1], lsifRange[2], lsifRange[3], nil
	}
	return 0, 0, 0, 0, errors.Newf("invalid LSIF range %v", lsifRange)
}

func (g *graph) getOrInsertSymbolInformationIDs(symbol string, localResultSetTable map[string]symbolInformationIDs) symbolInformationIDs {
	resultSetTable := g.symbolToResultSet
	if lsiftyped.IsLocalSymbol(symbol) {
		resultSetTable = localResultSetTable
	}
	ids, ok := resultSetTable[symbol]
	if !ok {
		ids = g.emitResultSet(&lsiftyped.SymbolInformation{Symbol: symbol}, "import")
		resultSetTable[symbol] = ids
	}
	return ids
}

func WriteNDJSON(elements []jsonElement, out io.Writer) error {
	w := writer.NewJSONWriter(out)
	for _, e := range elements {
		w.Write(e)
	}
	return w.Flush()
}

type jsonHoverContent struct {
	Kind  string `json:"kind,omitempty"`
	Value string `json:"value,omitempty"`
}
type jsonHoverResult struct {
	Contents jsonHoverContent `json:"contents"`
}
type jsonToolInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// jsonElement is similar to Element but it can be serialized to JSON to emit valid LSIF Graph output.
type jsonElement struct {
	ID               int              `json:"id"`
	Name             string           `json:"name,omitempty"`
	Version          string           `json:"version,omitempty"`
	Manager          string           `json:"manager,omitempty"`
	ProjectRoot      string           `json:"projectRoot,omitempty"`
	PositionEncoding string           `json:"positionEncoding,omitempty"`
	ToolInfo         *jsonToolInfo    `json:"toolInfo,omitempty"`
	Type             string           `json:"type,omitempty"`
	Label            string           `json:"label,omitempty"`
	Result           *jsonHoverResult `json:"result,omitempty"`
	Uri              string           `json:"uri,omitempty"`
	Start            *protocol.Pos    `json:"start,omitempty"`
	End              *protocol.Pos    `json:"end,omitempty"`
	InV              int              `json:"inV,omitempty"`
	InVs             []int            `json:"inVs,omitempty"`
	OutV             int              `json:"outV,omitempty"`
	Document         int              `json:"document,omitempty"`
	Identifier       string           `json:"identifier,omitempty"`
	Kind             string           `json:"kind,omitempty"`
	Scheme           string           `json:"scheme,omitempty"`
}

func ElementsToJsonElements(els []Element) []jsonElement {
	var r []jsonElement
	for _, el := range els {
		object := jsonElement{
			ID:    el.ID,
			Type:  el.Type,
			Label: el.Label,
		}
		if el.Type == "edge" {
			edge := el.Payload.(Edge)
			object.OutV = edge.OutV
			object.InV = edge.InV
			object.InVs = edge.InVs
			object.Document = edge.Document
		} else if el.Type == "vertex" {
			switch el.Label {
			case "hoverResult":
				object.Result = &jsonHoverResult{Contents: jsonHoverContent{
					Kind:  "markdown",
					Value: el.Payload.(string),
				}}
			case "document":
				object.Uri = el.Payload.(string)
			case "range":
				rng := el.Payload.(Range)
				object.Start = &rng.Start
				object.End = &rng.End
			case "metaData":
				metaData := el.Payload.(MetaData)
				object.Version = metaData.Version
				object.ProjectRoot = metaData.ProjectRoot
				object.PositionEncoding = metaData.PositionEncoding
				object.ToolInfo = &jsonToolInfo{
					Name:    metaData.ToolInfo.Name,
					Version: metaData.ToolInfo.Version,
				}
			case "moniker":
				moniker := el.Payload.(Moniker)
				object.Identifier = moniker.Identifier
				object.Kind = moniker.Kind
				object.Scheme = moniker.Scheme
			case "packageInformation":
				pkg := el.Payload.(PackageInformation)
				object.Name = pkg.Name
				object.Version = pkg.Version
				object.Manager = pkg.Manager
			case "definitionResult",
				"implementationResult",
				"referenceResult",
				"referenceResults",
				"resultSet",
				"textDocument/references",
				"textDocument/hover",
				"textDocument/definition":
			default:
				panic(fmt.Sprintf("unexpected LSIF element: %+v", el))
			}
		} else {
			panic(el.Type)
		}
		r = append(r, object)
	}
	return r
}
