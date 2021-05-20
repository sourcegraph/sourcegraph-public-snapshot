package graphqlbackend

type LSIFDocumentationPageArgs struct {
	PathID string
}

type DocumentationPageResolver interface {
	Tree() JSONValue
}
