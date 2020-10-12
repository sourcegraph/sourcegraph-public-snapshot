package graphqlbackend

import "github.com/sourcegraph/sourcegraph/internal/markdown"

type MarkdownResolver interface {
	Text() string
	HTML() string
}

type Markdown string

var _ MarkdownResolver = Markdown("")

func NewMarkdownResolver(text string) MarkdownResolver {
	return Markdown(text)
}

func (m Markdown) Text() string {
	return string(m)
}

func (m Markdown) HTML() string {
	return markdown.Render(string(m))
}
