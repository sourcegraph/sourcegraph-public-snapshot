// Functions to parse block-level elements.

package mmark

import (
	"bytes"
	"strconv"
	"strings"
	"unicode"
)

// Parse block-level data.
// Note: this function and many that it calls assume that
// the input buffer ends with a newline.
func (p *parser) block(out *bytes.Buffer, data []byte) {
	if len(data) == 0 || data[len(data)-1] != '\n' {
		panic("block input is missing terminating newline")
	}

	// this is called recursively: enforce a maximum depth
	if p.nesting >= p.maxNesting {
		return
	}
	p.nesting++

	// parse out one block-level construct at a time
	for len(data) > 0 {
		// IAL
		//
		// {.class #id key=value}
		if data[0] == '{' {
			if j := p.isInlineAttr(data); j > 0 {
				data = data[j:]
				continue
			}
		}

		// part header:
		//
		// -# Part
		if p.flags&EXTENSION_PARTS != 0 {
			if p.isPartHeader(data) {
				data = data[p.partHeader(out, data):]
				continue
			}
		}

		// prefixed header:
		//
		// # Header 1
		// ## Header 2
		// ...
		// ###### Header 6
		if p.isPrefixHeader(data) {
			data = data[p.prefixHeader(out, data):]
			continue
		}

		// special header:
		//
		// .# Abstract
		// .# Preface
		if p.isSpecialHeader(data) {
			if i := p.specialHeader(out, data); i > 0 {
				data = data[i:]
				continue
			}
		}

		// block of preformatted HTML:
		//
		// <div>
		//     ...
		// </div>
		if data[0] == '<' {
			if i := p.html(out, data, true); i > 0 {
				data = data[i:]
				continue
			}
		}

		// title block in TOML
		//
		// % stuff = "foo"
		// % port = 1024
		if p.flags&EXTENSION_TITLEBLOCK_TOML != 0 {
			if data[0] == '%' {
				if i := p.titleBlock(out, data, true); i > 0 {
					data = data[i:]
					continue
				}
			}
		}

		// blank lines.  note: returns the # of bytes to skip
		if i := p.isEmpty(data); i > 0 {
			data = data[i:]
			continue
		}

		// indented code block:
		//
		//     func max(a, b int) int {
		//         if a > b {
		//             return a
		//         }
		//         return b
		//      }
		if p.codePrefix(data) > 0 {
			data = data[p.code(out, data):]
			continue
		}

		// fenced code block:
		//
		// ``` go
		// func fact(n int) int {
		//     if n <= 1 {
		//         return n
		//     }
		//     return n * fact(n-1)
		// }
		// ```
		if p.flags&EXTENSION_FENCED_CODE != 0 {
			if i := p.fencedCode(out, data, true); i > 0 {
				data = data[i:]
				continue
			}
		}

		// horizontal rule:
		//
		// ------
		// or
		// ******
		// or
		// ______
		if p.isHRule(data) {
			p.r.HRule(out)
			var i int
			for i = 0; data[i] != '\n'; i++ {
			}
			data = data[i:]
			continue
		}

		// Aside quote:
		//
		// A> This is an aside
		// A> I found on the web
		if p.asidePrefix(data) > 0 {
			data = data[p.aside(out, data):]
			continue
		}

		// Note quote:
		//
		// N> This is an aside
		// N> I found on the web
		if p.notePrefix(data) > 0 {
			data = data[p.note(out, data):]
			continue
		}

		// Figure "quote":
		//
		// F> ![](image)
		// F> ![](image)
		// Figure: Caption.
		if p.figurePrefix(data) > 0 {
			data = data[p.figure(out, data):]
			continue
		}

		// block quote:
		//
		// > A big quote I found somewhere
		// > on the web
		if p.quotePrefix(data) > 0 {
			data = data[p.quote(out, data):]
			continue
		}

		// table:
		//
		// Name  | Age | Phone
		// ------|-----|---------
		// Bob   | 31  | 555-1234
		// Alice | 27  | 555-4321
		// Table: this is a caption
		if p.flags&EXTENSION_TABLES != 0 {
			if i := p.table(out, data); i > 0 {
				data = data[i:]
				continue
			}
		}

		// block table:
		// (cells contain block elements)
		//
		// |-------|-----|---------
		// | Name  | Age | Phone
		// | ------|-----|---------
		// | Bob   | 31  | 555-1234
		// | Alice | 27  | 555-4321
		// |-------|-----|---------
		// | Bob   | 31  | 555-1234
		// | Alice | 27  | 555-4321
		if p.flags&EXTENSION_TABLES != 0 {
			if i := p.blockTable(out, data); i > 0 {
				data = data[i:]
				continue
			}
		}

		// a definition list:
		//
		// Item1
		// :	Definition1
		// Item2
		// :	Definition2
		if p.dliPrefix(data) > 0 {
			p.insideDefinitionList = true
			data = data[p.list(out, data, _LIST_TYPE_DEFINITION, 0, nil):]
			p.insideDefinitionList = false
		}

		// an itemized/unordered list:
		//
		// * Item 1
		// * Item 2
		//
		// also works with + or -
		if p.uliPrefix(data) > 0 {
			data = data[p.list(out, data, 0, 0, nil):]
			continue
		}

		// a numbered/ordered list:
		//
		// 1. Item 1
		// 2. Item 2
		if i := p.oliPrefix(data); i > 0 {
			start := 0
			if i > 2 {
				start, _ = strconv.Atoi(string(data[:i-2])) // this cannot fail because we just est. the thing *is* a number, and if it does start is zero anyway.
			}

			data = data[p.list(out, data, _LIST_TYPE_ORDERED, start, nil):]
			continue
		}

		// a numberd/ordered list:
		//
		// ii.  Item 1
		// ii.  Item 2
		if p.rliPrefix(data) > 0 {
			data = data[p.list(out, data, _LIST_TYPE_ORDERED|_LIST_TYPE_ORDERED_ROMAN_LOWER, 0, nil):]
			continue
		}

		// a numberd/ordered list:
		//
		// II.  Item 1
		// II.  Item 2
		if p.rliPrefixU(data) > 0 {
			data = data[p.list(out, data, _LIST_TYPE_ORDERED|_LIST_TYPE_ORDERED_ROMAN_UPPER, 0, nil):]
			continue
		}

		// a numberd/ordered list:
		//
		// a.  Item 1
		// b.  Item 2
		if p.aliPrefix(data) > 0 {
			data = data[p.list(out, data, _LIST_TYPE_ORDERED|_LIST_TYPE_ORDERED_ALPHA_LOWER, 0, nil):]
			continue
		}

		// a numberd/ordered list:
		//
		// A.  Item 1
		// B.  Item 2
		if p.aliPrefixU(data) > 0 {
			data = data[p.list(out, data, _LIST_TYPE_ORDERED|_LIST_TYPE_ORDERED_ALPHA_UPPER, 0, nil):]
			continue
		}

		// an example lists:
		//
		// (@good)  Item1
		// (@good)  Item2
		if i := p.eliPrefix(data); i > 0 {
			group := data[2 : i-2]
			data = data[p.list(out, data, _LIST_TYPE_ORDERED|_LIST_TYPE_ORDERED_GROUP, 0, group):]
			continue
		}
		// anything else must look like a normal paragraph
		// note: this finds underlined headers, too
		data = data[p.paragraph(out, data):]
	}

	p.nesting--
}

