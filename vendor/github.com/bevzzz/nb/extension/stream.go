package extension

import (
	"github.com/bevzzz/nb"
	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
)

// NewStream overrides the default rendering function for "stream" and "error" output cells.
// These will often be formatted with ANSI-color codes, which you may want to replace with
// styled HTML tags or strip from the output completely.
//
// For example, use [ansihtml] with a dedicated adapter:
//
//	extension.NewStream(
//		adapter.AnsiHtml(ansihtml.ConvertToHTML)
//	)
//
// [ansihtml]: https://github.com/robert-nix/ansihtml
func NewStream(f render.RenderCellFunc) nb.Extension {
	return &stream{
		render: f,
	}
}

type stream struct {
	render render.RenderCellFunc
}

var _ nb.Extension = (*stream)(nil)
var _ render.CellRenderer = (*stream)(nil)

// RegisterFuncs registers a new RenderCellFunc for stream output cells.
func (s *stream) RegisterFuncs(reg render.RenderCellFuncRegistry) {
	reg.Register(render.Pref{Type: schema.Stream, MimeType: common.Stdout}, s.render)
	reg.Register(render.Pref{Type: schema.Stream, MimeType: common.Stderr}, s.render)
	reg.Register(render.Pref{Type: schema.Error, MimeType: common.Stderr}, s.render)
}

// Extend adds stream as a cell renderer.
func (s *stream) Extend(n *nb.Notebook) {
	n.Renderer().AddOptions(render.WithCellRenderers(s))
}
