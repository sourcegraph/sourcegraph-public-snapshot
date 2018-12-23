package graphqlbackend

// A resolver for the GraphQL type GenericSearchMatch
type searchResultMatchResolver struct {
	url        string
	preview    string
	highlights []*rangeResolver
}

func (m *searchResultMatchResolver) URL() string {
	return m.url
}

func (m *searchResultMatchResolver) Preview() *markdownResolver {
	return &markdownResolver{text: m.preview}
}

func (m *searchResultMatchResolver) Highlights() []*rangeResolver {
	return m.highlights
}