func (p *parser) isPrefixHeader(data []byte) bool {
	// CommonMark: up to three spaces allowed
	k := 0
	for k < len(data) && data[k] == ' ' {
		k++
	}
	if k == len(data) || k > 3 {
		return false
	}
	data = data[k:]

	if data[0] != '#' {
		return false
	}

	if p.flags&EXTENSION_SPACE_HEADERS != 0 {
		level := 0
		for level < 6 && data[level] == '#' {
			level++
		}
		if !iswhitespace(data[level]) {
			return false
		}
	}
	return true
}

func (p *parser) prefixHeader(out *bytes.Buffer, data []byte) int {
	// CommonMark: up to three spaces allowed
	k := 0
	for k < len(data) && data[k] == ' ' {
		k++
	}
	if k == len(data) || k > 3 {
		return 0
	}
	data = data[k:]

	level := 0
	for level < 6 && data[level] == '#' {
		level++
	}
	i, end := 0, 0
	for i = level; iswhitespace(data[i]); i++ {
	}
	for end = i; data[end] != '\n'; end++ {
	}
	skip := end
	id := ""
	if p.flags&EXTENSION_HEADER_IDS != 0 {
		j, k := 0, 0
		// find start/end of header id
		for j = i; j < end-1 && (data[j] != '{' || data[j+1] != '#'); j++ {
		}
		for k = j + 1; k < end && data[k] != '}'; k++ {
		}
		// extract header id iff found
		if j < end && k < end {
			id = string(data[j+2 : k])
			end = j
			skip = k + 1
			for end > 0 && data[end-1] == ' ' {
				end--
			}
		}
	}
	// CommonMark spaces *after* the header
	for end > 0 && data[end-1] == ' ' {
		end--
	}
	for end > 0 && data[end-1] == '#' {
		// CommonMark: a # directly following the header name is allowed and we
		// should keep it
		if end > 1 && data[end-2] != '#' && !iswhitespace(data[end-2]) {
			end++
			break
		}
		end--
	}
	for end > 0 && iswhitespace(data[end-1]) {
		end--
	}
	if end > i {
		if id == "" && p.flags&EXTENSION_AUTO_HEADER_IDS != 0 {
			id = createSanitizedAnchorName(string(data[i:end]))
		}
		work := func() bool {
			p.inline(out, data[i:end])
			return true
		}
		if id != "" {
			if v, ok := p.anchors[id]; ok && p.flags&EXTENSION_UNIQUE_HEADER_IDS != 0 {
				p.anchors[id]++
				// anchor found
				id += "-" + strconv.Itoa(v)
			} else {
				p.anchors[id] = 1
			}
		}

		p.r.SetInlineAttr(p.ial)
		p.ial = nil

		p.r.Header(out, work, level, id)
	}
	return skip + k
}

func (p *parser) isUnderlinedHeader(data []byte) int {
	// test of level 1 header
	if data[0] == '=' {
		i := 1
		for data[i] == '=' {
			i++
		}
		for iswhitespace(data[i]) {
			i++
		}
		if data[i] == '\n' {
			return 1
		} else {
			return 0
		}
	}

	// test of level 2 header
	if data[0] == '-' {
		i := 1
		for data[i] == '-' {
			i++
		}
		for iswhitespace(data[i]) {
			i++
		}
		if data[i] == '\n' {
			return 2
		} else {
			return 0
		}
	}

	return 0
}

func (p *parser) isPartHeader(data []byte) bool {
	k := 0
	for k < len(data) && data[k] == ' ' {
		k++
	}
	if k == len(data) || k > 3 {
		return false
	}
	data = data[k:]
	if len(data) < 3 {
		return false
	}

	if data[0] != '-' || data[1] != '#' {
		return false
	}

	if p.flags&EXTENSION_SPACE_HEADERS != 0 {
		if !iswhitespace(data[2]) {
			return false
		}
	}
	return true
}

func (p *parser) isSpecialHeader(data []byte) bool {
	k := 0
	for k < len(data) && data[k] == ' ' {
		k++
	}
	if k == len(data) || k > 3 {
		return false
	}
	data = data[k:]
	if len(data) < 3 {
		return false
	}

	if data[0] != '.' || data[1] != '#' {
		return false
	}

	if p.flags&EXTENSION_SPACE_HEADERS != 0 {
		if !iswhitespace(data[2]) {
			return false
		}
	}
	return true
}

func (p *parser) specialHeader(out *bytes.Buffer, data []byte) int {
	k := 0
	for k < len(data) && data[k] == ' ' {
		k++
	}
	if k == len(data) || k > 3 {
		return 0
	}
	data = data[k:]
	if len(data) < 3 {
		return 0
	}

	if data[0] != '.' || data[1] != '#' {
		return 0
	}

	i, end := 0, 0
	for i = 2; iswhitespace(data[i]); i++ {
	}
	for end = i; data[end] != '\n'; end++ {
	}
	skip := end
	id := ""

	if p.flags&EXTENSION_HEADER_IDS != 0 {
		j, k := 0, 0
		// find start/end of header id
		for j = i; j < end-1 && (data[j] != '{' || data[j+1] != '#'); j++ {
		}
		for k = j + 1; k < end && data[k] != '}'; k++ {
		}
		// extract header id iff found
		if j < end && k < end {
			id = string(data[j+2 : k])
			end = j
			skip = k + 1
			for end > 0 && data[end-1] == ' ' {
				end--
			}
		}
	}
	// CommonMark spaces *after* the header
	for end > 0 && data[end-1] == ' ' {
		end--
	}
	// Remove this, not true for this header
	for end > 0 && data[end-1] == '#' {
		// CommonMark: a # directly following the header name is allowed and we
		// should keep it
		if end > 1 && data[end-2] != '#' && !iswhitespace(data[end-2]) {
			end++
			break
		}
		end--
	}
	for end > 0 && iswhitespace(data[end-1]) {
		end--
	}
	if end > i {
		if id == "" && p.flags&EXTENSION_AUTO_HEADER_IDS != 0 {
			id = createSanitizedAnchorName(string(data[i:end]))
		}
		work := func() bool {
			p.inline(out, data[i:end])
			return true
		}
		p.r.SetInlineAttr(p.ial)
		p.ial = nil

		name := bytes.ToLower(data[i:end])
		if bytes.Compare(name, []byte("abstract")) != 0 &&
			bytes.Compare(name, []byte("preface")) != 0 {
			return 0
		}
		if id != "" {
			if v, ok := p.anchors[id]; ok && p.flags&EXTENSION_UNIQUE_HEADER_IDS != 0 {
				p.anchors[id]++
				// anchor found
				id += "-" + strconv.Itoa(v)
			} else {
				p.anchors[id] = 1
			}
		}

		p.r.SpecialHeader(out, name, work, id)
	}
	return skip + k
}

