package protocol

// Sourcegraph extension to LSIF: documentation.
// See https://github.com/slimsag/language-server-protocol/pull/2

// DocumentationEdge is an edge which connects a Documentation Vertex to a Project
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

// A Documentation vertex describes hierarchial project-wide documentation.
// It represents documentation for a programming construct (variable, function, etc.) or group of
// programming constructs in a workspace (library, package, crate, module, etc.)
//
// The exact structure of the documentation depends on what makes sense for the specific language
// and concepts being described.
type Documentation struct {
	Vertex

	// A human-readable URL slug identifier for this documentation. It should be unique relative to
	// sibling Documentation.
	Slug string `json:"slug"`

	// Whether or not this Documentation is the beginning of a new major section, meaning it and its
	// its children should be e.g. displayed on their own dedicate page.
	NewPage bool `json:"newPage"`

	// A single-line label to display for this documentation in e.g. the index of a book. For
	// example, the name of a group of documentation, the name of a library, the signature of a
	// function or class, etc.
	Title InteractiveMarkupContent `json:"title"`

	// A detailed multi-line string that contains detailed documentation for the section described by
	// the title.
	Detail InteractiveMarkupContent `json:"detail"`

	// Tags about the type of content this documentation contains.
	Tags []DocumentationTag `json:"tags"`

	// Documentation that should be logically nested below this documentation itself, expressed as
	// "documentation" vertex IDs. For example, this documentation may describe a class and have
	// children describing each method of the class.
	Children []uint64 `json:"children"`
}

// NewDocumentation initializes a "documentation" vertex with the given ID. The caller should then
// populate the required fields.
func NewDocumentation(id uint64) Documentation {
	return Documentation{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexDocumentation,
		},
	}
}

type InteractiveMarkupContent struct {
	// The actual content which should be considered interactive.
	Content MarkupContent `json:"content"`

	// Ranges in the `content` string mapping to an associated "documentation" vertex ID, allowing
	// for text in one piece of documentation to link to another.
	Documentation map[Range]uint64 `json:"documentation"`

	// Ranges in the `content` string mapping to an associated "hoverResult" vertex ID, allowing
	// for text in one piece of documentation to include a hover result.
	//
	// The hover result could be specific to the documentation itself, or e.g. part of a type
	// signature being hovered over.
	Hover map[Range]uint64 `json:"hover"`

	// Ranges in the `content` string mapping to an associated "definitionResult" vertex ID,
	// allowing for text in one piece of documentation to include a definition result.
	//
	// This enables users to e.g. go-to-definition on a type signature inside documentation.
	Definition map[Range]uint64 `json:"definition"`

	// Ranges in the `content` string mapping to an associated "referenceResult" vertex ID,
	// allowing for text in one piece of documentation to include a definition result.
	//
	// This enables users to e.g. find-references on a type signature inside documentation.
	References map[Range]uint64 `json:"references"`
}

type DocumentationTag string

const (
	// The documentation describes a concept that is exported externally.
	DocumentationExported DocumentationTag = "exported"

	// The documentation describes a concept that is unexported / internal.
	DocumentationUnexported DocumentationTag = "unexported"

	// The documentation describes a concept that is deprecated.
	DocumentationDeprecated DocumentationTag = "deprecated"
)
