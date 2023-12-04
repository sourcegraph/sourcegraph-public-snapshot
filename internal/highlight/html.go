package highlight

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

// DocumentToSplitHTML returns a list of each line of HTML.
func DocumentToSplitHTML(code string, document *scip.Document, includeLineNumbers bool) ([]template.HTML, error) {
	rows := []*html.Node{}
	var currentCell *html.Node

	addRow := func(row int32) {
		tr, cell := newHtmlRow(row, includeLineNumbers)

		// Add our newest row to our list
		rows = append(rows, tr)

		// Set current cell that we should append text to
		currentCell = cell
	}

	addText := func(kind scip.SyntaxKind, line string) {
		appendTextToNode(currentCell, kind, line)
	}

	scipToHTML(code, document, addRow, addText, nil)

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

// DocumentToHTML creates one HTML blob for the entire document
func DocumentToHTML(code string, document *scip.Document) (template.HTML, error) {
	table := &html.Node{Type: html.ElementNode, DataAtom: atom.Table, Data: atom.Table.String()}
	var currentCell *html.Node

	addRow := func(row int32) {
		tr, cell := newHtmlRow(row, true)

		// Add new row to our table
		table.AppendChild(tr)

		// Set current curent cell to the code cell in the row
		currentCell = cell
	}

	addText := func(kind scip.SyntaxKind, line string) {
		appendTextToNode(currentCell, kind, line)
	}

	scipToHTML(code, document, addRow, addText, nil)

	var buf bytes.Buffer
	if err := html.Render(&buf, table); err != nil {
		return "", err
	}

	return template.HTML(buf.String()), nil
}

// safeSlice is used to prevent us from panicing in production when we
// request a slice of runes for lsifToHTML. It's possible that the incoming
// bundle may be malformed, so this way we'll at least try to highlight things
// (and hopefully pick up in the correct place on the next line if things went weird)
func safeSlice(text []rune, start, finish int32) string {
	if start > finish {
		return ""
	}

	if int(start) > len(text) {
		return ""
	}

	if int(finish) > len(text) {
		return string(text[start:])
	}

	return string(text[start:finish])
}

// scipToHTML iterates on code and a document to dispatch to AddRow and AddText
// which can be used to generate different kinds of HTML.
func scipToHTML(
	code string,
	document *scip.Document,
	addRow func(row int32),
	addText func(kind scip.SyntaxKind, line string),
	validLines map[int32]bool,
) {
	manager := newHtmlManager(code, validLines, document.Occurrences, addRow, addText)
	manager.process()
}

type htmlManager struct {
	// Ways to add new rows and text to the output HTML
	addRow  func(row int32)
	addText func(kind scip.SyntaxKind, line string)

	lines       [][]rune
	occurrences []*scip.Occurrence
	row         int32
	occIdx      int

	// Can be nil, should be ignored if nil
	validLines map[int32]bool
}

func newHtmlManager(
	code string,
	validLines map[int32]bool,
	occurrences []*scip.Occurrence,
	addRow func(row int32),
	addText func(kind scip.SyntaxKind, line string),
) *htmlManager {
	splitStringLines := strings.Split(code, "\n")

	// Why split into runes?
	//   Well, my young ascii grasshopper, we are using lines and _columns_
	//   and columns expect things to be indexed by column, not by byte offset.
	//
	//   If we use byte offset (which is what you get when you do myString[x:y])
	//   then you'll be in big trouble for displaying purposes (and probably run over
	//   the end of things).
	//
	//   So, we get a list of list of runes to interact with, which can be correctly
	//   indexed and sliced based on columns.
	//
	//   This could probably use a library (or we are doing something similar elsewhere
	//   and I just didn't know about it)
	splitLines := make([][]rune, len(splitStringLines))
	for idx, line := range splitStringLines {
		// Ensure that line doesn't have trailing \r characters (we already split on \n)
		line = strings.TrimSuffix(line, "\r")

		// important for e.g. selecting whitespace in the produced table
		if line == "" {
			line = "\n"
		}

		splitLines[idx] = []rune(line)
	}

	return &htmlManager{
		lines:   splitLines,
		row:     0,
		addRow:  addRow,
		addText: addText,

		occIdx:      0,
		occurrences: occurrences,

		validLines: validLines,
	}
}

func (t *htmlManager) process() {
	rowCount := int32(len(t.lines))
	for t.row < rowCount {
		t.processRow()
		t.nextRow()
	}
}

func (t *htmlManager) processRow() {
	if !t.validRow(t.row) {
		return
	}

	t.addRow(t.row)

	lineCharacter := int32(0)
	for t.occIdx < len(t.occurrences) && t.occurrences[t.occIdx].Range[0] < t.row+1 {
		occ := t.occurrences[t.occIdx]
		t.occIdx += 1

		lineCharacter = t.processOneOcc(occ, lineCharacter)
	}

	// Add the rest of the line with no syntax highlighting (since it may not be covered by an occurrence).
	line := t.currentLine()
	if lineCharacter != int32(len(line)) {
		t.addText(scip.SyntaxKind_UnspecifiedSyntaxKind, safeSlice(line, lineCharacter, int32(len(line))))
	}
}

func (t *htmlManager) processOneOcc(occ *scip.Occurrence, lineCharacter int32) int32 {
	startRow, startCharacter, endRow, endCharacter := normalizeSCIPRange(occ.Range)

	// We may not have handled all the occurrences up until now
	// so skip the ones where the ranges do not overlap.
	if endRow < t.row {
		return 0
	}

	// Only add the "missed" text if
	if startRow == t.row && lineCharacter != startCharacter {
		currentLine := t.currentLine()
		t.addText(scip.SyntaxKind_UnspecifiedSyntaxKind, safeSlice(currentLine, lineCharacter, startCharacter))
	}

	if startRow == endRow {
		t.processSingleLineOcc(occ, startRow, startCharacter, endCharacter)
	} else {
		t.processMultiLineOcc(occ, startRow, startCharacter, endRow, endCharacter)
	}

	return endCharacter
}

func (t *htmlManager) processMultiLineOcc(occ *scip.Occurrence, startRow, startCharacter, endRow, endCharacter int32) {
	maxRow := int32(len(t.lines))

	// Process the first line
	if t.row == startRow {
		line := t.currentLine()
		t.addText(occ.SyntaxKind, safeSlice(line, startCharacter, int32(len(line))))
		t.nextRow()
		t.addRow(t.row)
	}

	// Process all middle lines (which are fully contained by this occurrence)
	for t.row < endRow && t.row < maxRow {
		t.addText(occ.SyntaxKind, string(t.currentLine()))
		t.nextRow()

		if t.row >= maxRow {
			break
		}

		t.addRow(t.row)
	}

	// Process the last line.
	//   There may be other matches on this line
	if t.row == endRow && t.row < maxRow {
		t.addText(occ.SyntaxKind, safeSlice(t.currentLine(), 0, endCharacter))
		// NOTE:
		//   We do not add nextRow()
		//   This is because this is always handled above. So do not add that here.
	}
}

func (t *htmlManager) processSingleLineOcc(occ *scip.Occurrence, startRow, startCharacter, endCharacter int32) {
	if startRow != t.row {
		return
	}

	line := t.lines[startRow]
	t.addText(occ.SyntaxKind, safeSlice(line, startCharacter, endCharacter))
}

func (t *htmlManager) validRow(row int32) bool { return t.validLines == nil || t.validLines[row] }
func (t *htmlManager) nextRow()                { t.row += 1 }

func (t *htmlManager) currentLine() []rune {
	if t.row >= int32(len(t.lines)) {
		return []rune{}
	}

	return t.lines[t.row]
}

// appendTextToNode formats the text to the right css class and appends to the current
// html node
func appendTextToNode(tr *html.Node, kind scip.SyntaxKind, text string) {
	if text == "" {
		return
	}

	var class string
	if kind != scip.SyntaxKind_UnspecifiedSyntaxKind {
		class = "hl-typed-" + scip.SyntaxKind_name[int32(kind)]
	}

	span := &html.Node{Type: html.ElementNode, DataAtom: atom.Span, Data: atom.Span.String()}
	if class != "" {
		span.Attr = append(span.Attr, html.Attribute{Key: "class", Val: class})
	}
	tr.AppendChild(span)
	spanText := &html.Node{Type: html.TextNode, Data: text}
	span.AppendChild(spanText)
}

// newHtmlRow creates a new row in the style of syntect's tables.
func newHtmlRow(row int32, includeLineNumbers bool) (htmlRow, htmlCode *html.Node) {
	tr := &html.Node{Type: html.ElementNode, DataAtom: atom.Tr, Data: atom.Tr.String()}

	if includeLineNumbers {
		tdLineNumber := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
		tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "class", Val: "line"})
		tdLineNumber.Attr = append(tdLineNumber.Attr, html.Attribute{Key: "data-line", Val: fmt.Sprint(row + 1)})
		tr.AppendChild(tdLineNumber)
	}

	codeCell := &html.Node{Type: html.ElementNode, DataAtom: atom.Td, Data: atom.Td.String()}
	codeCell.Attr = append(codeCell.Attr, html.Attribute{Key: "class", Val: "code"})
	tr.AppendChild(codeCell)

	return tr, codeCell
}

func normalizeSCIPRange(r []int32) (int32, int32, int32, int32) {
	if len(r) == 3 {
		return r[0], r[1], r[0], r[2]
	} else {
		return r[0], r[1], r[2], r[3]
	}
}