func (p *parser) partHeader(out *bytes.Buffer, data []byte) int {
	k := 0
	for k < len(data) && data[k] == ' ' {
		k++
	}
	if k == len(data) || k > 3 {
		return 0
	}
	data = data[k:]
	if len(data) < 3 {
		return 0
	}

	if data[0] != '-' || data[1] != '#' {
		return 0
	}

	i, end := 0, 0
	for i = 2; iswhitespace(data[i]); i++ {
	}
	for end = i; data[end] != '\n'; end++ {
	}
	skip := end
	id := ""

	if p.flags&EXTENSION_HEADER_IDS != 0 {
		j, k := 0, 0
		// find start/end of header id
		for j = i; j < end-1 && (data[j] != '{' || data[j+1] != '#'); j++ {
		}
		for k = j + 1; k < end && data[k] != '}'; k++ {
		}
		// extract header id iff found
		if j < end && k < end {
			id = string(data[j+2 : k])
			end = j
			skip = k + 1
			for end > 0 && data[end-1] == ' ' {
				end--
			}
		}
	}
	// CommonMark spaces *after* the header
	for end > 0 && data[end-1] == ' ' {
		end--
	}
	for end > 0 && data[end-1] == '#' {
		// CommonMark: a # directly following the header name is allowed and we
		// should keep it
		if end > 1 && data[end-2] != '#' && !iswhitespace(data[end-2]) {
			end++
			break
		}
		end--
	}
	for end > 0 && iswhitespace(data[end-1]) {
		end--
	}
	if end > i {
		if id == "" && p.flags&EXTENSION_AUTO_HEADER_IDS != 0 {
			id = createSanitizedAnchorName(string(data[i:end]))
		}
		work := func() bool {
			p.inline(out, data[i:end])
			return true
		}
		if id != "" {
			if v, ok := p.anchors[id]; ok && p.flags&EXTENSION_UNIQUE_HEADER_IDS != 0 {
				p.anchors[id]++
				// anchor found
				id += "-" + strconv.Itoa(v)
			} else {
				p.anchors[id] = 1
			}
		}

		p.r.SetInlineAttr(p.ial)
		p.ial = nil

		p.r.Part(out, work, id)
	}
	return skip + k
}

func (p *parser) titleBlock(out *bytes.Buffer, data []byte, doRender bool) int {
	if p.titleblock {
		return 0
	}
	if data[0] != '%' {
		return 0
	}
	splitData := bytes.Split(data, []byte("\n"))
	var i int
	for idx, b := range splitData {
		if !bytes.HasPrefix(b, []byte("%")) {
			i = idx // - 1
			break
		}
	}
	p.titleblock = true
	data = bytes.Join(splitData[0:i], []byte("\n"))
	block := p.titleBlockTOML(out, data)
	p.r.TitleBlockTOML(out, &block)
	return len(data)
}

func (p *parser) html(out *bytes.Buffer, data []byte, doRender bool) int {
	var i, j int

	// identify the opening tag
	if data[0] != '<' {
		return 0
	}
	curtag, tagfound := p.htmlFindTag(data[1:])

	// handle special cases
	if !tagfound {
		// check for an HTML comment
		if size := p.htmlComment(out, data, doRender); size > 0 {
			return size
		}

		// check for an <hr> tag
		if size := p.htmlHr(out, data, doRender); size > 0 {
			return size
		}

		// check for an <reference>
		if size := p.htmlReference(out, data, doRender); size > 0 {
			return size
		}

		// no special case recognized
		return 0
	}

	// look for an unindented matching closing tag
	// followed by a blank line
	found := false

	// if not found, try a second pass looking for indented match
	// but not if tag is "ins" or "del" (following original Markdown.pl)
	if !found && curtag != "ins" && curtag != "del" {
		i = 1
		for i < len(data) {
			i++
			for i < len(data) && !(data[i-1] == '<' && data[i] == '/') {
				i++
			}

			if i+2+len(curtag) >= len(data) {
				break
			}

			j = p.htmlFindEnd(curtag, data[i-1:])

			if j > 0 {
				i += j - 1
				found = true
				break
			}
		}
	}

	if !found {
		return 0
	}

	// the end of the block has been found
	if doRender {
		// trim newlines
		end := i
		for end > 0 && data[end-1] == '\n' {
			end--
		}
		p.r.BlockHtml(out, data[:end])
	}

	return i
}

// HTML comment, lax form
func (p *parser) htmlComment(out *bytes.Buffer, data []byte, doRender bool) int {
	if data[0] != '<' || data[1] != '!' || data[2] != '-' || data[3] != '-' {
		return 0
	}

	i := 5

	// scan for an end-of-comment marker, across lines if necessary
	for i < len(data) && !(data[i-2] == '-' && data[i-1] == '-' && data[i] == '>') {
		i++
	}
	i++

	// no end-of-comment marker
	if i >= len(data) {
		return 0
	}

	// needs to end with a blank line
	if j := p.isEmpty(data[i:]); j > 0 {
		size := i + j
		if doRender {
			// trim trailing newlines
			end := size
			for end > 0 && data[end-1] == '\n' {
				end--
			}
			// breaks the tests if we parse this
			//			var cooked bytes.Buffer
			//			p.inline(&cooked, data[:end])

			p.r.SetInlineAttr(p.ial)
			p.ial = nil

			p.r.CommentHtml(out, data[:end])
		}
		return size
	}

	return 0
}

// HR, which is the only self-closing block tag considered
func (p *parser) htmlHr(out *bytes.Buffer, data []byte, doRender bool) int {
	if data[0] != '<' || (data[1] != 'h' && data[1] != 'H') || (data[2] != 'r' && data[2] != 'R') {
		return 0
	}
	if data[3] != ' ' && data[3] != '/' && data[3] != '>' {
		// not an <hr> tag after all; at least not a valid one
		return 0
	}

	i := 3
	for data[i] != '>' && data[i] != '\n' {
		i++
	}

	if data[i] == '>' {
		i++
		if j := p.isEmpty(data[i:]); j > 0 {
			size := i + j
			if doRender {
				// trim newlines
				end := size
				for end > 0 && data[end-1] == '\n' {
					end--
				}
				p.r.BlockHtml(out, data[:end])
			}
			return size
		}
	}

	return 0
}

// HTML reference, actually xml, but keep the spirit and call it html
func (p *parser) htmlReference(out *bytes.Buffer, data []byte, doRender bool) int {
	if !bytes.HasPrefix(data, []byte("<reference ")) {
		return 0
	}

	i := 10
	// scan for an end-of-reference marker, across lines if necessary
	for i < len(data) &&
		!(data[i-10] == 'r' && data[i-9] == 'e' && data[i-8] == 'f' &&
			data[i-7] == 'e' && data[i-6] == 'r' && data[i-5] == 'e' &&
			data[i-4] == 'n' && data[i-3] == 'c' && data[i-2] == 'e' &&
			data[i-1] == '>') {
		i++
	}
	i++

	// no end-of-reference marker
	if i >= len(data) {
		return 0
	}

	// needs to end with a blank line
	if j := p.isEmpty(data[i:]); j > 0 {
		size := i + j
		if doRender {
			// trim trailing newlines
			end := size
			for end > 0 && data[end-1] == '\n' {
				end--
			}
			anchor := bytes.Index(data[:end], []byte("anchor="))
			if anchor == -1 {
				// nothing found, not a real reference
				return 0
			}
			// look for the some tag after anchor=
			open := data[anchor+7]
			i := anchor + 7 + 2
			for i < end && data[i-1] != open {
				i++
			}
			if i >= end {
				return 0
			}
			anchorStr := string(data[anchor+7+1 : i-1])
			if c, ok := p.citations[anchorStr]; !ok {
				p.citations[anchorStr] = &citation{xml: data[:end]}
			} else {
				c.xml = data[:end]
			}
		}
		return size
	}

	return 0
}

