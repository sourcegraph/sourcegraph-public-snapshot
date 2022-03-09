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

func lsifToHTML(code string, document *lsiftyped.Document) (template.HTML, error) {
	splitLines := strings.Split(code, "\n")
	occurences := document.Occurrences

	table := &html.Node{Type: html.ElementNode, DataAtom: atom.Table, Data: atom.Table.String()}

	row, occIndex := int32(0), 0
	for row < int32(len(splitLines)) {
		line := strings.TrimSuffix(splitLines[row], "\r")
		if line == "" {
			line = "\n" // important for e.g. selecting whitespace in the produced table
		}

		codeCell := addRow(table, row)

		lineCharacter := 0
		for occIndex < len(occurences) && occurences[occIndex].Range[0] < int32(row+1) {
			occ := occurences[occIndex]
			occIndex += 1

			var synClass string
			if occ.SyntaxKind != lsiftyped.SyntaxKind_UnspecifiedSyntaxKind {
				synClass = "hl-typed-" + lsiftyped.SyntaxKind_name[int32(occ.SyntaxKind)]
			}

			startRow, startCharacter, endRow, endCharacter := getRange(occ.Range)
			appendText(codeCell, "", line[lineCharacter:startCharacter])

			if startRow != endRow {
				appendText(codeCell, synClass, line[startCharacter:])

				row += 1
				for row < endRow {
					line = splitLines[row]

					codeCell = addRow(table, row)
					appendText(codeCell, synClass, line)

					row += 1
				}

				line = splitLines[row]
				codeCell = addRow(table, row)
				appendText(codeCell, synClass, line[:endCharacter])
			} else {
				appendText(codeCell, synClass, line[startCharacter:endCharacter])
			}

			lineCharacter = int(endCharacter)
		}

		appendText(codeCell, "", line[lineCharacter:])

		row += 1
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, table); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// TODO: This could potentially be an anonymous func
// that captures the info that we want within the other scope.
func appendText(tr *html.Node, class, line string) {
	if line == "" {
		return
	}

	span := &html.Node{Type: html.ElementNode, DataAtom: atom.Span, Data: atom.Span.String()}
	if class != "" {
		span.Attr = append(span.Attr, html.Attribute{Key: "class", Val: class})
	}
	tr.AppendChild(span)
	spanText := &html.Node{Type: html.TextNode, Data: line}
	span.AppendChild(spanText)
}

func addRow(table *html.Node, row int32) *html.Node {

	tr := &html.Node{Type: html.ElementNode, DataAtom: atom.Tr, Data: atom.Tr.String()}
	table.AppendChild(tr)

	tdLineNumber := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
	tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "class", Val: "line"})
	tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "data-line", Val: fmt.Sprint(row + 1)})
	tr.AppendChild(tdLineNumber)

	codeCell := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
	codeCell.Attr = append(codeCell.Attr, html.Attribute{Key: "class", Val: "code"})
	tr.AppendChild(codeCell)

	return codeCell
}

func getRange(r []int32) (int32, int32, int32, int32) {
	if len(r) == 3 {
		return r[0], r[1], r[0], r[2]
	} else {
		return r[0], r[1], r[2], r[3]
	}
}
