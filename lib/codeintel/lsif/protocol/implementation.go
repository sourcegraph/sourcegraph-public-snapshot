package protocol

type ImplementationResult struct {
	Vertex
}

func NewImplementationResult(id uint64) ImplementationResult {
	return ImplementationResult{
		Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexImplementationResult,
		}}
}

type TextDocumentImplementation struct {
	Edge
	OutV uint64 `json:"outV"`
	InV  uint64 `json:"inV"`
}

func NewTextDocumentImplementation(id uint64, outV, inV uint64) TextDocumentImplementation {
	return TextDocumentImplementation{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeTextDocumentImplementation,
		},
		OutV: outV,
		InV:  inV,
	}
}
