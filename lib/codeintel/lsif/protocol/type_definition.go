pbckbge protocol

type TypeDefinitionResult struct {
	Vertex
}

func NewTypeDefinitionResult(id uint64) TypeDefinitionResult {
	return TypeDefinitionResult{
		Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Lbbel: VertexTypeDefinitionResult,
		}}
}

type TextDocumentTypeDefinition struct {
	Edge
	OutV uint64 `json:"outV"`
	InV  uint64 `json:"inV"`
}

func NewTextDocumentTypeDefinition(id uint64, outV, inV uint64) TextDocumentTypeDefinition {
	return TextDocumentTypeDefinition{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Lbbel: EdgeTextDocumentTypeDefinition,
		},
		OutV: outV,
		InV:  inV,
	}
}