func (p *parser) htmlFindTag(data []byte) (string, bool) {
	i := 0
	for isalnum(data[i]) {
		i++
	}
	key := string(data[:i])
	if blockTags[key] {
		return key, true
	}
	return "", false
}

func (p *parser) htmlFindEnd(tag string, data []byte) int {
	// assume data[0] == '<' && data[1] == '/' already tested

	// check if tag is a match
	closetag := []byte("</" + tag + ">")
	if !bytes.HasPrefix(data, closetag) {
		return 0
	}
	i := len(closetag)

	// check that the rest of the line is blank
	skip := 0
	if skip = p.isEmpty(data[i:]); skip == 0 {
		return 0
	}
	i += skip
	skip = 0

	if i >= len(data) {
		return i
	}

	if p.flags&EXTENSION_LAX_HTML_BLOCKS != 0 {
		return i
	}
	if skip = p.isEmpty(data[i:]); skip == 0 {
		// following line must be blank
		return 0
	}

	return i + skip
}

func (p *parser) isEmpty(data []byte) int {
	// it is okay to call isEmpty on an empty buffer
	if len(data) == 0 {
		return 0
	}

	var i int
	for i = 0; i < len(data) && data[i] != '\n'; i++ {
		if data[i] != ' ' && data[i] != '\t' {
			return 0
		}
	}
	return i + 1
}

func (p *parser) isHRule(data []byte) bool {
	i := 0

	// skip up to three spaces
	for i < 3 && data[i] == ' ' {
		i++
	}

	// look at the hrule char
	if data[i] != '*' && data[i] != '-' && data[i] != '_' {
		return false
	}
	c := data[i]

	// the whole line must be the char or whitespace
	n := 0
	for data[i] != '\n' {
		switch {
		case data[i] == c:
			n++
		case data[i] != ' ':
			return false
		}
		i++
	}

	return n >= 3
}

func (p *parser) isFencedCode(data []byte, syntax **string, oldmarker string) (skip int, marker string) {
	i, size := 0, 0
	skip = 0

	// skip up to three spaces
	for i < len(data) && i < 3 && data[i] == ' ' {
		i++
	}
	if i >= len(data) {
		return
	}

	// check for the marker characters: ~ or `
	if data[i] != '~' && data[i] != '`' {
		return
	}

	c := data[i]

	// the whole line must be the same char per CommonMark whitespace is not allowed
	for i < len(data) && data[i] == c {
		size++
		i++
	}
	if i >= len(data) {
		return
	}
	// if we find spaces and them some more markers, this is not a fenced code block
	j := i
	for j < len(data) && data[j] == ' ' {
		j++
	}
	if j >= len(data) {
		return
	}
	if data[j] == c {
		return
	}

	// the marker char must occur at least 3 times
	if size < 3 {
		return
	}
	marker = string(data[i-size : i])

	// if this is the end marker, it must be at least as long as the
	// original marker we have
	if oldmarker != "" && !strings.HasPrefix(marker, oldmarker) {
		return
	}

	if syntax != nil {
		syn := 0

		for i < len(data) && data[i] == ' ' {
			i++
		}
		if i >= len(data) {
			return
		}

		syntaxStart := i

		if data[i] == '{' {
			i++
			syntaxStart++

			for i < len(data) && data[i] != '}' && data[i] != '\n' {
				syn++
				i++
			}

			if i >= len(data) && data[i] != '}' {
				return
			}

			// strip all whitespace at the beginning and the end
			// of the {} block
			for syn > 0 && isspace(data[syntaxStart]) {
				syntaxStart++
				syn--
			}

			for syn > 0 && isspace(data[syntaxStart+syn-1]) {
				syn--
			}

			i++
		} else {
			for i < len(data) && !isspace(data[i]) {
				syn++
				i++
			}
		}

		language := string(data[syntaxStart : syntaxStart+syn])
		*syntax = &language
	}

	// CommonMark: skip garbage until end of line
	for i < len(data) && data[i] != '\n' {
		i++
	}

	skip = i + 1
	return
}

func (p *parser) fencedCode(out *bytes.Buffer, data []byte, doRender bool) int {
	var lang *string
	beg, marker := p.isFencedCode(data, &lang, "")
	if beg == 0 {
		return 0
	}

	co := ""
	if p.ial != nil {
		// enabled, any non-empty value
		co = p.ial.Value("callout")
	}

	if beg >= len(data) {
		// only the marker and end of doc. CommonMark dictates this is valid

		p.r.SetInlineAttr(p.ial)
		p.ial = nil

		// Data here?
		p.r.BlockCode(out, nil, "", nil, p.insideFigure, false)

		return len(data)
	}

	// CommonMark: if indented strip this many leading spaces from code block
	indent := 0
	for indent < beg && data[indent] == ' ' {
		indent++
	}

	var work bytes.Buffer

	for {
		// safe to assume beg < len(data)

		// check for the end of the code block
		fenceEnd, _ := p.isFencedCode(data[beg:], nil, marker)
		if fenceEnd != 0 {
			beg += fenceEnd
			break
		}

		// copy the current line
		end := beg
		for end < len(data) && data[end] != '\n' {
			end++
		}
		end++

		// did we reach the end of the buffer without a closing marker?
		// CommonMark: end of buffer closes fencec code block
		if end >= len(data) {
			work.Write(data[beg:])
			beg = end
			break
		}

		// CommmonMark, strip beginning spaces
		s := 0
		for s < indent && data[beg] == ' ' {
			beg++
			s++
		}

		// verbatim copy to the working buffer
		if doRender {
			work.Write(data[beg:end])
		}
		beg = end
	}
	var caption bytes.Buffer
	line := beg
	j := beg
	if bytes.HasPrefix(bytes.TrimSpace(data[j:]), []byte("Figure: ")) {
		for line < len(data) {
			j++
			// find the end of this line
			for data[j-1] != '\n' {
				j++
			}
			if p.isEmpty(data[line:j]) > 0 {
				break
			}
			line = j
		}
		p.inline(&caption, data[beg+8:j-1])
	}

	syntax := ""
	if lang != nil {
		syntax = *lang
	}

	if doRender {
		p.r.SetInlineAttr(p.ial)
		p.ial = nil
		if co != "" {
			var callout bytes.Buffer
			callouts(p, &callout, work.Bytes(), 0, co)
			p.r.BlockCode(out, callout.Bytes(), syntax, caption.Bytes(), p.insideFigure, true)
		} else {
			p.callouts = nil
			p.r.BlockCode(out, work.Bytes(), syntax, caption.Bytes(), p.insideFigure, false)
		}
	}

	return j
}

