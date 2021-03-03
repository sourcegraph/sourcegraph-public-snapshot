package protocol

type Element struct {
	ID   uint64      `json:"id"`
	Type ElementType `json:"type"`
}

type ElementType string

const (
	ElementVertex ElementType = "vertex"
	ElementEdge   ElementType = "edge"
)

type Vertex struct {
	Element
	Label VertexLabel `json:"label"`
}

type VertexLabel string

const (
	VertexMetaData             VertexLabel = "metaData"
	VertexProject              VertexLabel = "project"
	VertexRange                VertexLabel = "range"
	VertexLocation             VertexLabel = "location"
	VertexDocument             VertexLabel = "document"
	VertexMoniker              VertexLabel = "moniker"
	VertexPackageInformation   VertexLabel = "packageInformation"
	VertexResultSet            VertexLabel = "resultSet"
	VertexDocumentSymbolResult VertexLabel = "documentSymbolResult"
	VertexFoldingRangeResult   VertexLabel = "foldingRangeResult"
	VertexDocumentLinkResult   VertexLabel = "documentLinkResult"
	VertexDianosticResult      VertexLabel = "diagnosticResult"
	VertexDeclarationResult    VertexLabel = "declarationResult"
	VertexDefinitionResult     VertexLabel = "definitionResult"
	VertexTypeDefinitionResult VertexLabel = "typeDefinitionResult"
	VertexHoverResult          VertexLabel = "hoverResult"
	VertexReferenceResult      VertexLabel = "referenceResult"
	VertexImplementationResult VertexLabel = "implementationResult"
)

type Edge struct {
	Element
	Label EdgeLabel `json:"label"`
}

type EdgeLabel string

const (
	EdgeContains                   EdgeLabel = "contains"
	EdgeItem                       EdgeLabel = "item"
	EdgeNext                       EdgeLabel = "next"
	EdgeMoniker                    EdgeLabel = "moniker"
	EdgeNextMoniker                EdgeLabel = "nextMoniker"
	EdgePackageInformation         EdgeLabel = "packageInformation"
	EdgeTextDocumentDocumentSymbol EdgeLabel = "textDocument/documentSymbol"
	EdgeTextDocumentFoldingRange   EdgeLabel = "textDocument/foldingRange"
	EdgeTextDocumentDocumentLink   EdgeLabel = "textDocument/documentLink"
	EdgeTextDocumentDiagnostic     EdgeLabel = "textDocument/diagnostic"
	EdgeTextDocumentDefinition     EdgeLabel = "textDocument/definition"
	EdgeTextDocumentDeclaration    EdgeLabel = "textDocument/declaration"
	EdgeTextDocumentTypeDefinition EdgeLabel = "textDocument/typeDefinition"
	EdgeTextDocumentHover          EdgeLabel = "textDocument/hover"
	EdgeTextDocumentReferences     EdgeLabel = "textDocument/references"
	EdgeTextDocumentImplementation EdgeLabel = "textDocument/implementation"
)
