package graphqlbackend

import "github.com/sourcegraph/sourcegraph/pkg/markdown"

// A resolver for the GraphQL type GenericSearchMatch
type genericSearchMatchResolver struct {
	url        string
	body       string
	highlights []*highlightedRange
}

func (m *genericSearchMatchResolver) URL() string {
	return m.url
}

func (m *genericSearchMatchResolver) Body() (*markdownResolver, error) {
	var md = markdownResolver{text: m.body}
	html, err := markdown.Render(m.body, nil)
	if err != nil {
		return &md, err
	}
	md.html = &html
	return &md, nil
}

func (m *genericSearchMatchResolver) Highlights() []*highlightedRange {
	return m.highlights
}