func (p *parser) table(out *bytes.Buffer, data []byte) int {
	var (
		header bytes.Buffer
		body   bytes.Buffer
		footer bytes.Buffer
	)
	i, columns := p.tableHeader(&header, data)
	if i == 0 {
		return 0
	}

	foot := false

	for i < len(data) {
		if j := p.isTableFooter(data[i:]); j > 0 && !foot {
			foot = true
			i += j
			continue
		}
		pipes, rowStart := 0, i
		for ; data[i] != '\n'; i++ {
			if data[i] == '|' {
				pipes++
			}
		}

		if pipes == 0 {
			i = rowStart
			break
		}

		// include the newline in data sent to tableRow
		i++
		if foot {
			p.tableRow(&footer, data[rowStart:i], columns, false)
			continue
		}
		p.tableRow(&body, data[rowStart:i], columns, false)
	}
	var caption bytes.Buffer
	line := i
	j := i
	if bytes.HasPrefix(data[j:], []byte("Table: ")) {
		for line < len(data) {
			j++
			// find the end of this line
			for data[j-1] != '\n' {
				j++
			}
			if p.isEmpty(data[line:j]) > 0 {
				break
			}
			line = j
		}
		p.inline(&caption, data[i+7:j-1]) // +7 for 'Table: '
	}

	p.r.SetInlineAttr(p.ial)
	p.ial = nil

	p.r.Table(out, header.Bytes(), body.Bytes(), footer.Bytes(), columns, caption.Bytes())

	return j
}

func (p *parser) blockTable(out *bytes.Buffer, data []byte) int {
	var (
		header  bytes.Buffer
		body    bytes.Buffer
		footer  bytes.Buffer
		rowWork bytes.Buffer
	)
	i := p.isBlockTableHeader(data)
	if i == 0 || i == len(data) {
		return 0
	}
	j, columns := p.tableHeader(&header, data[i:])
	if i == 0 {
		return 0
	}
	i += j
	// each cell in a row gets multiple lines which we store per column, we
	// process the buffers when we see a row separator (isBlockTableHeader)
	bodies := make([]bytes.Buffer, len(columns))

	foot := false

	j = 0
	for i < len(data) {
		if j = p.isTableFooter(data[i:]); j > 0 && !foot {
			// prepare previous ones
			foot = true
			i += j
			continue
		}
		if j = p.isBlockTableHeader(data[i:]); j > 0 {
			switch foot {
			case false: // separator before any footer
				var cellWork bytes.Buffer
				for c := 0; c < len(columns); c++ {
					cellWork.Truncate(0)
					if bodies[c].Len() > 0 {
						p.block(&cellWork, bodies[c].Bytes())
						bodies[c].Truncate(0)
					}
					p.r.TableCell(&rowWork, cellWork.Bytes(), columns[c])
				}
				p.r.TableRow(&body, rowWork.Bytes())
				rowWork.Truncate(0)
				i += j
				continue

			case true: // closing separator that closes the table
				i += j
				continue
			}
		}

		pipes, rowStart := 0, i
		for ; data[i] != '\n'; i++ {
			if data[i] == '|' {
				pipes++
			}
		}

		if pipes == 0 {
			i = rowStart
			break
		}

		// include the newline in data sent to tableRow and blockTabeRow
		i++
		if foot {
			p.tableRow(&footer, data[rowStart:i], columns, false)
		} else {
			p.blockTableRow(bodies, data[rowStart:i])
		}
	}
	// are there cells left to process?
	if len(bodies) > 0 && bodies[0].Len() != 0 {
		for c := 0; c < len(columns); c++ {
			var cellWork bytes.Buffer
			cellWork.Truncate(0)
			if bodies[c].Len() > 0 {
				p.block(&cellWork, bodies[c].Bytes())
				bodies[c].Truncate(0)
			}
			p.r.TableCell(&rowWork, cellWork.Bytes(), columns[c])
		}
		p.r.TableRow(&body, rowWork.Bytes())
	}

	var caption bytes.Buffer
	line := i
	j = i
	if bytes.HasPrefix(data[j:], []byte("Table: ")) {
		for line < len(data) {
			j++
			// find the end of this line
			for data[j-1] != '\n' {
				j++
			}
			if p.isEmpty(data[line:j]) > 0 {
				break
			}
			line = j
		}
		p.inline(&caption, data[i+7:j-1]) // +7 for 'Table: '
	}

	p.r.SetInlineAttr(p.ial)
	p.ial = nil

	p.r.Table(out, header.Bytes(), body.Bytes(), footer.Bytes(), columns, caption.Bytes())

	return j
}

// check if the specified position is preceeded by an odd number of backslashes
func isBackslashEscaped(data []byte, i int) bool {
	backslashes := 0
	for i-backslashes-1 >= 0 && data[i-backslashes-1] == '\\' {
		backslashes++
	}
	return backslashes&1 == 1
}

func (p *parser) tableHeader(out *bytes.Buffer, data []byte) (size int, columns []int) {
	i := 0
	colCount := 1
	for i = 0; data[i] != '\n'; i++ {
		if data[i] == '|' && !isBackslashEscaped(data, i) {
			colCount++
		}
	}

	// doesn't look like a table header
	if colCount == 1 {
		return
	}

	// include the newline in the data sent to tableRow
	header := data[:i+1]

	// column count ignores pipes at beginning or end of line
	if data[0] == '|' {
		colCount--
	}
	if i > 2 && data[i-1] == '|' && !isBackslashEscaped(data, i-1) {
		colCount--
	}

	columns = make([]int, colCount)

	// move on to the header underline
	i++
	if i >= len(data) {
		return
	}

	if data[i] == '|' && !isBackslashEscaped(data, i) {
		i++
	}
	for data[i] == ' ' {
		i++
	}

	// each column header is of form: / *:?-+:? *|/ with # dashes + # colons >= 3
	// and trailing | optional on last column
	col := 0
	for data[i] != '\n' {
		dashes := 0

		if data[i] == ':' {
			i++
			columns[col] |= _TABLE_ALIGNMENT_LEFT
			dashes++
		}
		for data[i] == '-' {
			i++
			dashes++
		}
		if data[i] == ':' {
			i++
			columns[col] |= _TABLE_ALIGNMENT_RIGHT
			dashes++
		}
		for data[i] == ' ' {
			i++
		}

		// end of column test is messy
		switch {
		case dashes < 3:
			// not a valid column
			return

		case data[i] == '|' && !isBackslashEscaped(data, i):
			// marker found, now skip past trailing whitespace
			col++
			i++
			for data[i] == ' ' {
				i++
			}

			// trailing junk found after last column
			if col >= colCount && data[i] != '\n' {
				return
			}

		case (data[i] != '|' || isBackslashEscaped(data, i)) && col+1 < colCount:
			// something else found where marker was required
			return

		case data[i] == '\n':
			// marker is optional for the last column
			col++

		default:
			// trailing junk found after last column
			return
		}
	}
	if col != colCount {
		return
	}

	p.tableRow(out, header, columns, true)
	size = i + 1
	return
}

