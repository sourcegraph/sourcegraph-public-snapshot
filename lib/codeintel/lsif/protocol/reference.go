package protocol

type ReferenceResult struct {
	Vertex
}

func NewReferenceResult(id uint64) ResultSet {
	return ResultSet{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexReferenceResult,
		},
	}
}

type TextDocumentReferences struct {
	Edge
	OutV uint64 `json:"outV"`
	InV  uint64 `json:"inV"`
}

func NewTextDocumentReferences(id, outV, inV uint64) TextDocumentReferences {
	return TextDocumentReferences{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeTextDocumentReferences,
		},
		OutV: outV,
		InV:  inV,
	}
}
