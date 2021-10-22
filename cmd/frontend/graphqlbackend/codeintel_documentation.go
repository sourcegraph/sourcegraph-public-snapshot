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

type DocumentationSearchArgs struct {
	Query string
	Repos *[]string
}

type DocumentationSearchResultsResolver interface {
	Results() []DocumentationSearchResultResolver
}

type DocumentationSearchResultResolver interface {
	Lang() string
	RepoName() string
	SearchKey() string
	PathID() string
	Label() string
	Detail() string
	Tags() []string
}
