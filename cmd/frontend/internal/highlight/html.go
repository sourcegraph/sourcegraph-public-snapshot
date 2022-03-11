package highlight

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type lsifFormatter interface {
	AddRow(row int32)
	AddText(kind lsiftyped.SyntaxKind, line string)
}

func lsifToHTML(
	code string,
	document *lsiftyped.Document,
	AddRow func(row int32),
	AddText func(kind lsiftyped.SyntaxKind, line string),
	validLines map[int32]bool,
) {
	splitLines := strings.Split(code, "\n")
	occurences := document.Occurrences

	row, occIndex := int32(0), 0
	for row < int32(len(splitLines)) {
		// skip invalid lines, when passed
		if validLines != nil && !validLines[row] {
			row += 1
			continue
		}

		line := strings.TrimSuffix(splitLines[row], "\r")
		if line == "" {
			line = "\n" // important for e.g. selecting whitespace in the produced table
		}

		AddRow(row)

		lineCharacter := 0
		for occIndex < len(occurences) && occurences[occIndex].Range[0] < row+1 {
			occ := occurences[occIndex]
			occIndex += 1

			startRow, startCharacter, endRow, endCharacter := getRange(occ.Range)
			AddText(occ.SyntaxKind, line[lineCharacter:startCharacter])

			if startRow != endRow {
				AddText(occ.SyntaxKind, line[startCharacter:])

				row += 1
				for row < endRow {
					line = splitLines[row]

					AddRow(row)
					AddText(occ.SyntaxKind, line)

					row += 1
				}

				line = splitLines[row]
				AddRow(row)
				AddText(occ.SyntaxKind, line[:endCharacter])
			} else {
				AddText(occ.SyntaxKind, line[startCharacter:endCharacter])
			}

			lineCharacter = int(endCharacter)
		}

		AddText(lsiftyped.SyntaxKind_UnspecifiedSyntaxKind, line[lineCharacter:])

		row += 1
	}
}

// appendText formats the text to the right css class and appends to the current
// html node
func appendText(tr *html.Node, kind lsiftyped.SyntaxKind, text string) {
	if text == "" {
		return
	}

	var class string
	if kind != lsiftyped.SyntaxKind_UnspecifiedSyntaxKind {
		class = "hl-typed-" + lsiftyped.SyntaxKind_name[int32(kind)]
	}

	span := &html.Node{Type: html.ElementNode, DataAtom: atom.Span, Data: atom.Span.String()}
	if class != "" {
		span.Attr = append(span.Attr, html.Attribute{Key: "class", Val: class})
	}
	tr.AppendChild(span)
	spanText := &html.Node{Type: html.TextNode, Data: text}
	span.AppendChild(spanText)
}

func newRow(row int32) (htmlRow, htmlCode *html.Node) {
	tr := &html.Node{Type: html.ElementNode, DataAtom: atom.Tr, Data: atom.Tr.String()}

	tdLineNumber := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
	tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "class", Val: "line"})
	tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "data-line", Val: fmt.Sprint(row + 1)})
	tr.AppendChild(tdLineNumber)

	codeCell := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
	codeCell.Attr = append(codeCell.Attr, html.Attribute{Key: "class", Val: "code"})
	tr.AppendChild(codeCell)

	return tr, codeCell
}

func getRange(r []int32) (int32, int32, int32, int32) {
	if len(r) == 3 {
		return r[0], r[1], r[0], r[2]
	} else {
		return r[0], r[1], r[2], r[3]
	}
}

func DocumentToHTML(code string, document *lsiftyped.Document) (template.HTML, error) {
	table := &html.Node{Type: html.ElementNode, DataAtom: atom.Table, Data: atom.Table.String()}
	var currentCell *html.Node

	addRow := func(row int32) {
		tr, cell := newRow(row)

		// Add new row to our table
		table.AppendChild(tr)

		// Set current curent cell to the code cell in the row
		currentCell = cell
	}

	addText := func(kind lsiftyped.SyntaxKind, line string) {
		appendText(currentCell, kind, line)
	}

	lsifToHTML(code, document, addRow, addText, nil)

	var buf bytes.Buffer
	if err := html.Render(&buf, table); err != nil {
		return "", err
	}

	return template.HTML(buf.String()), nil
}

func DocumentToSplitHTML(code string, document *lsiftyped.Document) ([]template.HTML, error) {
	rows := []*html.Node{}
	var currentCell *html.Node

	addRow := func(row int32) {
		tr, cell := newRow(row)

		// Add our newest row to our list
		rows = append(rows, tr)

		// Set current cell that we should append text to
		currentCell = cell
	}

	addText := func(kind lsiftyped.SyntaxKind, line string) {
		appendText(currentCell, kind, line)
	}

	lsifToHTML(code, document, addRow, addText, nil)

	lines := make([]template.HTML, 0, len(rows))
	for _, line := range rows {

		var lineHTML bytes.Buffer
		if err := html.Render(&lineHTML, line); err != nil {
			return nil, err
		}

		lines = append(lines, template.HTML(lineHTML.String()))
	}

	return lines, nil
}
