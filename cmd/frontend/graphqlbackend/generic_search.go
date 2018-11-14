package graphqlbackend

import (
	"context"
	"html/template"

	"github.com/sourcegraph/sourcegraph/pkg/highlight"
	"github.com/sourcegraph/sourcegraph/pkg/markdown"
)

type genericSearchResultResolver struct {
	icon    string
	label   string
	url     string
	detail  *string
	results []*genericSearchMatchResolver
}

func (g *genericSearchResultResolver) Icon() string {
	return g.icon
}
func (g *genericSearchResultResolver) Label() string {
	label, err := markdown.Render(g.label, nil)
	if err != nil {
		return ""
	}

	return label
}

func (g *genericSearchResultResolver) URL() string {
	return g.url
}

func (g *genericSearchResultResolver) Detail() *string {
	return g.detail
}

type genericSearchMatchResolver struct {
	url        string
	body       string
	language   string
	highlights []*highlightedRange
}

func (g *genericSearchResultResolver) Results() []*genericSearchMatchResolver {
	return g.results
}

func (m *genericSearchMatchResolver) URL() string {
	return m.url
}

func (m *genericSearchMatchResolver) Body(ctx context.Context) string {
	return m.Highlight(ctx)
}

func (m *genericSearchMatchResolver) Highlights(ctx context.Context) []*highlightedRange {
	return m.highlights
}

func (m *genericSearchMatchResolver) MarkdownRenderedBody(ctx context.Context) string {
	body, err := markdown.Render(m.body, nil)
	if err != nil {
		return ""
	}
	return body
}

func (m *genericSearchMatchResolver) Language(ctx context.Context) string {
	return m.language
}

func (m *genericSearchMatchResolver) Highlight(ctx context.Context) string {
	if len(m.language) == 0 {
		return m.body
	}
	var (
		html   template.HTML
		result = &highlightedFileResolver{}
	)
	// Check if the body should be highlighted.
	var err error

	dummyFile := "file.txt"
	if len(m.language) > 0 {
		dummyFile = "file." + m.language
	}
	html, result.aborted, err = highlight.Code(ctx, m.body, dummyFile, true, true)
	if err != nil {
		return ""
	}
	result.html = string(html)
	return result.html
}
