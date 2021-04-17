package protocol

// Documentation is a Vertex which includes a DocumentationResult.
type Documentation struct {
	Vertex
	Result *DocumentationResult `json:"result"`
}

func NewDocumentation(id uint64, result *DocumentationResult) Documentation {
	return Documentation{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexDocumentation,
		},
		Result: result,
	}
}

// DocumentationEdge is an edge which connects a Documentation Vertex to a Moniker
// Vertex.
type DocumentationEdge struct {
	Edge
	InV  uint64 `json:"inV"`
	OutV uint64 `json:"outV"`
}

func NewDocumentationEdge(id, inV, outV uint64) DocumentationEdge {
	return DocumentationEdge{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeDocumentation,
		},
		OutV: outV,
		InV:  inV,
	}
}

type DocumentationResult struct {
	// Slug is a human-readable URL slug identifier for this documentation. It should be unique
	// relative to sibling Documentation.
	Slug          string
	NewChapter    bool
	Label, Detail InteractiveMarkupContent
	Visibility    DocumentationVisibility

	// VertexDocumentation IDs
	Children []uint64
}

type InteractiveMarkupContent struct {
	Content MarkupContent

	// Content Range -> VertexDocumentation ID
	Documentation map[Range]uint64

	// Content Range -> VertexHoverResult ID
	Hover map[Range]uint64

	// Content Range -> VertexDefinitionResult ID
	Definition map[Range]uint64

	// Content Range -> VertexReferencesResult ID
	References map[Range]uint64
}

type DocumentationVisibility string

const (
	Deprecated DocumentationVisibility = "deprecated"
	Exported   DocumentationVisibility = "exported"
	Unexported DocumentationVisibility = "unexported"
)
