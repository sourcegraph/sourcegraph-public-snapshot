package graphqlbackend

// A resolver for the GraphQL type GenericSearchMatch
type searchResultMatchResolver struct {
	url        string
	body       string
	highlights []*rangeResolver
}

func (m *searchResultMatchResolver) URL() string {
	return m.url
}

func (m *searchResultMatchResolver) Body() *markdownResolver {
	return &markdownResolver{text: m.body}
}

func (m *searchResultMatchResolver) Highlights() []*rangeResolver {
	return m.highlights
}
