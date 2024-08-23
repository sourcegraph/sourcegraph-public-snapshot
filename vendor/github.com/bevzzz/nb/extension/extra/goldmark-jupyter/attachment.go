// Package jupyter provides extensions for goldmark and nb. Together they add support
// for inline images, which have their data stored as cell attachments, in markdown cells.
//
// How it is achieved:
//
//  1. Goldmark extends nb with a custom "markdown" cell renderer which
//     stores cell attachments to the parser.Context on every render.
//
//  2. Attachments extends goldmark with a custom link parser (ast.KindLink)
//     and an image NodeRenderFunc.
//
//     The parser is context-aware and will get the related mime-bundle from the context
//     and store it to node attributes for every link whose destination looks like "attachments:image.png"
//
//     Custom image renderer writes base64-encoded data from the mime-bundle if one's present,
//     falling back to the destination URL.
package jupyter

import (
	"io"
	"regexp"

	"github.com/bevzzz/nb"
	"github.com/bevzzz/nb/extension"
	"github.com/bevzzz/nb/schema"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Attachments adds support for Jupyter [cell attachments] to goldmark parser and renderer.
//
// [cell attachments]: https://nbformat.readthedocs.io/en/latest/format_description.html#cell-attachments
func Attachments(opts ...html.Option) goldmark.Extender {
	c := html.NewConfig()
	for _, opt := range opts {
		opt.SetHTMLOption(&c)
	}
	return &attachments{
		config: c,
	}
}

// Goldmark overrides the default rendering function for markdown cells
// and stores cell attachments to the parser.Context on every render.
func Goldmark(md goldmark.Markdown) nb.Extension {
	return extension.NewMarkdown(
		func(w io.Writer, c schema.Cell) error {
			ctx := newContext(c)
			return md.Convert(c.Text(), w, parser.WithContext(ctx))
		},
	)
}

var (
	// key is a context key for storing cell attachments.
	key = parser.NewContextKey()

	// name is the name of a node attribute that holds the mime-bundle.
	// This package uses node attributes as a proxy for rendering context, 
	// so <mime-bundle> will never be added to the HTML output. The name is
	// intentionally [invalid] to avoid name-clashes with othen potential attributes.
	//
	// [invalid]: https://www.w3.org/TR/2011/WD-html5-20110525/syntax.html#attributes-0
	name = []byte("<mime-bundle>")
)

// newContext adds mime-bundles from cell attachements to a new parse.Context.
func newContext(cell schema.Cell) parser.Context {
	ctx := parser.NewContext()
	if c, ok := cell.(schema.HasAttachments); ok {
		ctx.Set(key, c.Attachments())
	}
	return ctx
}

// linkParser adds base64-encoded image data from parser.Context to node's attributes.
type linkParser struct {
	link parser.InlineParser // link is goldmark's default link parser.
}

func newLinkParser() *linkParser {
	return &linkParser{
		link: parser.NewLinkParser(),
	}
}

var _ parser.InlineParser = (*linkParser)(nil)

func (p *linkParser) Trigger() []byte {
	return p.link.Trigger()
}

// attachedFile retrieves the name of the attached file from the link's destination.
var attachedFile = regexp.MustCompile(`attachment:(\w+\.\w+)$`)

// Parse stores mime-bundle in node attributes for links whose destination is an attachment.
func (p *linkParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) (n ast.Node) {
	n = p.link.Parse(parent, block, pc)

	img, ok := n.(*ast.Image)
	if !ok {
		// goldmark's default link parser will return a "state node" whenever it's triggered
		// by the opening bracket of the link's alt-text "[" or any intermediate characters.
		// We only want to intercept when the link is done parsing and we get a valid *ast.Image.
		return n
	}

	submatch := attachedFile.FindSubmatch(img.Destination)
	if len(submatch) < 2 {
		return
	}
	filename := submatch[1]

	att, ok := pc.Get(key).(schema.Attachments)
	if att == nil || !ok {
		return
	}

	// Admittedly
	data := att.MimeBundle(string(filename))
	n.SetAttribute(name, data)
	return
}

// image renders inline images from cell attachments.
type image struct {
	html.Config
}

var _ renderer.NodeRenderer = (*image)(nil)

func (img *image) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, img.render)
}

// render borrows heavily from goldmark's [renderImage].
//
// [renderImage]: https://github.com/yuin/goldmark/blob/90c46e0829c11ca8d1010856b2a6f6f88bfc68a3/renderer/html/html.go#L673
func (img *image) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.Image)
	_, _ = w.WriteString("<img src=\"")

	attr, hasAttachments := n.Attribute(name)
	if !hasAttachments {
		if img.Unsafe || !html.IsDangerousURL(n.Destination) {
			_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		}
	} else if mb, ok := attr.(schema.MimeBundle); ok {
		// Here we do not need to extract the filename again, as it is sufficient
		// that the mime-bundle is present in the attributes.
		io.WriteString(w, "data:")
		io.WriteString(w, mb.MimeType())
		io.WriteString(w, ";base64, ")
		w.Write(mb.Text())
	}

	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(nodeToHTMLText(n, source))
	_ = w.WriteByte('"')

	if n.Title != nil {
		_, _ = w.WriteString(` title="`)
		img.Writer.Write(w, n.Title)
		_ = w.WriteByte('"')
	}

	if n.Attributes() != nil {
		html.RenderAttributes(w, n, html.ImageAttributeFilter)
	}

	if img.XHTML {
		_, _ = w.WriteString(" />")
	} else {
		_, _ = w.WriteString(">")
	}

	return ast.WalkSkipChildren, nil
}

// attachments implements goldmark.Extender.
type attachments struct {
	config html.Config
}

var _ goldmark.Extender = (*attachments)(nil)

// Extends adds custom link parser and image renderer.
//
// Priorities are selected based on the ones used in goldmark.
func (a *attachments) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(
		parser.WithInlineParsers(util.Prioritized(newLinkParser(), 199)), // default: 200
	)
	md.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(&image{Config: a.config}, 999)), // default: 1000
	)
}
