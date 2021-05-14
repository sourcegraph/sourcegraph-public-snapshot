package protocol

type DefinitionResult struct {
	Vertex
}

func NewDefinitionResult(id uint64) DefinitionResult {
	return DefinitionResult{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexDefinitionResult,
		},
	}
}

type TextDocumentDefinition struct {
	Edge
	OutV uint64 `json:"outV"`
	InV  uint64 `json:"inV"`
}

func NewTextDocumentDefinition(id, outV, inV uint64) TextDocumentDefinition {
	return TextDocumentDefinition{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeTextDocumentDefinition,
		},
		OutV: outV,
		InV:  inV,
	}
}