func (p *parser) tableRow(out *bytes.Buffer, data []byte, columns []int, header bool) {
	i, col := 0, 0
	var rowWork bytes.Buffer

	if data[i] == '|' && !isBackslashEscaped(data, i) {
		i++
	}

	for col = 0; col < len(columns) && i < len(data); col++ {
		for data[i] == ' ' {
			i++
		}

		cellStart := i

		for (data[i] != '|' || isBackslashEscaped(data, i)) && data[i] != '\n' {
			i++
		}

		cellEnd := i

		// skip the end-of-cell marker, possibly taking us past end of buffer
		i++

		for cellEnd > cellStart && data[cellEnd-1] == ' ' {
			cellEnd--
		}

		var cellWork bytes.Buffer
		p.inline(&cellWork, data[cellStart:cellEnd])

		if header {
			p.r.TableHeaderCell(&rowWork, cellWork.Bytes(), columns[col])
		} else {
			p.r.TableCell(&rowWork, cellWork.Bytes(), columns[col])
		}
	}

	// pad it out with empty columns to get the right number
	for ; col < len(columns); col++ {
		if header {
			p.r.TableHeaderCell(&rowWork, nil, columns[col])
		} else {
			p.r.TableCell(&rowWork, nil, columns[col])
		}
	}

	// silently ignore rows with too many cells

	p.r.TableRow(out, rowWork.Bytes())
}

func (p *parser) blockTableRow(out []bytes.Buffer, data []byte) {
	i, col := 0, 0

	if data[i] == '|' && !isBackslashEscaped(data, i) {
		i++
	}

	for col = 0; col < len(out) && i < len(data); col++ {
		space := i
		for data[i] == ' ' {
			space++
			i++
		}

		cellStart := i

		for (data[i] != '|' || isBackslashEscaped(data, i)) && data[i] != '\n' {
			i++
		}

		cellEnd := i

		// skip the end-of-cell marker, possibly taking us past end of buffer
		i++

		for cellEnd > cellStart && data[cellEnd-1] == ' ' {
			cellEnd--
		}
		out[col].Write(data[cellStart:cellEnd])
		out[col].WriteByte('\n')
	}
}

// optional | or + at the beginning, then at least 3 equals
func (p *parser) isTableFooter(data []byte) int {
	i := 0
	if data[i] == '|' || data[i] == '+' {
		i++
	}
	if len(data[i:]) < 4 {
		return 0
	}
	if data[i+1] != '=' && data[i+2] != '=' && data[i+3] != '=' {
		return 0
	}
	for i < len(data) && data[i] != '\n' {
		i++
	}
	return i + 1
}

// this starts a table and also serves as a row divider, basically three dashes with optional | or + at the start
func (p *parser) isBlockTableHeader(data []byte) int {
	i := 0
	if data[i] != '|' && data[i] != '+' {
		return 0
	}
	i++
	if len(data[i:]) < 4 {
		return 0
	}
	if data[i+1] != '-' && data[i+2] != '-' && data[i+3] != '-' {
		return 0
	}
	for i < len(data) && data[i] != '\n' {
		i++
	}
	return i + 1
}

// returns prefix length for block code
func (p *parser) codePrefix(data []byte) int {
	if data[0] == ' ' && data[1] == ' ' && data[2] == ' ' && data[3] == ' ' {
		return 4
	}
	return 0
}

func (p *parser) code(out *bytes.Buffer, data []byte) int {
	var work bytes.Buffer
	i := 0
	for i < len(data) {
		beg := i
		for data[i] != '\n' {
			i++
		}
		i++

		blankline := p.isEmpty(data[beg:i]) > 0
		if pre := p.codePrefix(data[beg:i]); pre > 0 {
			beg += pre
		} else if !blankline {
			// non-empty, non-prefixed line breaks the pre
			i = beg
			break
		}

		// verbatim copy to the working buffer
		if blankline {
			work.WriteByte('\n')
		} else {
			work.Write(data[beg:i])
		}
	}
	caption := ""
	line := i
	j := i
	// In the case of F> there may be spaces in front of it
	if bytes.HasPrefix(bytes.TrimSpace(data[j:]), []byte("Figure: ")) {
		for line < len(data) {
			j++
			// find the end of this line
			for data[j-1] != '\n' {
				j++
			}
			if p.isEmpty(data[line:j]) > 0 {
				break
			}
			line = j
		}
		// save for later processing.
		caption = string(data[i+8 : j-1]) // +8 for 'Figure: '
	}

	// trim all the \n off the end of work
	workbytes := work.Bytes()
	eol := len(workbytes)
	for eol > 0 && workbytes[eol-1] == '\n' {
		eol--
	}
	if eol != len(workbytes) {
		work.Truncate(eol)
	}

	work.WriteByte('\n')

	co := ""
	if p.ial != nil {
		// enabled, any non-empty value
		co = p.ial.Value("callout")
	}

	p.r.SetInlineAttr(p.ial)
	p.ial = nil

	var capb bytes.Buffer
	if co != "" {
		var callout bytes.Buffer
		callouts(p, &callout, work.Bytes(), 0, co)
		p.inline(&capb, []byte(caption))
		p.r.BlockCode(out, callout.Bytes(), "", capb.Bytes(), p.insideFigure, true)
	} else {
		p.callouts = nil
		p.inline(&capb, []byte(caption))
		p.r.BlockCode(out, work.Bytes(), "", capb.Bytes(), p.insideFigure, false)
	}

	return j
}

// returns unordered list item prefix
func (p *parser) uliPrefix(data []byte) int {
	i := 0
	if len(data) < 3 {
		return 0
	}

	// start with up to 3 spaces
	for i < 3 && data[i] == ' ' {
		i++
	}

	// need a *, +, #, or - followed by a space
	if (data[i] != '*' && data[i] != '+' && data[i] != '-' && !iswhitespace(data[i])) ||
		!iswhitespace(data[i+1]) {
		return 0
	}
	return i + 2
}

// returns ordered list item prefix
func (p *parser) oliPrefix(data []byte) int {
	i := 0
	if len(data) < 3 {
		return 0
	}

	// start with up to 3 spaces
	for i < 3 && data[i] == ' ' {
		i++
	}

	// count the digits
	start := i
	for isnum(data[i]) {
		i++
	}

	// we need >= 1 digits followed by a dot or brace and a space
	if start == i || (data[i] != '.' && data[i] != ')') || !iswhitespace(data[i+1]) {
		return 0
	}
	return i + 2
}

// returns ordered list item prefix for alpha ordered list
func (p *parser) aliPrefix(data []byte) int {
	i := 0
	if len(data) < 4 {
		return 0
	}

	// start with up to 3 spaces
	for i < 3 && data[i] == ' ' {
		i++
	}

	// count the digits
	start := i
	for data[i] >= 'a' && data[i] <= 'z' {
		i++
	}

	// we need >= 1 letter followed by a dot and two spaces
	if start == i || (data[i] != '.' && data[i] != ')') || !iswhitespace(data[i+1]) || !iswhitespace(data[i+2]) {
		return 0
	}
	if i-start > 2 {
		// crazy list, i.e. too many letters.
		return 0
	}

	return i + 3
}

