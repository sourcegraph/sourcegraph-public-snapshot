package html

import (
	"html"
	"io"

	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
)

type Config struct {
	CSSWriter io.Writer
}

type Option func(*Config)

// WithCSSWriter registers a writer for CSS stylesheet.
func WithCSSWriter(w io.Writer) Option {
	return func(c *Config) {
		c.CSSWriter = w
	}
}

// Renderer renders the notebook as HTML.
// It supports "markdown", "code", and "raw" cells with different mime-types of the their data.
type Renderer struct {
	render.CellWrapper
	cfg Config
}

// NewRenderer configures a new HTML renderer and embeds a *Wrapper to implement render.CellWrapper.
func NewRenderer(opts ...Option) *Renderer {
	var cfg Config
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Renderer{
		CellWrapper: &Wrapper{
			Config: cfg,
		},
		cfg: cfg,
	}
}

func (r *Renderer) RegisterFuncs(reg render.RenderCellFuncRegistry) {
	// r.renderMarkdown should provide exact MimeType to override "text/*".
	reg.Register(render.Pref{Type: schema.Markdown, MimeType: common.MarkdownText}, r.renderMarkdown)
	reg.Register(render.Pref{Type: schema.Code}, r.renderCode)

	// Stream (stdout+stderr) and "error" outputs.
	reg.Register(render.Pref{Type: schema.Stream}, r.renderRaw)
	reg.Register(render.Pref{MimeType: common.Stderr}, r.renderRaw) // renders both "error" output and "stderr" stream

	// Various types of raw cell contents and display_data/execute_result outputs.
	reg.Register(render.Pref{MimeType: "application/json"}, r.renderRaw)
	reg.Register(render.Pref{MimeType: "text/*"}, r.renderRaw)
	reg.Register(render.Pref{MimeType: "text/html"}, r.renderRawHTML)
	reg.Register(render.Pref{MimeType: "image/*"}, r.renderImage)
}

// renderMarkdown renders markdown cells as pre-formatted text.
func (r *Renderer) renderMarkdown(w io.Writer, cell schema.Cell) error {
	io.WriteString(w, "<pre>")
	w.Write(cell.Text())
	io.WriteString(w, "</pre>")
	return nil
}

// renderCode renders the code blob and the code outputs.
func (r *Renderer) renderCode(w io.Writer, cell schema.Cell) error {
	code, ok := cell.(schema.CodeCell)
	if !ok {
		io.WriteString(w, "<pre><code>")
		w.Write(cell.Text())
		io.WriteString(w, "</code></pre>")
		return nil
	}

	io.WriteString(w, "<pre><code class=\"language-") // TODO: not sure if that's useful here
	io.WriteString(w, code.Language())
	io.WriteString(w, "\">")
	w.Write(code.Text())
	io.WriteString(w, "</code></pre>")

	return nil
}

// renderRawHTML writers raw contents of the cell directly to the document.
func (r *Renderer) renderRawHTML(w io.Writer, cell schema.Cell) error {
	w.Write(cell.Text())
	return nil
}

// renderImage writes base64-encoded image data.
func (r *Renderer) renderImage(w io.Writer, cell schema.Cell) error {
	io.WriteString(w, "<img src=\"data:")
	io.WriteString(w, string(cell.MimeType()))
	io.WriteString(w, ";base64, ")
	w.Write(cell.Text())
	io.WriteString(w, "\" />\n")
	return nil
}

// renderRaw writes raw contents of the cell in a new container.
func (r *Renderer) renderRaw(w io.Writer, cell schema.Cell) error {
	io.WriteString(w, "<pre>")
	txt := cell.Text()
	// Escape, because raw text may contain special HTML characters.
	escaped := html.EscapeString(string(txt[:]))
	w.Write([]byte(escaped))
	io.WriteString(w, "</pre>")
	return nil
}
