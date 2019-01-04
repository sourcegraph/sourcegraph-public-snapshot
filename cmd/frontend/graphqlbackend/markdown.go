package graphqlbackend

import "github.com/sourcegraph/sourcegraph/pkg/markdown"

type markdownResolver struct {
	text string
}

func (m *markdownResolver) Text() string {
	return m.text
}

func (m *markdownResolver) HTML() string {
	html, err := markdown.Render(m.text, nil)
	if err != nil {
		return ""
	}
	return html
}