// returns ordered list item prefix for alpha uppercase ordered list
func (p *parser) aliPrefixU(data []byte) int {
	i := 0
	if len(data) < 4 {
		return 0
	}

	// start with up to 3 spaces
	for i < 3 && data[i] == ' ' {
		i++
	}

	// count the digits
	start := i
	for isupper(data[i]) {
		i++
	}

	// we need >= 1 letter followed by a dot and  two spaces
	if start == i || (data[i] != '.' && data[i] != ')') || !iswhitespace(data[i+1]) || !iswhitespace(data[i+2]) {
		return 0
	}
	if i-start > 2 {
		// crazy list, i.e. too many letters.
		return 0
	}
	return i + 3
}

// returns ordered list item prefix for roman ordered list
func (p *parser) rliPrefix(data []byte) int {
	i := 0
	if len(data) < 4 {
		return 0
	}

	// start with up to 3 spaces
	for i < 3 && data[i] == ' ' {
		i++
	}

	// count the digits
	start := i
	for isroman(data[i], false) {
		i++
	}

	// we need >= 1 letter followed by a dot and  two spaces
	if start == i || (data[i] != '.' && data[i] != ')') || !iswhitespace(data[i+1]) || !iswhitespace(data[i+2]) {
		return 0
	}
	return i + 3
}

// returns ordered list item prefix for roman uppercase ordered list
func (p *parser) rliPrefixU(data []byte) int {
	i := 0
	if len(data) < 4 {
		return 0
	}

	// start with up to 3 spaces
	for i < 3 && data[i] == ' ' {
		i++
	}

	// count the digits
	start := i
	for isroman(data[i], true) {
		i++
	}

	// we need >= 1 letter followed by a dot and  two spaces
	if start == i || (data[i] != '.' && data[i] != ')') || !iswhitespace(data[i+1]) || !iswhitespace(data[i+2]) {
		return 0
	}
	return i + 3
}

// returns definition list item prefix
func (p *parser) dliPrefix(data []byte) int {
	// return the index of where the term ends
	i := 0
	for data[i] != '\n' && i < len(data) {
		i++
	}
	if i == 0 || i == len(data) {
		return 0
	}
	// start with up to 3 spaces before :
	j := 0
	for j < 3 && iswhitespace(data[i+j]) && i+j < len(data) {
		j++
	}
	i++
	if i >= len(data) {
		return 0
	}
	if data[i] == ':' {
		return i + 1
	}
	return 0
}

// returns example list item prefix
func (p *parser) eliPrefix(data []byte) int {
	i := 0
	if len(data) < 6 {
		return 0
	}

	// start with up to 3 spaces
	for i < 3 && iswhitespace(data[i]) {
		i++
	}

	// (@<tag>)
	if data[i] != '(' || data[i+1] != '@' {
		return 0
	}

	// count up until the closing )
	for data[i] != ')' {
		i++
		if i == len(data) {
			return 0
		}
	}
	// now two spaces
	if data[i] != ')' || !iswhitespace(data[i+1]) || !iswhitespace(data[i+2]) {
		return 0
	}
	return i + 2
}

// parse ordered or unordered or definition list block
func (p *parser) list(out *bytes.Buffer, data []byte, flags, start int, group []byte) int {
	p.insideList++
	defer func() {
		p.insideList--
	}()
	i := 0
	flags |= _LIST_ITEM_BEGINNING_OF_LIST
	work := func() bool {
		for i < len(data) {
			skip := p.listItem(out, data[i:], &flags)
			i += skip

			if skip == 0 || flags&_LIST_ITEM_END_OF_LIST != 0 {
				break
			}
			flags &= ^_LIST_ITEM_BEGINNING_OF_LIST
		}
		return true
	}
	if group != nil {
		gr := string(group)
		if _, ok := p.examples[gr]; ok {
			p.examples[gr]++
		} else {
			p.examples[gr] = 1
		}
	}

	p.r.SetInlineAttr(p.ial)
	p.ial = nil

	if p.insideList > 1 {
		flags |= _LIST_INSIDE_LIST
	} else {
		flags &= ^_LIST_INSIDE_LIST
	}

	p.r.List(out, work, flags, start, group)
	return i
}

// Parse a single list item.
// Assumes initial prefix is already removed if this is a sublist.
func (p *parser) listItem(out *bytes.Buffer, data []byte, flags *int) int {
	// keep track of the indentation of the first line
	itemIndent := 0
	for itemIndent < 3 && iswhitespace(data[itemIndent]) {
		itemIndent++
	}

	i := p.uliPrefix(data)
	if i == 0 {
		i = p.oliPrefix(data)
	}
	if i == 0 {
		i = p.aliPrefix(data)
	}
	if i == 0 {
		i = p.aliPrefixU(data)
	}
	if i == 0 {
		i = p.rliPrefix(data)
	}
	if i == 0 {
		i = p.rliPrefixU(data)
	}
	if i == 0 {
		i = p.eliPrefix(data)
	}
	if i == 0 {
		i = p.dliPrefix(data)
		if i > 0 {
			var rawTerm bytes.Buffer
			p.inline(&rawTerm, data[:i-2]) // -2 for : and the newline
			p.r.ListItem(out, rawTerm.Bytes(), *flags|_LIST_TYPE_TERM)
		}
	}

	if i == 0 {
		return 0
	}

	// skip leading whitespace on first line
	for iswhitespace(data[i]) {
		i++
	}

	// find the end of the line
	line := i
	for data[i-1] != '\n' {
		i++
	}

	// get working buffer
	var raw bytes.Buffer

	// put the first line into the working buffer
	raw.Write(data[line:i])
	line = i

	// process the following lines
	containsBlankLine := false
	sublist := 0

gatherlines:
	for line < len(data) {
		i++

		// find the end of this line
		for data[i-1] != '\n' {
			i++
		}

		// if it is an empty line, guess that it is part of this item
		// and move on to the next line
		if p.isEmpty(data[line:i]) > 0 {
			containsBlankLine = true
			line = i
			continue
		}

		// calculate the indentation
		indent := 0
		for indent < 4 && line+indent < i && data[line+indent] == ' ' {
			indent++
		}

		chunk := data[line+indent : i]

		// evaluate how this line fits in
		switch {
		// is this a nested list item?
		case (p.uliPrefix(chunk) > 0 && !p.isHRule(chunk)) ||
			p.aliPrefix(chunk) > 0 || p.aliPrefixU(chunk) > 0 ||
			p.rliPrefix(chunk) > 0 || p.rliPrefixU(chunk) > 0 ||
			p.oliPrefix(chunk) > 0 || p.eliPrefix(chunk) > 0 ||
			p.dliPrefix(data[line+indent:]) > 0:

			if containsBlankLine {
				*flags |= _LIST_ITEM_CONTAINS_BLOCK
			}

			// to be a nested list, it must be indented more
			// if not, it is the next item in the same list
			if indent <= itemIndent {
				break gatherlines
			}

			// is this the first item in the the nested list?
			if sublist == 0 {
				sublist = raw.Len()
			}

		// is this a nested prefix header?
		case p.isPrefixHeader(chunk):
			// if the header is not indented, it is not nested in the list
			// and thus ends the list
			if containsBlankLine && indent < 4 {
				*flags |= _LIST_ITEM_END_OF_LIST
				break gatherlines
			}
			*flags |= _LIST_ITEM_CONTAINS_BLOCK

		// anything following an empty line is only part
		// of this item if it is indented 4 spaces
		// (regardless of the indentation of the beginning of the item)
		// if the is beginning with ':   term', we have a new term
		case containsBlankLine && indent < 4:
			*flags |= _LIST_ITEM_END_OF_LIST
			break gatherlines

		// a blank line means this should be parsed as a block
		case containsBlankLine:
			raw.WriteByte('\n')
			*flags |= _LIST_ITEM_CONTAINS_BLOCK

		// CommonMark, rule breaks the list, but when indented it belong to the list
		case p.isHRule(chunk) && indent < 4:
			*flags |= _LIST_ITEM_END_OF_LIST
			break gatherlines

		}

		// if this line was preceeded by one or more blanks,
		// re-introduce the blank into the buffer
		if containsBlankLine {
			containsBlankLine = false
			raw.WriteByte('\n')
		}

		// add the line into the working buffer without prefix
		raw.Write(data[line+indent : i])

		line = i
	}
	rawBytes := raw.Bytes()

	// render the contents of the list item
	var cooked bytes.Buffer
	if *flags&_LIST_ITEM_CONTAINS_BLOCK != 0 {
		// intermediate render of block li
		if sublist > 0 {
			p.block(&cooked, rawBytes[:sublist])
			p.block(&cooked, rawBytes[sublist:])
		} else {
			p.block(&cooked, rawBytes)
		}
	} else {
		// intermediate render of inline li
		if sublist > 0 {
			p.inline(&cooked, rawBytes[:sublist])
			p.block(&cooked, rawBytes[sublist:])
		} else {
			p.inline(&cooked, rawBytes)
		}
	}

	// render the actual list item
	cookedBytes := cooked.Bytes()
	parsedEnd := len(cookedBytes)

	// strip trailing newlines
	for parsedEnd > 0 && cookedBytes[parsedEnd-1] == '\n' {
		parsedEnd--
	}
	p.r.ListItem(out, cookedBytes[:parsedEnd], *flags)

	return line
}

