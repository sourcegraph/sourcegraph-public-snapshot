package graphqlbackend

type LSIFDocumentationPageArgs struct {
	PathID string
}

type DocumentationPageResolver interface {
	Tree() JSONValue
}

type LSIFDocumentationPathInfoArgs struct {
	PathID      string
	MaxDepth    *int32
	IgnoreIndex *bool
}
