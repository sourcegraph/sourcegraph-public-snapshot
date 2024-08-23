package html

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
	"golang.org/x/net/html"
)

/*
TODO:
	- make class prefixes configurable (probably on the html.Config level).
*/

// Wrapper wraps cells in the HTML produced by the original Jupyter's nbconvert.
type Wrapper struct {
	Config
}

var _ render.CellWrapper = (*Wrapper)(nil)

func (wr *Wrapper) WrapAll(w io.Writer, render func(io.Writer) error) error {
	tag := tagger{Writer: w}
	defer tag.Close()

	if wr.CSSWriter != nil {
		wr.CSSWriter.Write(jupyterCSS)
	}

	tag.Open("div", attributes{"class": {"jp-Notebook"}})
	return render(w)
}

func (wr *Wrapper) Wrap(w io.Writer, cell schema.Cell, render render.RenderCellFunc) error {
	tag := tagger{Writer: w}
	defer tag.Close()

	var ct string
	switch cell.Type() {
	case schema.Markdown:
		ct = "jp-MarkdownCell"
	case schema.Code:
		ct = "jp-CodeCell"
		// TODO: if no outputs, add jp-mod-noOutputs class
	case schema.Raw:
		ct = "jp-RawCell"
	}

	tag.Open("div", attributes{"class": {"jp-Cell", ct, "jp-Notebook-cell"}})
	render(w, cell)
	return nil
}

func (wr *Wrapper) WrapInput(w io.Writer, cell schema.Cell, render render.RenderCellFunc) error {
	tag := tagger{Writer: w}
	defer tag.Close()

	tag.Open("div", attributes{
		"class":    {"jp-Cell-inputWrapper"},
		"tabindex": {0}})

	tag.Open("div", attributes{"class": {"jp-Collapser", "jp-InputCollapser", "jp-Cell-inputCollapser"}})
	io.WriteString(w, " ")
	tag.CloseLast()

	// TODO: add collapser-child <div class="jp-Collapser-child"></div> and collapsing functionality
	// Pure CSS Collapsible: https://www.digitalocean.com/community/tutorials/css-collapsible

	tag.Open("div", attributes{"class": {"jp-InputArea", "jp-Cell-inputArea"}})

	// Prompt In:[1]
	tag.OpenInline("div", attributes{"class": {"jp-InputPrompt", "jp-InputArea-prompt"}})
	if ex, ok := cell.(interface{ ExecutionCount() int }); ok {
		fmt.Fprintf(w, "In\u00a0[%d]:", ex.ExecutionCount())
	}
	tag.CloseLast()

	switch cell.Type() {
	case schema.Code:
		tag.Open("div", attributes{
			"class": {
				"jp-CodeMirrorEditor",
				"jp-Editor",
				"jp-InputArea-editor",
			},
			"data-type": {"inline"},
		})

		tag.Open("div", attributes{"class": {"cm-editor", "cm-s-jupyter"}})
		tag.Open("div", attributes{"class": {"highlight", "hl-ipython3"}})

	case schema.Markdown:
		tag.Open("div", attributes{
			"class": {
				"jp-RenderedMarkdown",
				"jp-MarkdownOutput",
				"jp-RenderedHTMLCommon",
			},
			"data-mime-type": {common.MarkdownText},
		})
	}

	_ = render(w, cell)
	return nil
}

