package graphqlbackend

// A resolver for the GraphQL type GenericSearchMatch
type genericSearchMatchResolver struct {
	url        string
	body       string
	highlights []*highlightedRange
}

func (m *genericSearchMatchResolver) URL() string {
	return m.url
}

func (m *genericSearchMatchResolver) Body() *markdownResolver {
	return &markdownResolver{text: m.body}
}

func (m *genericSearchMatchResolver) Highlights() []*highlightedRange {
	return m.highlights
}
