package graphqlbackend

import "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/markdown"

type markdownResolver struct {
	text string
}

func (m *markdownResolver) Text() string {
	return m.text
}

func (m *markdownResolver) HTML() string {
	return markdown.Render(m.text, nil)
}
