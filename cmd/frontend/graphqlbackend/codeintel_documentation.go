package graphqlbackend

// TODO(slimsag): apidocs: define in schema.graphql

type LSIFDocumentationPageArgs struct {
	PathID string
}

type MarkupContentResolver interface {
	PlainText() *string
	Markdown() *string
}

type DocumentationNodeChildResolver interface {
	Node() DocumentationNodeResolver
	PathID() string
}

type DocumentationNodeResolver interface {
	PathID() string
	Slug() string
	NewPage() bool
	Tags() []string
	Label() MarkupContentResolver
	Detail() MarkupContentResolver
	Children() []DocumentationPageResolver
}

type DocumentationPageResolver interface {
	Tree() DocumentationNodeResolver
}
