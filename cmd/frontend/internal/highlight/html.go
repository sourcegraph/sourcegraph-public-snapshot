package highlight

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type lsifFormatter interface {
	AddRow(row int32)
	AddText(kind lsiftyped.SyntaxKind, line string)
}

func lsifToHTML(code string, document *lsiftyped.Document, formatter lsifFormatter) {
	splitLines := strings.Split(code, "\n")
	occurences := document.Occurrences

	// table := &html.Node{Type: html.ElementNode, DataAtom: atom.Table, Data: atom.Table.String()}

	row, occIndex := int32(0), 0
	for row < int32(len(splitLines)) {
		line := strings.TrimSuffix(splitLines[row], "\r")
		if line == "" {
			line = "\n" // important for e.g. selecting whitespace in the produced table
		}

		// codeCell := addRow(table, row)
		formatter.AddRow(row)

		lineCharacter := 0
		for occIndex < len(occurences) && occurences[occIndex].Range[0] < row+1 {
			occ := occurences[occIndex]
			occIndex += 1

			startRow, startCharacter, endRow, endCharacter := getRange(occ.Range)
			// appendText(codeCell, "", line[lineCharacter:startCharacter])
			formatter.AddText(occ.SyntaxKind, line[lineCharacter:startCharacter])

			if startRow != endRow {
				// appendText(codeCell, synClass, line[startCharacter:])
				formatter.AddText(occ.SyntaxKind, line[startCharacter:])

				row += 1
				for row < endRow {
					line = splitLines[row]

					// codeCell = addRow(table, row)
					formatter.AddRow(row)
					// appendText(codeCell, synClass, line)
					formatter.AddText(occ.SyntaxKind, line)

					row += 1
				}

				line = splitLines[row]
				// codeCell = addRow(table, row)
				formatter.AddRow(row)
				// appendText(codeCell, synClass, line[:endCharacter])
				formatter.AddText(occ.SyntaxKind, line[:endCharacter])
			} else {
				// appendText(codeCell, synClass, line[startCharacter:endCharacter])
				formatter.AddText(occ.SyntaxKind, line[startCharacter:endCharacter])
			}

			lineCharacter = int(endCharacter)
		}

		// appendText(codeCell, "", line[lineCharacter:])
		formatter.AddText(lsiftyped.SyntaxKind_UnspecifiedSyntaxKind, line[lineCharacter:])

		row += 1
	}

	// var buf bytes.Buffer
	// if err := html.Render(&buf, table); err != nil {
	// 	return "", err
	// }
	// return template.HTML(buf.String()), nil
}

// TODO: This could potentially be an anonymous func
// that captures the info that we want within the other scope.
func appendText(tr *html.Node, kind lsiftyped.SyntaxKind, line string) {
	if line == "" {
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
	spanText := &html.Node{Type: html.TextNode, Data: line}
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

// lsifHTML
type lsifHTML struct {
	table       *html.Node
	currentCell *html.Node
}

func (f *lsifHTML) AddRow(row int32) {
	tr, cell := newRow(row)

	// Add new row to our table
	f.table.AppendChild(tr)

	// Set current curent cell to the code cell in the row
	f.currentCell = cell

	// tr := &html.Node{Type: html.ElementNode, DataAtom: atom.Tr, Data: atom.Tr.String()}
	//
	// tdLineNumber := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
	// tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "class", Val: "line"})
	// tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "data-line", Val: fmt.Sprint(row + 1)})
	// tr.AppendChild(tdLineNumber)
	//
	// f.currentCell = &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
	// f.currentCell.Attr = append(f.currentCell.Attr, html.Attribute{Key: "class", Val: "code"})
	// tr.AppendChild(f.currentCell)
}

func (f *lsifHTML) AddText(kind lsiftyped.SyntaxKind, line string) {
	appendText(f.currentCell, kind, line)
}

// lsifLined
type lsifLined struct {
	rows        []*html.Node
	currentCell *html.Node
}

func (f *lsifLined) AddRow(row int32) {
	tr, cell := newRow(row)

	f.rows = append(f.rows, tr)
	f.currentCell = cell
}

func (f *lsifLined) AddText(kind lsiftyped.SyntaxKind, line string) {
	appendText(f.currentCell, kind, line)
}
