package protocol

// Sourcegraph extension to LSIF: documentation.
// See https://github.com/slimsag/language-server-protocol/pull/2

// A "documentationResult" edge connects a "project" or "resultSet" vertex to a
// "documentationResult" vertex.
//
// It allows one to attach extensive documentation to a project or range (via being attached to a
// "resultSet" vertex). Combined with the "documentationChildren" edge, this can be used to
// represent hierarchical documentation.
type DocumentationResultEdge struct {
	Edge

	// The "documentationResult" vertex ID.
	InV uint64 `json:"inV"`

	// A "project" or "resultSet" vertex ID.
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

// A "documentationChildren" edge connects one "documentationResult" vertex (the parent) to its
// children "documentationResult" vertices.
//
// It allows one represent hierarchical documentation like:
//
//	"project" (e.g. an HTTP library)
//	-> "documentationResult" (e.g. "HTTP library" library documentation)
//	  -> "documentationResult" (e.g. docs for the "Server" class in the HTTP library)
//	    -> "documentationResult" (e.g. docs for the "Listen" method on the "Server" class)
//	    -> "documentationResult" (e.g. docs for the "Shutdown" method on the "Server" class)
//	      -> ...
//
// Note: the "project" -> "documentationResult" attachment above is expressed via a
// "documentationResult" edge, since the parent is not a "documentationResult" vertex.
type DocumentationChildrenEdge struct {
	Edge

	// The ordered children "documentationResult" vertex IDs.
	InVs []uint64 `json:"inVs"`

	// The parent "documentationResult" vertex ID.
	OutV uint64 `json:"outV"`
}

func NewDocumentationChildrenEdge(id uint64, inVs []uint64, outV uint64) DocumentationChildrenEdge {
	return DocumentationChildrenEdge{
		Edge: Edge{
			Element: Element{
				ID:   id,
				Type: ElementEdge,
			},
			Label: EdgeSourcegraphDocumentationChildren,
		},
		OutV: outV,
		InVs: inVs,
	}
}

// A "documentationResult" vertex
type DocumentationResult struct {
	Vertex
	Result Documentation `json:"result"`
}

// NewDocumentationResult creates a new "documentationResult" vertex.
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
//  1. A "documentationString" vertex with `type: "label"`, which is a one-line label or this section
//     of documentation.
//  1. A "documentationString" vertex with `type: "detail"`, which is a multi-line detailed string
//     for this section of documentation.
//
// Both are attached to the documentationResult via a "documentationString" edge.
//
// If the label or detail vertex is missing, or the label string is empty (has no content) then a
// client should consider all "documentationResult" vertices in the entire LSIF dump to be invalid
// and malformed, and ignore them.
//
// If no detail is available (such as a function with no documentation), a `type:"detail"`
// "documentationString" should still be emitted - but with an empty string for the MarkupContent.
// This enables validators to ensure the indexer knows how to emit both label and detail strings
// properly, and just chose to emit none specifically.
//
// If this documentationResult is for the project root, the identifier and searchKey should be an
// empty string.
//
// If a pages' only purpose is to connect other pages below it (i.e. it is an index page), it
// should have empty label and detail strings attached.
type Documentation struct {
	// A human readable identifier for this documentationResult, uniquely identifying it within the
	// scope of the parent page (or an empty string, if this is the root documentationResult.)
	//
	// Clients may build a path identifiers to a specific documentationResult _page_ by joining the
	// identifiers of each documentationResult with `newPage: true` starting from the desired root
	// until the target page is reached. For example, if trying to build a paths from the workspace
	// root to a Go method "ServeHTTP" on a "Router" struct, you may have the following
	// documentationResults describing the Go package structure:
	//
	//  	[
	//  	  {identifier: "",                 newPage: true},
	//  	  {identifier: "internal",         newPage: true},
	//  	  {identifier: "pkg",              newPage: true},
	//  	  {identifier: "mux",              newPage: true},
	//  	  {identifier: "Router",           newPage: false},
	//  	  {identifier: "Router.ServeHTTP", newPage: false},
	//  	]
	//
	// The first entry (identifier="") is root documentationResult of the workspace. Note that
	// since identifiers are unique relative to the parent page, the `Router` struct and the
	// `Router.ServeHTTP` method have unique identifiers relative to the parent `mux` page.
	// Thus, to build a path to either simply join all the `newPage: true` identifiers
	// ("/internal/pkg/mux") and use the identifier of any child once `newPage: false` is reached:
	//
	//  	/internal/pkg/mux#Router
	//  	/internal/pkg/mux#Router.ServeHTTP
	//
	// The identifier is relative to the parent page so that language indexers may choose to format
	// the effective e.g. URL hash in a way that makes sense in the given language, e.g. C++ for
	// example could use `Router::ServeHTTP` instead of a "." joiner.
	//
	// An identifier may contain any characters, including slashes and "#". If clients intend to
	// use identifiers in a context where those characters are forbidden (e.g. URLs) then they must
	// replace them with something else.
	Identifier string `json:"identifier"`

	// Whether or not this Documentation is the beginning of a new major section, meaning it and its
	// its children should be e.g. displayed on their own dedicated page.
	NewPage bool `json:"newPage"`

	// SearchKey is a key which can be used to implement search for a specific documentationResult.
	// For example, in Go this may look like `mux.Router` or `mux.Router.ServeHTTP`. It should be
	// of a format that makes sense to users of the language being documented.
	//
	// Search keys are not required to be unique. It is desirable for them to be generally unique
	// within the scope of the workspace itself, or a project within the workspace (if the
	// documentation is for something in a project.) However, it is not desirable for it to be unique
	// globally across workspaces (you can think of the searchKey as always being prefixed with the
	// workspace URI.)
	//
	// If a search key is describing a project within the workspace itself, it is encouraged for it
	// to be unique within the context of the workspace. Sometimes this means using a full project
	// path/name (e.g. `github.com/gorilla/mux/router` or `com.JodaOrg.JodaTime`) is required - while
	// in other contexts the shortened name (`router` or `JodaTime`) may be sufficient.
	//
	// If a search key is describing a symbol within a project, a shortened project path/name prefix
	// is usually sufficient: using `router.New` over `github.com/gorilla/mux/router.New` or
	// `JodaTime.Time.now` over `com.JodaOrg.JodaTime.Time.now` is preferred. Clients will display
	// enough additional information to disambiguiate between any conflicts (see below.)
	//
	// Clients are encouraged to treat matches towards the left of the string with higher relevance
	// than matches towards the end of the string. For example, it is typically the case that search
	// keys will start with the project/package/library/etc name, followed by namespaces, then a
	// specific symbol. For example, if a user searches for `gorilla/mux.Error` the desired ranking
	// for three theoretical semi-conflicting results would be:
	//
	// * github.com/gorilla/mux.Error (near exact match)
	// * github.com/gorilla/router.Error (`mux` not matching on left, result ranked lower)
	// * github.com/sourcegraph/mux.Error (`gorilla` not matching on left, result ranked lower)
	//
	// Clients are encouraged to use smart case sensitivity by default: if the user is searching for
	// a mixed-case query, the search should be case-sensitive (and otherwise not.)
	//
	// Since search keys may not be unique, clients are encouraged to display alongside the search
	// key other information about the documentation that will disambiguate identical keys. The
	// following in specific is encouraged:
	//
	// * Always display the `label` string, which provides e.g. a one-line function signature.
	// * Optionally display the `detail` string (e.g. when considering a specific result), as it
	//   contains detailed information that can help disambiguate.
	// * Always display the path identifier to the documentationResult _somewhere_, even if it is
	//   a much more subtle location (see `identifier` docs), as it is a truly unique path to the
	//   documentation and can be a final way for users to disambiguate if all other options fail.
	//
	// An empty string indicates this documentationResult should not be indexed by a search engine.
	SearchKey string `json:"searchKey"`

	// Tags about the type of content this documentation contains.
	Tags []Tag `json:"tags"`
}

type Tag string

const (
	// The documentation describes a concept that is private/unexported, not a public/exported
	// concept.
	TagPrivate Tag = "private"

	// The documentation describes a concept that is deprecated.
	TagDeprecated Tag = "deprecated"

	// The documentation describes e.g. a test function or concept related to testing.
	TagTest Tag = "test"

	// The documentation describes e.g. a benchmark function or concept related to benchmarking.
	TagBenchmark Tag = "benchmark"

	// The documentation describes e.g. an example function or example code.
	TagExample Tag = "example"

	// The documentation describes license information.
	TagLicense Tag = "license"

	// The documentation describes owner information.
	TagOwner Tag = "owner"

	// The documentation describes package/library registry information, e.g. where a package is
	// available for download.
	TagRegistryInfo Tag = "owner"

	// Tags derived from SymbolKind
	TagFile          Tag = "file"
	TagModule        Tag = "module"
	TagNamespace     Tag = "namespace"
	TagPackage       Tag = "package"
	TagClass         Tag = "class"
	TagMethod        Tag = "method"
	TagProperty      Tag = "property"
	TagField         Tag = "field"
	TagConstructor   Tag = "constructor"
	TagEnum          Tag = "enum"
	TagInterface     Tag = "interface"
	TagFunction      Tag = "function"
	TagVariable      Tag = "variable"
	TagConstant      Tag = "constant"
	TagString        Tag = "string"
	TagNumber        Tag = "number"
	TagBoolean       Tag = "boolean"
	TagArray         Tag = "array"
	TagObject        Tag = "object"
	TagKey           Tag = "key"
	TagNull          Tag = "null"
	TagEnumNumber    Tag = "enumNumber"
	TagStruct        Tag = "struct"
	TagEvent         Tag = "event"
	TagOperator      Tag = "operator"
	TagTypeParameter Tag = "typeParameter"
)

// A "documentationString" edge connects a "documentationResult" vertex to its label or detail
// strings, which are "documentationString" vertices. The overall structure looks like the
// following roughly:
//
//	{id: 53, type:"vertex", label:"documentationResult", result:{identifier:"httpserver", ...}}
//	{id: 54, type:"vertex", label:"documentationString", result:{kind:"plaintext", "value": "A single-line label for an HTTPServer instance"}}
//	{id: 55, type:"vertex", label:"documentationString", result:{kind:"plaintext", "value": "A multi-line\n detailed\n explanation of an HTTPServer instance, what it does, etc."}}
//	{id: 54, type:"edge", label:"documentationString", inV: 54, outV: 53, kind:"label"}
//	{id: 54, type:"edge", label:"documentationString", inV: 55, outV: 53, kind:"detail"}
//
// Hover, definition, etc. results can then be attached to ranges within the "documentationString"
// vertices themselves (vertex 54 / 55), see the docs for DocumentationString for more details.
type DocumentationStringEdge struct {
	Edge

	// The "documentationString" vertex ID.
	InV uint64 `json:"inV"`

	// The "documentationResult" vertex ID.
	OutV uint64 `json:"outV"`

	// Whether this links the "label" or "detail" string of the documentation.
	Kind DocumentationStringKind `json:"kind"`
}

type DocumentationStringKind string

const (
	// A single-line label to display for this documentation in e.g. the index of a book. For
	// example, the name of a group of documentation, the name of a library, the signature of a
	// function or class, etc.
	DocumentationStringKindLabel DocumentationStringKind = "label"

	// A detailed multi-line string that contains detailed documentation for the section described by
	// the title.
	DocumentationStringKindDetail DocumentationStringKind = "detail"
)

func NewDocumentationStringEdge(id, inV, outV uint64, kind DocumentationStringKind) DocumentationStringEdge {
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
		Kind: kind,
	}
}

// A "documentationString" vertex is referred to by a "documentationResult" vertex using a
// "documentationString" edge. It represents the actual string of content for the documentation's
// label (a one-line string) or detail (a multi-line string).
//
// A "documentationString" vertex can itself be linked to "range" vertices (which describe a range
// in the documentation string's markup content itself) using a "contains" edge. This enables
// ranges within a documentation string to have:
//
//   - "hoverResult"s (e.g. you can hover over a type signature in the documentation string and get info)
//   - "definitionResult" and "referenceResults"
//   - "documentationResult" itself - allowing a range of text in one documentation to link to another
//     documentation section (e.g. in the same way a hyperlink works in HTML.)
//   - "moniker" to link to another project's hover/definition/documentation results, across
//     repositories.
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
