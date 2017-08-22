package ui2

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/sourcegraph/gosyntect"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	syntectServer = env.Get("SRC_SYNTECT_SERVER", "", "syntect_server HTTP(s) address")
	client        *gosyntect.Client
)

func init() {
	client = gosyntect.New(syntectServer)
}

// highlight highlights the given code with the given file extension (no
// leading ".") and returns the properly escaped HTML table representing
// the highlighted code.
func highlight(code, extension string) (template.HTML, error) {
	resp, err := client.Highlight(&gosyntect.Query{
		Code:      code,
		Extension: extension,
		Theme:     "Visual Studio Dark", // In the future, we could let the user choose the theme.
	})
	if err != nil {
		return "", err
	}
	// Note: resp.Data is properly HTML escaped by syntect_server
	table, err := preSpansToTable(resp.Data)
	if err != nil {
		return "", err
	}
	return template.HTML(table), nil
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
// 		<td>1</td>
// 		<td><span style="color:#foobar">thecode.line1</span></td>
// 	</tr>
// 	<tr>
// 		<td>2</td>
// 		<td><span style="color:#foobar">thecode.line2</span></td>
// 	</tr>
// 	</table>
//
func preSpansToTable(h string) (string, error) {
	doc, err := html.Parse(bytes.NewReader([]byte(h)))
	if err != nil {
		return "", err
	}

	body := doc.FirstChild.LastChild // html->body
	pre := body.FirstChild
	if pre == nil || pre.Type != html.ElementNode || pre.DataAtom != atom.Pre {
		return "", fmt.Errorf("exupected html->body->pre, found %+v", pre)
	}
	span := pre.FirstChild
	if span == nil || span.Type != html.ElementNode || span.DataAtom != atom.Span {
		return "", fmt.Errorf("exupected html->body->pre->span, found %+v", span)
	}

	// We will walk over all of the <span> elements and add them to an existing
	// code cell td, creating a new code cell td each time a newline is
	// encountered.
	var (
		table    = &html.Node{Type: html.ElementNode, DataAtom: atom.Table, Data: atom.Table.String()}
		next     = span // span or TextNode
		rows     int
		codeCell *html.Node
	)
	newRow := func() {
		rows++
		tr := &html.Node{Type: html.ElementNode, DataAtom: atom.Tr, Data: atom.Tr.String()}
		table.AppendChild(tr)

		tdLineNumber := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
		tr.AppendChild(tdLineNumber)

		lineNumber := &html.Node{Type: html.TextNode, Data: fmt.Sprint(rows)}
		tdLineNumber.AppendChild(lineNumber)

		codeCell = &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
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
		case next.Type == html.TextNode:
			// Text node, create a new table row for each newline.
			newlines := strings.Count(next.Data, "\n")
			for i := 0; i < newlines; i++ {
				newRow()
			}
		default:
			return "", fmt.Errorf("unexpected HTML structre (encountered %+v)", next)
		}
		next = nextSibling
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, table); err != nil {
		return "", err
	}
	return buf.String(), nil
}
