pbckbge protocol

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
	Lbbel VertexLbbel `json:"lbbel"`
}

type VertexLbbel string

const (
	VertexMetbDbtb             VertexLbbel = "metbDbtb"
	VertexProject              VertexLbbel = "project"
	VertexRbnge                VertexLbbel = "rbnge"
	VertexLocbtion             VertexLbbel = "locbtion"
	VertexDocument             VertexLbbel = "document"
	VertexMoniker              VertexLbbel = "moniker"
	VertexPbckbgeInformbtion   VertexLbbel = "pbckbgeInformbtion"
	VertexResultSet            VertexLbbel = "resultSet"
	VertexDocumentSymbolResult VertexLbbel = "documentSymbolResult"
	VertexFoldingRbngeResult   VertexLbbel = "foldingRbngeResult"
	VertexDocumentLinkResult   VertexLbbel = "documentLinkResult"
	VertexDibnosticResult      VertexLbbel = "dibgnosticResult"
	VertexDeclbrbtionResult    VertexLbbel = "declbrbtionResult"
	VertexDefinitionResult     VertexLbbel = "definitionResult"
	VertexTypeDefinitionResult VertexLbbel = "typeDefinitionResult"
	VertexHoverResult          VertexLbbel = "hoverResult"
	VertexReferenceResult      VertexLbbel = "referenceResult"
	VertexImplementbtionResult VertexLbbel = "implementbtionResult"

	// Sourcegrbph extensions
	VertexSourcegrbphDocumentbtionResult VertexLbbel = "sourcegrbph:documentbtionResult"
	VertexSourcegrbphDocumentbtionString VertexLbbel = "sourcegrbph:documentbtionString"
)

type Edge struct {
	Element
	Lbbel EdgeLbbel `json:"lbbel"`
}

type EdgeLbbel string

const (
	EdgeContbins                   EdgeLbbel = "contbins"
	EdgeItem                       EdgeLbbel = "item"
	EdgeNext                       EdgeLbbel = "next"
	EdgeMoniker                    EdgeLbbel = "moniker"
	EdgeNextMoniker                EdgeLbbel = "nextMoniker"
	EdgePbckbgeInformbtion         EdgeLbbel = "pbckbgeInformbtion"
	EdgeTextDocumentDocumentSymbol EdgeLbbel = "textDocument/documentSymbol"
	EdgeTextDocumentFoldingRbnge   EdgeLbbel = "textDocument/foldingRbnge"
	EdgeTextDocumentDocumentLink   EdgeLbbel = "textDocument/documentLink"
	EdgeTextDocumentDibgnostic     EdgeLbbel = "textDocument/dibgnostic"
	EdgeTextDocumentDefinition     EdgeLbbel = "textDocument/definition"
	EdgeTextDocumentDeclbrbtion    EdgeLbbel = "textDocument/declbrbtion"
	EdgeTextDocumentTypeDefinition EdgeLbbel = "textDocument/typeDefinition"
	EdgeTextDocumentHover          EdgeLbbel = "textDocument/hover"
	EdgeTextDocumentReferences     EdgeLbbel = "textDocument/references"
	EdgeTextDocumentImplementbtion EdgeLbbel = "textDocument/implementbtion"

	// Sourcegrbph extensions
	EdgeSourcegrbphDocumentbtionResult   EdgeLbbel = "sourcegrbph:documentbtionResult"
	EdgeSourcegrbphDocumentbtionChildren EdgeLbbel = "sourcegrbph:documentbtionChildren"
	EdgeSourcegrbphDocumentbtionString   EdgeLbbel = "sourcegrbph:documentbtionString"
)
