package protocol

// Sourcegraph extension to LSIF: documentation.
// See https://github.com/slimsag/language-server-protocol/pull/2

// A "documentationResult" edge connects a "project" or "resultSet" vertex to a
// "documentationResult" vertex.
//
// It allows one to attach extensive documentation to a project, range (via being attached to a
// "resultSet" vertex) or another "documentationResult" for representing hierarchical documentation
// like:
//
// "project" (e.g. an HTTP library)
// -> "documentationResult" (e.g. "HTTP library" library documentation)
//   -> "documentationResult" (e.g. docs for the "Server" class in the HTTP library)
//     -> "documentationResult" (e.g. docs for the "Listen" method on the "Server" class)
//
type DocumentationResultEdge struct {
	Edge
	
	// A "project" or "resultSet" vertex ID.
	InV  uint64 `json:"inV"`

	// The "sourcegraph:documentationResult" vertex ID.
	OutV uint64 `json:"outV"`
}

func NewDocumentationResultEdge(id, inV, outV uint64) DocumentationResultEdge {
	return DocumentationResultEdge{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeSourcegraphDocumentationResult,
		},
		OutV: outV,
		InV:  inV,
	}
}

// A "documentationResult" vertex 
type DocumentationResult struct {
	Vertex
	Result Documentation `json:"result"`
}

// NewDocumentationResult creates a new "sourcegraph:documentationResult" vertex.
func NewDocumentationResult(id uint64, result Documentation) DocumentationResult {
	return DocumentationResult{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexSourcegraphDocumentationResult,
		},
		Result: result,
	}
}

// A "documentationResult" vertex describes hierarchial project-wide documentation. It represents
// documentation for a programming construct (variable, function, etc.) or group of programming
// constructs in a workspace (library, package, crate, module, etc.)
//
// The exact structure of the documentation depends on what makes sense for the specific language
// and concepts being described.
//
// Attached to this vertex MUST be two "documentationString" vertices:
//
// 1. The {type: "label"} string, a one-line label or this section of documentation.
// 1. The {type: "detail"} string, a multi-line detailed string for this section of documentation.
//
// They can be attached using a "documentationStringEdge".
type Documentation struct {
	// A human-readable URL slug identifier for this documentation. It should be unique relative to
	// sibling Documentation.
	Slug string `json:"slug"`

	// Whether or not this Documentation is the beginning of a new major section, meaning it and its
	// its children should be e.g. displayed on their own dedicate page.
	NewPage bool `json:"newPage"`

	// Tags about the type of content this documentation contains.
	Tags []DocumentationTag `json:"tags"`
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

// A "documentationStringEdge" is an edge which connects a "documentationResult" vertex to its
// label or detail strings, which are "documentationString" vertices.
type DocumentationStringEdge struct {
	Edge
	
	// The "documentationResult" vertex ID.
	InV  uint64 `json:"inV"`

	// The "documentationString" vertex ID.
	OutV uint64 `json:"outV"`

	// Whether this links the "label"" or "detail" string of the documentation.
	Type DocumentationStringType `json:"type"`
}

type DocumentationStringType string

const (
	// A single-line label to display for this documentation in e.g. the index of a book. For
	// example, the name of a group of documentation, the name of a library, the signature of a
	// function or class, etc.
	DocumentationStringTypeLabel DocumentationStringType = "label"

	// A detailed multi-line string that contains detailed documentation for the section described by
	// the title.
	DocumentationStringTypeDetail DocumentationStringType = "detail"
)

func NewDocumentationStringEdge(id, inV, outV uint64) DocumentationStringEdge {
	return DocumentationStringEdge{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeSourcegraphDocumentationString,
		},
		OutV: outV,
		InV:  inV,
	}
}

// A "documentationString" vertex is referred to by a "documentationResult" vertex using a
// "documentationString" edge. It represents the actual string of content for the documentation's
// label (a one-line string) or detail (a multi-line string).
//
// A "documentationString" vertex can itself be linked to "range" vertices (which describe a range
// in the documentation string's markup content itself) using a ??????WHAT EDGE??????. This enables
// ranges within a documentation string to have:
//
// * "hoverResult"s (e.g. you can hover over a type signature in the documentation string and get info)
// * "definitionResult" and "referenceResults"
// * "documentationResult" itself - allowing a range of text in one documentation to link to another
//   documentation section (e.g. in the same way a hyperlink works in HTML.)
// * "moniker" to link to another project's hover/definition/documentation results, across
//   repositories.
//
type DocumentationString struct {
	Vertex
	Result MarkupContent `json:"result"`
}

// NewDocumentationString creates a new "documentationString" vertex.
func NewDocumentationString(id uint64, result MarkupContent) DocumentationString {
	return DocumentationString{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Label: VertexSourcegraphDocumentationString,
		},
		Result: result,
	}
}