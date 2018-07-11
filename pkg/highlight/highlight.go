package highlight

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path"
	"strings"
	"time"

	"github.com/sourcegraph/gosyntect"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	syntectServer = env.Get("SRC_SYNTECT_SERVER", "http://syntect-server:9238", "syntect_server HTTP(s) address")
	client        *gosyntect.Client
)

func init() {
	client = gosyntect.New(syntectServer)
}

// Code highlights the given code with the given filepath (must contain at
// least the file name + extension) and returns the properly escaped HTML table
// representing the highlighted code.
//
// The returned boolean represents whether or not highlighting was aborted due
// to timeout. In this scenario, a plain text table is returned.
func Code(ctx context.Context, code, filepath string, disableTimeout bool, isLightTheme bool) (template.HTML, bool, error) {
	if !disableTimeout {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
	}
	themechoice := "Sourcegraph"
	if isLightTheme {
		themechoice = "Solarized (light)"
	}

	extension := strings.TrimPrefix(path.Ext(filepath), ".")
	if extension == "" {
		// When the file does not have an extension, fall back to the basename
		// (e.g. for Dockerfile).
		extension = path.Base(filepath)
	}

	// Trim a single newline from the end of the file. This means that a file
	// "a\n\n\n\n" will show line numbers 1-4 rather than 1-5, i.e. no blank
	// line will be shown at the end of the file corresponding to the last
	// newline.
	//
	// This matches other online code reading tools such as e.g. GitHub; see
	// https://github.com/sourcegraph/sourcegraph/issues/8024 for more
	// background.
	code = strings.TrimSuffix(code, "\n")

	resp, err := client.Highlight(ctx, &gosyntect.Query{
		// Legacy: Passing this in here so that new Sourcegraph versions work
		// with old syntect_server versions. We will remove this in the future.
		Extension: extension,

		Code:     code,
		Filepath: filepath,
		Theme:    themechoice,
	})

	if ctx.Err() == context.DeadlineExceeded {
		// Timeout, so render plain table.
		table, err2 := generatePlainTable(code)
		return table, true, err2
	} else if err != nil {
		postTooLarge := strings.HasSuffix(err.Error(), "EOF")
		// TODO(slimsag): gosyntect should provide concrete error types here - we have invalid extension here for backward compatibility.
		// can remove in ~1 month.
		invalidSyntax := strings.Contains(err.Error(), "invalid extension") || strings.Contains(err.Error(), "invalid syntax")
		if invalidSyntax || postTooLarge {
			// Failed to highlight code, e.g. for a text file. We still need to
			// generate the table.
			table, err2 := generatePlainTable(code)
			return table, false, err2
		}
		return "", false, err
	}
	// Note: resp.Data is properly HTML escaped by syntect_server
	table, err := preSpansToTable(resp.Data)
	if err != nil {
		return "", false, err
	}
	return template.HTML(table), false, nil
}

