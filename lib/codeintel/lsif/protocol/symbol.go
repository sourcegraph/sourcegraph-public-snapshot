package protocol

type RangeBasedDocumentSymbol struct {
	// ID is the range ID associated with this symbol.
	ID       uint64                      `json:"id"`
	Children []*RangeBasedDocumentSymbol `json:"children,omitempty"`
}

type DocumentSymbolResult struct {
	Vertex

	// Note: the LSIF spec also says Result can be an array of lsp.DocumentSymbol instances, but we
	// don't yet support that here.
	Result []*RangeBasedDocumentSymbol `json:"result"`
}

func NewDocumentSymbolResult(id uint64, result []*RangeBasedDocumentSymbol) DocumentSymbolResult {
	return DocumentSymbolResult{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexDocumentSymbolResult,
		},
		Result: result,
	}
}

type DocumentSymbolEdge struct {
	Edge
	InV  uint64 `json:"inV"`
	OutV uint64 `json:"outV"`
}

func NewDocumentSymbolEdge(id, inV, outV uint64) DocumentSymbolEdge {
	return DocumentSymbolEdge{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeTextDocumentDocumentSymbol,
		},
		OutV: outV,
		InV:  inV,
	}
}

// bazel run //lib/lsif/protocol:write_symbol_kind (or bazel run //dev:write_all_generated)
// SymbolKind corresponds to lsp.SymbolKind in the LSP spec. See
// https://microsoft.github.io/language-server-protocol/specifications/specification-3-17/#textDocument_documentSymbol
type SymbolKind int

const (
	File          SymbolKind = 1
	Module        SymbolKind = 2
	Namespace     SymbolKind = 3
	Package       SymbolKind = 4
	Class         SymbolKind = 5
	Method        SymbolKind = 6
	Property      SymbolKind = 7
	Field         SymbolKind = 8
	Constructor   SymbolKind = 9
	Enum          SymbolKind = 10
	Interface     SymbolKind = 11
	Function      SymbolKind = 12
	Variable      SymbolKind = 13
	Constant      SymbolKind = 14
	String        SymbolKind = 15
	Number        SymbolKind = 16
	Boolean       SymbolKind = 17
	Array         SymbolKind = 18
	Object        SymbolKind = 19
	Key           SymbolKind = 20
	Null          SymbolKind = 21
	EnumMember    SymbolKind = 22
	Struct        SymbolKind = 23
	Event         SymbolKind = 24
	Operator      SymbolKind = 25
	TypeParameter SymbolKind = 26
)

// bazel run //lib/lsif/protocol:write_symbol_tag (or bazel run //dev:write_all_generated)
// SymbolTag corresponds to lsp.SymbolTag in the LSP spec. See
// https://microsoft.github.io/language-server-protocol/specifications/specification-3-17/#textDocument_documentSymbol
type SymbolTag int

const (
	Deprecated SymbolTag = 1

	// These are custom extensions, see https://github.com/microsoft/language-server-protocol/issues/98
	Exported   SymbolTag = 100
	Unexported SymbolTag = 101
)
