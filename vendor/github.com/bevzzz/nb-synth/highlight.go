package synth

import (
	"html"
	"io"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/bevzzz/nb"
	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
)

func NewHighlighting(opts ...Option) nb.Extension {
	return &highlighting{
		options: opts,
	}
}

var Highlighting = NewHighlighting()

type Option func(*Config)

type Config struct {
	Style           string
	CustomStyle     *chroma.Style
	GuessLanguage   bool
	Coalesce        bool
	FormatOptions   []chromahtml.Option
	TokeniseOptions *chroma.TokeniseOptions
	CSSWriter       io.Writer
}

func NewConfig() Config {
	return Config{
		Style:    "github",
		Coalesce: true,
	}
}

func WithStyle(s string) Option {
	return func(c *Config) { c.Style = s }
}

func WithCustomStyle(s *chroma.Style) Option {
	return func(c *Config) { c.CustomStyle = s }
}

func WithGuessLanguage(v bool) Option {
	return func(c *Config) { c.GuessLanguage = v }
}

func WithCoalesce(v bool) Option {
	return func(c *Config) { c.Coalesce = v }
}

func WithFormatOptions(opts ...chromahtml.Option) Option {
	return func(c *Config) {
		c.FormatOptions = append(c.FormatOptions, opts...)
	}
}

func WithTokenizeOptions(opt *chroma.TokeniseOptions) Option {
	return func(c *Config) { c.TokeniseOptions = opt }
}

func WithCSSWriter(w io.Writer) Option {
	return func(c *Config) {
		c.CSSWriter = w
	}
}

func newRenderer(opts ...Option) *renderer {
	r := renderer{
		Config: NewConfig(),
	}
	for _, opt := range opts {
		opt(&r.Config)
	}
	return &r
}

type renderer struct {
	Config
}

var _ render.CellRenderer = (*renderer)(nil)

func (r *renderer) RegisterFuncs(reg render.RenderCellFuncRegistry) {
	reg.Register(render.Pref{Type: schema.Code}, r.renderCode)
	reg.Register(render.Pref{MimeType: "application/json"}, r.renderData)
	reg.Register(render.Pref{MimeType: "text/xml"}, r.renderData)
	reg.Register(render.Pref{MimeType: "application/*xml"}, r.renderData)
}

func (r *renderer) renderCode(w io.Writer, cell schema.Cell) error {
	code := cell.(schema.CodeCell)
	codeString := string(code.Text())

	var lexer chroma.Lexer
	if lang := code.Language(); lang != "" {
		lexer = lexers.Get(lang)
	} else if mime := code.MimeType(); mime != "" {
		lexer = lexers.MatchMimeType(mime)
	}

	if lexer == nil {
		if !r.GuessLanguage {
			return r.renderRaw(w, code)
		}
		lexer = lexers.Analyse(codeString)
		if lexer == nil {
			lexer = lexers.Fallback
		}
	}

	if r.Coalesce {
		lexer = chroma.Coalesce(lexer)
	}

	it, err := lexer.Tokenise(r.TokeniseOptions, codeString)
	if err != nil {
		return r.renderRaw(w, code)
	}

	style := r.CustomStyle
	if style == nil {
		style = styles.Get(r.Style)
	}

	f := chromahtml.New(r.FormatOptions...)
	_ = f.Format(w, style, it)

	if r.CSSWriter != nil {
		_ = f.WriteCSS(w, style)
	}
	return nil
}

// renderRaw renders cell contents as preformatted text (<pre>).
func (r *renderer) renderRaw(w io.Writer, c schema.Cell) error {
	io.WriteString(w, "<pre>")
	txt := c.Text()
	// Escape, because raw text may contain special HTML characters.
	escaped := html.EscapeString(string(txt[:]))
	w.Write([]byte(escaped))
	io.WriteString(w, "</pre>")
	return nil
}

// renderData renders outputs in text-based data formats like JSON or XML.
func (r *renderer) renderData(w io.Writer, c schema.Cell) error {
	jsonString := string(c.Text())
	lexer := lexers.MatchMimeType(c.MimeType())
	if lexer == nil {
		return r.renderRaw(w, c)
	}

	if r.Coalesce {
		lexer = chroma.Coalesce(lexer)
	}

	it, err := lexer.Tokenise(r.TokeniseOptions, jsonString)
	if err != nil {
		return r.renderRaw(w, c)
	}

	style := r.CustomStyle
	if style == nil {
		style = styles.Get(r.Style)
	}

	f := chromahtml.New(r.FormatOptions...)
	_ = f.Format(w, style, it)

	if r.CSSWriter != nil {
		_ = f.WriteCSS(w, style)
	}
	return nil
}

type highlighting struct {
	options []Option
}

var _ nb.Extension = (*highlighting)(nil)

func (h *highlighting) Extend(n *nb.Notebook) {
	n.Renderer().AddOptions(render.WithCellRenderers(
		newRenderer(h.options...),
	))
}
