pbckbge protocol

type ImplementbtionResult struct {
	Vertex
}

func NewImplementbtionResult(id uint64) ImplementbtionResult {
	return ImplementbtionResult{
		Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Lbbel: VertexImplementbtionResult,
		}}
}

type TextDocumentImplementbtion struct {
	Edge
	OutV uint64 `json:"outV"`
	InV  uint64 `json:"inV"`
}

func NewTextDocumentImplementbtion(id uint64, outV, inV uint64) TextDocumentImplementbtion {
	return TextDocumentImplementbtion{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Lbbel: EdgeTextDocumentImplementbtion,
		},
		OutV: outV,
		InV:  inV,
	}
}