func (wr *Wrapper) WrapOutput(w io.Writer, cell schema.Outputter, render render.RenderCellFunc) error {
	tag := tagger{Writer: w}
	defer tag.Close()

	tag.Open("div", attributes{"class": {"jp-Cell-outputWrapper"}})
	tag.OpenInline("div", attributes{"class": {"jp-Collapser", "jp-OutputCollapser", "jp-Cell-outputCollapser"}})
	tag.CloseLast()
	tag.Open("div", attributes{"class": {"jp-OutputArea jp-Cell-outputArea"}})

	// TODO: jp-RenderedJavaScript is a thing and so is jp-RenderedLatex (but I don't think we need to do anything about the latter)

	var child bool
	var childClass = "jp-OutputArea-child"
	var datamimetype string
	var outputtypeclass string

	if outs := cell.Outputs(); len(outs) > 0 {
		datamimetype = outs[0].MimeType()
		first := outs[0]

		switch first.Type() {
		case schema.ExecuteResult:
			outputtypeclass = "jp-OutputArea-executeResult"
			child = true
		case schema.Error:
			child = true
		case schema.Stream:
			datamimetype = common.PlainText
			child = true
		}
	}

	var renderedClass string
	if strings.HasPrefix(datamimetype, "text/") || datamimetype == "application/json" {
		childClass += " jp-OutputArea-executeResult"
		renderedClass = "jp-RenderedText"
		if datamimetype == "text/html" {
			renderedClass = "jp-RenderedHTMLCommon jp-RenderedHTML"
		}
	} else if strings.HasPrefix(datamimetype, "image/") {
		renderedClass = "jp-RenderedImage"
		child = true
	} else if datamimetype == common.Stderr {
		renderedClass = "jp-RenderedText"
	}

	if child {
		tag.Open("div", attributes{"class": {childClass}})
	}

	tag.OpenInline("div", attributes{"class": {"jp-OutputPrompt", "jp-OutputArea-prompt"}})
	for _, out := range cell.Outputs() {
		if ex, ok := out.(interface{ ExecutionCount() int }); ok {
			fmt.Fprintf(w, "Out\u00a0[%d]:", ex.ExecutionCount())
			break
		}
	}
	tag.CloseLast()

	tag.Open("div", attributes{
		"class":          {renderedClass, "jp-OutputArea-output", outputtypeclass},
		"data-mime-type": {datamimetype},
	})
	for _, out := range cell.Outputs() {
		_ = render(w, out)
	}
	tag.CloseLast()
	return nil
}

// tagger is a straightforward utility for writing HTML tags.
//
// Example:
//
//	tag := tagger{Writer: os.Stdout}
//	defer tag.Close()
//	tag.Open("div", attributes{"class": {"box"}})
//	tag.Open("pre", attributes{"class": {"hl", "python"}})
//
// tagger also supports empty tags.
type tagger struct {
	Writer io.Writer
	opened []string
}

// Open opens the tag with the attributes.
func (t *tagger) Open(tag string, attr attributes) {
	t.openTag(tag, attr, true)
}

// Open inline does not add a newline '\n' after the opening tag.
func (t *tagger) OpenInline(tag string, attr attributes) {
	t.openTag(tag, attr, false)
}

// Empty creates an empty HTML tag, like <input />.
func (t *tagger) Empty(tag string, attr attributes) {
	io.WriteString(t.Writer, "<")
	io.WriteString(t.Writer, tag)
	attr.WriteTo(t.Writer)
	io.WriteString(t.Writer, " />")
}

func (t *tagger) openTag(tag string, attr attributes, newline bool) {
	io.WriteString(t.Writer, "<")
	io.WriteString(t.Writer, tag)
	attr.WriteTo(t.Writer)
	io.WriteString(t.Writer, ">")
	if newline {
		io.WriteString(t.Writer, "\n")
	}
	t.opened = append(t.opened, tag)
}

// Close closes all opened tags in reverse order.
// Always adds a newline after the tag.
func (t *tagger) Close() {
	l := len(t.opened)
	if l == 0 {
		return
	}
	for i := l - 1; i >= 0; i-- {
		t.closeTag(t.opened[i])
	}
}

func (t *tagger) CloseLast() {
	l := len(t.opened)
	if l == 0 {
		return
	}
	t.closeTag(t.opened[l-1])
	t.opened = t.opened[:l-1]
}

func (t *tagger) closeTag(tag string) {
	fmt.Fprintf(t.Writer, "</%s>\n", tag)
}

type attributes map[string][]interface{}

// WriteTo writes values of each attribute in a space-separated list, e.g. class="container box jp-NotebookCell".
// TODO: refactor to use html.Attribute from the beginning
func (attrs attributes) WriteTo(w io.Writer) (n64 int64, err error) {
	type Attribute struct {
		html.Attribute
		IsBool bool
	}

	var sorted []Attribute
	for k, values := range attrs {
		var v string
		attr := Attribute{
			Attribute: html.Attribute{Key: k},
		}

		if len(values) == 1 {
			if _, isBool := values[0].(bool); isBool {
				attr.IsBool = true
				sorted = append(sorted, attr)
				continue
			}
		}

		for i := range values {
			if i > 0 {
				v += " "
			}
			v += fmt.Sprintf("%v", values[i])
		}

		attr.Val = v
		sorted = append(sorted, attr)
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})

	for _, attr := range sorted {
		s := " "
		if attr.IsBool {
			// FIXME: 'continue' prevents the attribute from being written.
			s += attr.Key
			continue
		}
		s += fmt.Sprintf("%s=\"%s\"", attr.Key, attr.Val)

		var n int
		n, err = io.WriteString(w, s)
		if err != nil {
			return
		}
		n64 += int64(n)
	}
	return
}