// preSpansToTable takes the syntect data structure, which looks like:
//
// 	<pre>
// 	<span style="color:#foobar">thecode.line1</span>
// 	<span style="color:#foobar">thecode.line2</span>
// 	</pre>
//
// And turns it into a table in the format which the frontend expects:
//
// 	<table>
// 	<tr>
// 		<td class="line" data-line="1"></td>
// 		<td class="code"><span style="color:#foobar">thecode.line1</span></td>
// 	</tr>
// 	<tr>
// 		<td class="line" data-line="2"></td>
// 		<td class="code"><span style="color:#foobar">thecode.line2</span></td>
// 	</tr>
// 	</table>
//
func preSpansToTable(h string) (string, error) {
	// TODO(slimsag): remove conversion once we switch to blob2 frontend component.
	converted, err := convertNewlinesToNoNewlines([]byte(h))
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(bytes.NewReader(converted))
	if err != nil {
		return "", err
	}

	body := doc.FirstChild.LastChild // html->body
	pre := body.FirstChild
	if pre == nil || pre.Type != html.ElementNode || pre.DataAtom != atom.Pre {
		return "", fmt.Errorf("expected html->body->pre, found %+v", pre)
	}

	// We will walk over all of the <span> elements and add them to an existing
	// code cell td, creating a new code cell td each time a newline is
	// encountered.
	var (
		table    = &html.Node{Type: html.ElementNode, DataAtom: atom.Table, Data: atom.Table.String()}
		next     = pre.FirstChild // span or TextNode
		rows     int
		codeCell *html.Node
	)
	newRow := func() {
		// If the previous row did not have any children, then it was a blank
		// line. Blank lines always need a span with a newline character for
		// proper whitespace copy+paste support.
		if codeCell != nil && codeCell.FirstChild == nil {
			span := &html.Node{Type: html.ElementNode, DataAtom: atom.Span, Data: atom.Span.String()}
			codeCell.AppendChild(span)
			spanText := &html.Node{Type: html.TextNode, Data: "\n"}
			span.AppendChild(spanText)
		}

		rows++
		tr := &html.Node{Type: html.ElementNode, DataAtom: atom.Tr, Data: atom.Tr.String()}
		table.AppendChild(tr)

		tdLineNumber := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
		tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "class", Val: "line"})
		tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "data-line", Val: fmt.Sprint(rows)})
		tr.AppendChild(tdLineNumber)

		codeCell = &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
		codeCell.Attr = append(codeCell.Attr, html.Attribute{Key: "class", Val: "code"})
		tr.AppendChild(codeCell)
	}
	newRow()
	for next != nil {
		nextSibling := next.NextSibling
		switch {
		case next.Type == html.ElementNode && next.DataAtom == atom.Span:
			// Found a span, so add it to our current code cell td.
			next.Parent = nil
			next.PrevSibling = nil
			next.NextSibling = nil
			codeCell.AppendChild(next)

			// Scan the children for text nodes containing new lines so that we
			// can create new table rows.
			if next.FirstChild != nil {
				nextChild := next.FirstChild
				for nextChild != nil {
					switch {
					case nextChild.Type == html.TextNode:
						// Text node, create a new table row for each newline.
						newlines := strings.Count(nextChild.Data, "\n")
						for i := 0; i < newlines; i++ {
							newRow()
						}
					default:
						return "", fmt.Errorf("unexpected HTML child structure (encountered %+v)", nextChild)
					}
					nextChild = nextChild.NextSibling
				}
			}
		case next.Type == html.TextNode:
			// TODO(slimsag): Remove this case in the near future. For now, it
			// is kept in case someone tries to run a new Sourcegraph version
			// with an old syntect_server version.

			// Text node, create a new table row for each newline.
			newlines := strings.Count(next.Data, "\n")
			for i := 0; i < newlines; i++ {
				newRow()
			}
		default:
			return "", fmt.Errorf("unexpected HTML structure (encountered %+v)", next)
		}
		next = nextSibling
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, table); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func generatePlainTable(code string) (template.HTML, error) {
	table := &html.Node{Type: html.ElementNode, DataAtom: atom.Table, Data: atom.Table.String()}
	for row, line := range strings.Split(code, "\n") {
		line = strings.TrimSuffix(line, "\r") // CRLF files
		if line == "" {
			line = "\n" // important for e.g. selecting whitespace in the produced table
		}
		tr := &html.Node{Type: html.ElementNode, DataAtom: atom.Tr, Data: atom.Tr.String()}
		table.AppendChild(tr)

		tdLineNumber := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
		tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "class", Val: "line"})
		tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "data-line", Val: fmt.Sprint(row + 1)})
		tr.AppendChild(tdLineNumber)

		codeCell := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
		codeCell.Attr = append(codeCell.Attr, html.Attribute{Key: "class", Val: "code"})
		tr.AppendChild(codeCell)

		// Span to match same structure as what highlighting would usually generate.
		span := &html.Node{Type: html.ElementNode, DataAtom: atom.Span, Data: atom.Span.String()}
		codeCell.AppendChild(span)
		spanText := &html.Node{Type: html.TextNode, Data: line}
		span.AppendChild(spanText)
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, table); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// The latest version of syntect_server uses Syntect's "newlines" mode, which
// is more feature-complete than the previously used "no newlines" mode.
// However, this produces an HTML structure where newlines are in the last span
// element for a line:
//
// 	<pre>
// 	<span>line</span><span> one
// 	</span><span>
// 	</span><span>line</span><span> two
// 	</span></pre>
//
// Rather than where they previously were, as a text node _after_ the last span
// element for a line:
//
// 	<pre><span>line</span><span> one</span>
// 	<span></span>
// 	<span>line</span><span> two</span>
// 	</pre>
//
// This function translates the new HTML structure format to the old one, so
// that it works seamlessly with the old "blob" frontend component.
//
// The new "blob2" frontend component handles this seamlessly, and this
// conversion code can be removed in the near future once "blob2" is no longer
// feature-flagged. However, for now, as it is uncertain when the new "blob2"
// frontend component will replace the old one, so using this translation code
// is neccessary for now.
//
// See also https://github.com/sourcegraph/syntect_server/commit/1ff36cf3a77df80a7559e139b386c043438190ca
func convertNewlinesToNoNewlines(h []byte) ([]byte, error) {
	doc, err := html.Parse(bytes.NewReader([]byte(h)))
	if err != nil {
		return nil, err
	}

	body := doc.FirstChild.LastChild // html->body
	oldPre := body.FirstChild
	if oldPre == nil || oldPre.Type != html.ElementNode || oldPre.DataAtom != atom.Pre {
		return nil, fmt.Errorf("expected html->body->pre, found %+v", oldPre)
	}

	// We will walk over all of the <span> elements in the <pre> element and
	// add them to a new <pre> element, shifting the line endings outside.
	var (
		pre = &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Pre,
			Data:     oldPre.Data,
			Attr:     oldPre.Attr,
		}
		next = oldPre.FirstChild // span or TextNode
	)

	newRow := func() {
		/*
			// If the previous row did not have any children, then it was a blank
			// line. Blank lines always need a span with a newline character for
			// proper whitespace copy+paste support.
			if codeCell != nil && codeCell.FirstChild == nil {
				span := &html.Node{Type: html.ElementNode, DataAtom: atom.Span, Data: atom.Span.String()}
				codeCell.AppendChild(span)
				spanText := &html.Node{Type: html.TextNode, Data: "\n"}
				span.AppendChild(spanText)
			}

			rows++
			tr := &html.Node{Type: html.ElementNode, DataAtom: atom.Tr, Data: atom.Tr.String()}
			table.AppendChild(tr)

			tdLineNumber := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
			tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "class", Val: "line"})
			tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "data-line", Val: fmt.Sprint(rows)})
			tr.AppendChild(tdLineNumber)

			codeCell = &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
			codeCell.Attr = append(codeCell.Attr, html.Attribute{Key: "class", Val: "code"})
			tr.AppendChild(codeCell)
		*/
	}
	newRow()
	for next != nil {
		nextSibling := next.NextSibling
		switch {
		case next.Type == html.ElementNode && next.DataAtom == atom.Span:
			// Found a span, so add it to our new pre element.
			next.Parent = nil
			next.PrevSibling = nil
			next.NextSibling = nil
			pre.AppendChild(next)

			// Scan the span for text nodes containing new lines.
			if next.FirstChild != nil {
				nextChild := next.FirstChild
				for nextChild != nil {
					switch {
					case nextChild.Type == html.TextNode:
						// We found a text node, if it contains newlines then
						// trim it from the span's text node and insert a
						// newline textnode in the new pre element.
						if strings.Contains(nextChild.Data, "\n") {
							nextChild.Data = strings.TrimSuffix(nextChild.Data, "\n")
							nextChild.Data = strings.TrimSuffix(nextChild.Data, "\r")
							pre.AppendChild(&html.Node{Type: html.TextNode, Data: "\n"})
						}
					default:
						return nil, fmt.Errorf("unexpected HTML child structure (encountered %+v)", nextChild)
					}
					nextChild = nextChild.NextSibling
				}
			}

		case next.Type == html.TextNode:
			next.Parent = nil
			next.PrevSibling = nil
			next.NextSibling = nil
			pre.AppendChild(next)

		default:
			return nil, fmt.Errorf("unexpected HTML structure (encountered %+v)", next)
		}
		next = nextSibling
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, pre); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
