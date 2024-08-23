package extension

import (
	"github.com/bevzzz/nb"
	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
)

// NewMarkdown overrides the default rendering function for markdown cells.
//
// While its lax signature allows passing any arbitrary RenderCellFunc,
// it will be best used to extend nb with existing markdown converters.
// Package extension/adapters offers elegant wrappers for some of the popular options:
//
//	extension.NewMarkdown(
//		adapter.Blackfriday(blackfriday.MarkdownCommon)
//	)
//
// or
//
//	extension.NewMarkdown(
//		adapter.Goldmark(func(b []byte, w io.Writer) error {
//			return goldmark.Convert(b, w)
//		})
//	)
func NewMarkdown(f render.RenderCellFunc) nb.Extension {
	return &markdown{
		render: f,
	}
}

type markdown struct {
	render render.RenderCellFunc
}

var _ nb.Extension = (*markdown)(nil)
var _ render.CellRenderer = (*markdown)(nil)

// RegisterFuncs registers a new RenderCellFunc for markdown cells.
func (md *markdown) RegisterFuncs(reg render.RenderCellFuncRegistry) {
	reg.Register(render.Pref{Type: schema.Markdown, MimeType: common.MarkdownText}, md.render)
}

// Extend adds markdown as a cell renderer.
func (md *markdown) Extend(n *nb.Notebook) {
	n.Renderer().AddOptions(render.WithCellRenderers(md))
}