// render a single paragraph that has already been parsed out
func (p *parser) renderParagraph(out *bytes.Buffer, data []byte) {
	if len(data) == 0 {
		return
	}
	// trim leading spaces
	beg := 0
	for iswhitespace(data[beg]) {
		beg++
	}

	// trim trailing newline
	end := len(data) - 1

	// trim trailing spaces
	for end > beg && iswhitespace(data[end-1]) {
		end--
	}

	if isMatter(data) {
		p.inline(out, data[beg:end])
		return
	}
	p.displayMath = false
	work := func() bool {
		// if we are a single paragraph constisting entirely out of math
		// we set the displayMath to true
		k := 0
		if end-beg > 4 && data[beg] == '$' && data[beg+1] == '$' {
			for k = beg + 2; k < end-1; k++ {
				if data[k] == '$' && data[k+1] == '$' {
					break
				}
			}
			if k+2 == end {
				p.displayMath = true
			}
		}
		p.inline(out, data[beg:end])
		return true
	}

	flags := 0
	if p.insideDefinitionList {
		flags |= _LIST_TYPE_DEFINITION
	}
	if p.insideList > 0 {
		flags |= _LIST_INSIDE_LIST // Not really, just in a list
	} else {
		flags &= ^_LIST_INSIDE_LIST // Not really, just in a list
	}
	p.r.Paragraph(out, work, flags)
}

func (p *parser) paragraph(out *bytes.Buffer, data []byte) int {
	// prev: index of 1st char of previous line
	// line: index of 1st char of current line
	// i: index of cursor/end of current line
	var prev, line, i int

	// keep going until we find something to mark the end of the paragraph
	for i < len(data) {
		// mark the beginning of the current line
		prev = line
		current := data[i:]
		line = i

		// did we find a blank line marking the end of the paragraph?
		if n := p.isEmpty(current); n > 0 {
			p.renderParagraph(out, data[:i])
			return i + n
		}

		// an underline under some text marks a header, so our paragraph ended on prev line
		if i > 0 {
			if level := p.isUnderlinedHeader(current); level > 0 {
				// render the paragraph
				p.renderParagraph(out, data[:prev])

				// ignore leading and trailing whitespace
				eol := i - 1
				for prev < eol && iswhitespace(data[prev]) {
					prev++
				}
				for eol > prev && iswhitespace(data[eol-1]) {
					eol--
				}

				// render the header
				// this ugly double closure avoids forcing variables onto the heap
				work := func(o *bytes.Buffer, pp *parser, d []byte) func() bool {
					return func() bool {
						// this renders the name, but how to make attribute out of it
						pp.inline(o, d)
						return true
					}
				}(out, p, data[prev:eol])

				id := ""
				if p.flags&EXTENSION_AUTO_HEADER_IDS != 0 {
					id = createSanitizedAnchorName(string(data[prev:eol]))
				}

				p.r.SetInlineAttr(p.ial)
				p.ial = nil

				p.r.Header(out, work, level, id)

				// find the end of the underline
				for data[i] != '\n' {
					i++
				}
				return i
			}
		}

		// if the next line starts a block of HTML, then the paragraph ends here
		if p.flags&EXTENSION_LAX_HTML_BLOCKS != 0 {
			if data[i] == '<' && p.html(out, current, false) > 0 {
				// rewind to before the HTML block
				p.renderParagraph(out, data[:i])
				return i
			}
		}

		// if there's a prefixed header or a horizontal rule after this, paragraph is over
		if p.isPrefixHeader(current) || p.isHRule(current) {
			p.renderParagraph(out, data[:i])
			return i
		}

		// if there's a list after this, paragraph is over
		if p.flags&EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK != 0 {
			if p.uliPrefix(current) != 0 ||
				p.aliPrefixU(current) != 0 ||
				p.aliPrefix(current) != 0 ||
				p.rliPrefixU(current) != 0 ||
				p.rliPrefix(current) != 0 ||
				p.oliPrefix(current) != 0 ||
				p.eliPrefix(current) != 0 ||
				p.dliPrefix(current) != 0 ||
				p.quotePrefix(current) != 0 ||
				p.notePrefix(current) != 0 ||
				p.figurePrefix(current) != 0 ||
				p.asidePrefix(current) != 0 ||
				p.codePrefix(current) != 0 {
				p.renderParagraph(out, data[:i])
				return i
			}
		}

		// otherwise, scan to the beginning of the next line
		for data[i] != '\n' {
			i++
		}
		i++
	}

	p.renderParagraph(out, data[:i])
	return i
}

func createSanitizedAnchorName(text string) string {
	var anchorName []rune
	number := 0
	for _, r := range []rune(text) {
		switch {
		case r == ' ':
			anchorName = append(anchorName, '-')
		case unicode.IsNumber(r):
			number++
			fallthrough
		case unicode.IsLetter(r):
			anchorName = append(anchorName, unicode.ToLower(r))
		}
	}
	if number == len(anchorName) {
		anchorName = append([]rune("section-"), anchorName...)
	}
	return string(anchorName)
}

func isMatter(text []byte) bool {
	if string(text) == "{frontmatter}\n" {
		return true
	}
	if string(text) == "{mainmatter}\n" {
		return true
	}
	if string(text) == "{backmatter}\n" {
		return true
	}
	return false
}
