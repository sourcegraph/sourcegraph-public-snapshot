// Functions to parse inline elements.

package mmark

import (
	"bytes"
	"regexp"
	"strconv"
	"unicode/utf8"
)

var (
	urlRe      = `((https?|ftp):\/\/|\/)[-A-Za-z0-9+&@#\/%?=~_|!:,.;\(\)]+`
	anchorRe   = regexp.MustCompile(`^(<a\shref="` + urlRe + `"(\stitle="[^"<>]+")?\s?>` + urlRe + `<\/a>)`)
	htmlEntity = regexp.MustCompile(`&[a-z]{2,5};`)
)

// Functions to parse text within a block
// Each function returns the number of chars taken care of
// data is the complete block being rendered
// offset is the number of valid chars before the current cursor

func (p *parser) inline(out *bytes.Buffer, data []byte) {
	// this is called recursively: enforce a maximum depth
	if p.nesting >= p.maxNesting {
		return
	}
	p.nesting++

	i, end := 0, 0
	for i < len(data) {
		// copy inactive chars into the output
		for end < len(data) && p.inlineCallback[data[end]] == nil {
			end++
		}

		normalText(p, out, data[i:end])

		if end >= len(data) {
			break
		}
		i = end

		// call the trigger
		handler := p.inlineCallback[data[end]]
		if consumed := handler(p, out, data, i); consumed == 0 {
			end = i + 1
		} else {
			// skip past whatever the callback used
			i += consumed
			end = i
		}
	}

	p.nesting--
}

// single and double emphasis parsing
func emphasis(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	// CommonMark opening char preceded by alphanumeric and followed
	// by punctuation: not emphasis
	if offset > 0 && offset+1 < len(data) {
		if isalnum(data[offset-1]) && ispunct(data[offset+1]) {
			// except when data[offset+1] is not '**, because it could be intraword double emphasis.
			if data[offset+1] != '*' {
				return 0
			}
		}
	}

	data = data[offset:]
	c := data[0]
	ret := 0

	if len(data) > 2 && data[1] != c {
		// whitespace cannot follow an opening emphasis;
		// strikethrough only takes two characters '~~'
		if c == '~' && isspace(data[1]) {
			return 0
		}
		// an emphasis character followed by a space is just that: a lone character
		if isspace(data[1]) {
			return 0
		}
		switch c {
		case '~':
			if ret = subscript(p, out, data[1:], 0); ret == 0 {
				return 0
			}
		default:
			if ret = helperEmphasis(p, out, data[1:], c); ret == 0 {
				return 0
			}
		}
		return ret + 1
	}

	if len(data) > 3 && data[1] == c && data[2] != c {
		if isspace(data[2]) {
			return 0
		}
		if ret = helperDoubleEmphasis(p, out, data[2:], c); ret == 0 {
			return 0
		}

		return ret + 2
	}

	if len(data) > 4 && data[1] == c && data[2] == c && data[3] != c {
		if c == '~' || isspace(data[3]) {
			return 0
		}
		if ret = helperTripleEmphasis(p, out, data, 3, c); ret == 0 {
			return 0
		}

		return ret + 3
	}

	return 0
}

func codeSpan(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	data = data[offset:]

	nb := 0

	// count the number of backticks in the delimiter
	for nb < len(data) && data[nb] == '`' {
		nb++
	}

	// find the next delimiter
	i, end := 0, 0
	for end = nb; end < len(data) && i < nb; end++ {
		if data[end] == '`' {
			i++
		} else {
			i = 0
		}
	}

	// no matching delimiter?
	if i < nb && end >= len(data) {
		return 0
	}

	// trim outside whitespace
	fBegin := nb
	for fBegin < end && isspace(data[fBegin]) {
		fBegin++
	}

	fEnd := end - nb
	for fEnd > fBegin && isspace(data[fEnd-1]) {
		fEnd--
	}

	// render the code span
	if fBegin != fEnd {
		p.r.CodeSpan(out, data[fBegin:fEnd])
	}

	return end
}

func normalText(p *parser, out *bytes.Buffer, data []byte) {
	if len(p.abbreviations) == 0 {
		p.r.NormalText(out, data)
	} else {
		end := len(data)
		wordBeg := 0
		inWord := false
		for j := 0; j < end; j++ {
			switch {
			case !isspace(data[j]) && !inWord:
				inWord = true
				wordBeg = j
			case isspace(data[j]) && inWord:
				// first space after coming out of a word, output
				if t, ok := p.abbreviations[string(data[wordBeg:j])]; ok {
					p.r.Abbreviation(out, data[wordBeg:j], t.title)
				} else {
					p.r.NormalText(out, data[wordBeg:j])
				}
				p.r.NormalText(out, data[j:j+1])
				inWord = false
			case isspace(data[j]) && !inWord:
				p.r.NormalText(out, data[j:j+1])
			}
		}
		// if inWord == true, we haven't outputted the last word
		if inWord {
			if t, ok := p.abbreviations[string(data[wordBeg:end])]; ok {
				p.r.Abbreviation(out, data[wordBeg:end], t.title)
			} else {
				p.r.NormalText(out, data[wordBeg:end])
			}
		}
	}
}

// newline preceded by two spaces becomes <br>
// newline without two spaces works when EXTENSION_HARD_LINE_BREAK is enabled
func lineBreak(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	// remove trailing spaces from out
	outBytes := out.Bytes()
	end := len(outBytes)
	eol := end
	for eol > 0 && outBytes[eol-1] == ' ' {
		eol--
	}
	out.Truncate(eol)

	if offset > 1 && data[offset-1] == '\\' {
		out.Truncate(eol - 1)
		p.r.LineBreak(out)
		return 1
	}

	precededByTwoSpaces := offset >= 2 && data[offset-2] == ' ' && data[offset-1] == ' '
	precededByBackslash := offset >= 1 && data[offset-1] == '\\' // see http://spec.commonmark.org/0.18/#example-527
	precededByBackslash = precededByBackslash && p.flags&EXTENSION_BACKSLASH_LINE_BREAK != 0

	// should there be a hard line break here?
	if p.flags&EXTENSION_HARD_LINE_BREAK == 0 && !precededByTwoSpaces && !precededByBackslash {
		return 0
	}
	if precededByBackslash && eol > 0 {
		out.Truncate(eol - 1)
	}

	p.r.LineBreak(out)
	return 1
}

type linkType int

const (
	linkNormal linkType = iota
	linkImg
	linkDeferredFootnote
	linkInlineFootnote
	linkCitation
)

// '[': parse a link or an image or a footnote
func link(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	// no links allowed inside regular links, footnote, and deferred footnotes
	if p.insideLink && (offset > 0 && data[offset-1] == '[' || len(data)-1 > offset && data[offset+1] == '^') {
		return 0
	}

	// [text] == regular link
	// ![alt] == image
	// ^[text] == inline footnote
	// [^refId] == deferred footnote
	// [@text] == citation
	// [-@text] == citation, add to reference, but suppress output
	var t linkType
	if offset > 0 && data[offset-1] == '!' {
		t = linkImg
	} else if offset > 0 && data[offset-1] == '@' {
		t = linkCitation
	} else if offset > 0 && data[offset-1] == '-' {
		t = linkCitation
	} else if p.flags&EXTENSION_FOOTNOTES != 0 {
		if offset > 0 && data[offset-1] == '^' {
			t = linkInlineFootnote
		} else if len(data)-1 > offset && data[offset+1] == '^' {
			t = linkDeferredFootnote
		}
	}

	data = data[offset:]

	var (
		i           = 1
		noteId      int
		title, link []byte
		textHasNl   = false
	)

	if t == linkDeferredFootnote {
		i++
	}

	// look for the matching closing bracket
	for level := 1; level > 0 && i < len(data); i++ {
		switch {
		case data[i] == '\n':
			textHasNl = true

		case data[i-1] == '\\':
			continue

		case data[i] == '[':
			level++

		case data[i] == ']':
			level--
			if level <= 0 {
				i-- // compensate for extra i++ in for loop
			}
		}
	}

	if i >= len(data) {
		return 0
	}
	txtE := i
	i++

	// skip any amount of whitespace or newline
	// (this is much more lax than original markdown syntax)
	for i < len(data) && isspace(data[i]) {
		i++
	}

	// TODO(miek): parse p. 23 parts here (title)
	// [@!RFC2534 p. 23], normative
	// [@?RFC2535 p. 23], informative
	// [@[!|?]draft#1 text]
	// [-@RFC] : suppress output, but add to the citation list
	if (t == linkCitation || data[1] == '@' || data[1] == '-') && p.flags&EXTENSION_CITATION != 0 {
		var (
			spaceB   int
			id       []byte
			typ      byte
			suppress bool
			seq      int = -1
		)
		typ = 'i'
		k := 1
		if data[k] == '-' {
			suppress = true
			k++
		}
		k++
		if data[k] == '!' {
			typ = 'n'
			k++
		} else if data[k] == '?' {
			typ = 'i'
			k++
		}

		for j := k; j < txtE; j++ {
			if isspace(data[j]) {
				if spaceB == 0 {
					spaceB = j
					title = data[j+1 : txtE]
					id = data[k:spaceB]
				}
			}
		}
		if spaceB == 0 {
			id = data[k:txtE]
		}

		if id == nil {
			id = data[k:txtE]
		}
		for j := 0; j < len(id); j++ {
			if id[j] == '#' {
				chunk := id[j:]
				if len(chunk) > 1 {
					num, err := strconv.Atoi(string(chunk[1:]))
					if err == nil {
						seq = num
						id = id[:j]
						break
					}
				}
			}

		}
		// we might be liberal and check which item we got and update if we see new ones.
		if c, ok := p.citations[string(id)]; !ok {
			p.citations[string(id)] = &citation{link: id, title: title, typ: typ, seq: seq}
		} else {
			if c.link != nil {
				c.link = id
			}
			if c.title != nil {
				c.title = title
			}
			if c.seq == -1 && seq != -1 {
				c.seq = seq
			}
			if c.typ == 0 {
				c.typ = typ
			}
		}

		if !suppress {
			p.r.Citation(out, id, title)
		}
		return txtE + 1
	}

	// inline style link
	switch {
	case i < len(data) && data[i] == '(':
		// skip initial whitespace
		i++

		for i < len(data) && isspace(data[i]) {
			i++
		}

		linkB := i

		// look for link end: ' " ), check for new openning
		// braces and take this into account, this may lead
		// for overshooting and probably will require some
		// finetuning.
		brace := 0
	findlinkend:
		for i < len(data) {
			switch {
			case data[i] == '\\':
				i += 2

			case data[i] == '(':
				brace++
				i++

			case data[i] == ')':
				if brace <= 0 {
					break findlinkend
				}
				brace--
				i++

			case data[i] == '\'' || data[i] == '"':
				break findlinkend

			default:
				i++
			}
		}

		if i >= len(data) || brace > 0 {
			return 0
		}
		linkE := i

		// look for title end if present
		titleB, titleE := 0, 0
		if data[i] == '\'' || data[i] == '"' {
			i++
			titleB = i

		findtitleend:
			for i < len(data) {
				switch {
				case data[i] == '\\':
					i += 2

				case data[i] == ')':
					break findtitleend

				default:
					i++
				}
			}

			if i >= len(data) {
				return 0
			}

			// skip whitespace after title
			titleE = i - 1
			for titleE > titleB && isspace(data[titleE]) {
				titleE--
			}

			// check for closing quote presence
			if data[titleE] != '\'' && data[titleE] != '"' {
				titleB, titleE = 0, 0
				linkE = i
			}
		}

		// remove whitespace at the end of the link
		for linkE > linkB && isspace(data[linkE-1]) {
			linkE--
		}

		// remove optional angle brackets around the link
		if data[linkB] == '<' {
			linkB++
		}
		if data[linkE-1] == '>' {
			linkE--
		}

		// build escaped link and title
		if linkE > linkB {
			link = data[linkB:linkE]
		}

		if titleE > titleB {
			title = data[titleB:titleE]
		}

		i++

	// reference style link
	case i < len(data)-1 && data[i] == '[' && data[i+1] != '^':
		var id []byte

		// look for the id
		i++
		linkB := i
		for i < len(data) && data[i] != ']' {
			i++
		}
		if i >= len(data) {
			return 0
		}
		linkE := i

		// find the reference
		if linkB == linkE {
			if textHasNl {
				var b bytes.Buffer

				for j := 1; j < txtE; j++ {
					switch {
					case data[j] != '\n':
						b.WriteByte(data[j])
					case !isspace(data[j-1]):
						b.WriteByte(' ')
					}
				}

				id = b.Bytes()
			} else {
				id = data[1:txtE]
			}
		} else {
			id = data[linkB:linkE]
		}

		// find the reference with matching id (ids are case-insensitive)
		key := string(bytes.ToLower(id))
		lr, ok := p.refs[key]
		if !ok {
			return 0
		}

		// keep link and title from reference
		link = lr.link
		title = lr.title
		i++

	// shortcut reference style link or reference or inline footnote
	default:
		var id []byte

		// craft the id
		if textHasNl {
			var b bytes.Buffer

			for j := 1; j < txtE; j++ {
				switch {
				case data[j] != '\n':
					b.WriteByte(data[j])
				case !isspace(data[j-1]):
					b.WriteByte(' ')
				}
			}

			id = b.Bytes()
		} else {
			if t == linkDeferredFootnote {
				id = data[2:txtE] // get rid of the ^
			} else {
				id = data[1:txtE]
			}
		}

		key := string(bytes.ToLower(id))
		if t == linkInlineFootnote {
			// create a new reference
			noteId = len(p.notes) + 1

			var fragment []byte
			if len(id) > 0 {
				if len(id) < 16 {
					fragment = make([]byte, len(id))
				} else {
					fragment = make([]byte, 16)
				}
				copy(fragment, slugify(id))
			} else {
				fragment = append([]byte("footnote-"), []byte(strconv.Itoa(noteId))...)
			}

			ref := &reference{
				noteId:   noteId,
				hasBlock: false,
				link:     fragment,
				title:    id,
			}

			p.notes = append(p.notes, ref)

			link = ref.link
			title = ref.title
		} else {
			// find the reference with matching id
			lr, ok := p.refs[key]
			if !ok {
				return 0
			}

			if t == linkDeferredFootnote {
				lr.noteId = len(p.notes) + 1
				p.notes = append(p.notes, lr)
			}

			// keep link and title from reference
			link = lr.link
			// if inline footnote, title == footnote contents
			title = lr.title
			noteId = lr.noteId
		}

		// rewind the whitespace
		i = txtE + 1
	}

	// build content: img alt is escaped, link content is parsed
	var content bytes.Buffer
	if txtE > 1 {
		if t == linkImg {
			content.Write(data[1:txtE])
		} else {
			// links cannot contain other links, so turn off link parsing temporarily
			insideLink := p.insideLink
			p.insideLink = true
			p.inline(&content, data[1:txtE])
			p.insideLink = insideLink
		}
	}

	var uLink []byte
	if t == linkNormal || t == linkImg {
		if len(link) > 0 {
			var uLinkBuf bytes.Buffer
			unescapeText(&uLinkBuf, link)
			uLink = uLinkBuf.Bytes()
		}

		// links need something to click on and somewhere to go
		//if len(uLink) == 0 || (t == linkNormal && content.Len() == 0) {
		if len(uLink) == 0 {
			return 0
		}
	}

	// call the relevant rendering function
	switch t {
	case linkNormal:
		p.r.Link(out, uLink, title, content.Bytes())

	case linkImg:
		outSize := out.Len()
		outBytes := out.Bytes()
		if outSize > 0 && outBytes[outSize-1] == '!' {
			out.Truncate(outSize - 1)
		}

		var cooked bytes.Buffer
		p.inline(&cooked, title)

		p.r.SetInlineAttr(p.ial)
		p.ial = nil
		p.r.Image(out, uLink, cooked.Bytes(), content.Bytes(), p.insideFigure)

	case linkInlineFootnote:
		outSize := out.Len()
		outBytes := out.Bytes()
		if outSize > 0 && outBytes[outSize-1] == '^' {
			out.Truncate(outSize - 1)
		}

		p.r.FootnoteRef(out, link, noteId)

	case linkDeferredFootnote:
		p.r.FootnoteRef(out, link, noteId)

	default:
		return 0
	}

	return i
}

// '<' when tags or autolinks are allowed
func leftAngle(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	if p.flags&EXTENSION_INCLUDE != 0 {
		if j := p.codeInclude(out, data[offset:]); j > 0 {
			return j
		}
	}

	data = data[offset:]
	altype := _LINK_TYPE_NOT_AUTOLINK
	end := tagLength(data, &altype)

	if end > 2 {
		allnum := 0
		for j := 1; j < end-1; j++ {
			if isnum(data[j]) {
				allnum++
			}
		}
		if allnum+2 == end {
			index := string(data[1 : end-1])
			if _, ok := p.callouts[index]; ok {
				p.r.CalloutText(out, index, p.callouts[index])
				return end
			} else {
				return 0
			}
		}

		if altype != _LINK_TYPE_NOT_AUTOLINK {
			var uLink bytes.Buffer
			unescapeText(&uLink, data[1:end+1-2])
			if uLink.Len() > 0 {
				p.r.AutoLink(out, uLink.Bytes(), altype)
			}
		} else {
			p.r.RawHtmlTag(out, data[:end])
		}
	}

	return end
}

// '<' for callouts in code.
func callouts(p *parser, out *bytes.Buffer, data []byte, offset int, comment string) {
	p.codeBlock++
	p.callouts = make(map[string][]string)
	i := offset
	j := 0
	if comment != ";" && comment != "#" && comment != "//" {
		comment = ""
	}

	for i < len(data) {
		if data[i] == '\\' && i < len(data)-1 && data[i+1] == '<' {
			// skip \\
			out.WriteByte(data[i])
			i++
			continue
		}
		switch comment {
		case "#":
			if data[i] == '#' {
				if i+1 > len(data) {
					out.WriteByte(data[i])
					return
				}
				i++
			}
			if data[i] == '<' && i > 0 && data[i-1] != '\\' {
				if x := leftAngleCode(data[i:]); x > 0 {
					j++
					index := string(data[i+1 : i+x])
					p.callouts[index] = append(p.callouts[index], strconv.Itoa(j))
				}
			}
		case ";":
			if data[i] == ';' {
				if i+1 > len(data) {
					out.WriteByte(data[i])
					return
				}
				i++
			}
			if data[i] == '<' && i > 0 && data[i-1] != '\\' {
				if x := leftAngleCode(data[i:]); x > 0 {
					j++
					index := string(data[i+1 : i+x])
					p.callouts[index] = append(p.callouts[index], strconv.Itoa(j))
				}
			}
		case "//":
			if data[i] == '/' && i < len(data) && data[i+1] == '/' {
				if i+2 > len(data) {
					out.WriteByte(data[i])
					out.WriteByte(data[i+1])
					return
				}
				i += 2
			}
			if data[i] == '<' && i > 0 && data[i-1] != '\\' {
				if x := leftAngleCode(data[i:]); x > 0 {
					j++
					index := string(data[i+1 : i+x])
					p.callouts[index] = append(p.callouts[index], strconv.Itoa(j))
				}
			}
		case "":
			if data[i] == '<' && i > 0 && data[i-1] != '\\' {
				if x := leftAngleCode(data[i:]); x > 0 {
					j++
					index := string(data[i+1 : i+x])
					p.callouts[index] = append(p.callouts[index], strconv.Itoa(j))
				}
			}
		}

		out.WriteByte(data[i])
		i++
	}
	return
}

// return > 0 if <xxx> if found where xxx is a number > 0
// should be called when on a '<'
func leftAngleCode(data []byte) int {
	i := 1
	for i < len(data) {
		if data[i] == '>' {
			break
		}
		if !isnum(data[i]) {
			return 0
		}
		i++
	}
	return i
}

// '{' IAL or *matter, {{ is handled in the first pass
func leftBrace(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	data = data[offset:]
	if offset == 0 {
		// {*matter} are only valid at the beginning of the line
		switch s := string(data); true {
		case s == "{frontmatter}":
			p.r.DocumentMatter(out, _DOC_FRONT_MATTER)
			return len(data) + 1
		case s == "{mainmatter}":
			p.r.DocumentMatter(out, _DOC_MAIN_MATTER)
			return len(data) + 1
		case s == "{backmatter}":
			p.r.DocumentMatter(out, _DOC_BACK_MATTER)
			p.r.References(out, p.citations)
			p.appendix = true
			return len(data) + 1
		}
	}
	if j := p.isInlineAttr(data); j > 0 {
		return j
	}
	return 0
}

// '\\' backslash escape
var escapeChars = []byte("\\`*_{}[]()#+-.!:|&<>~^")

func escape(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	data = data[offset:]

	if len(data) > 1 {
		if bytes.IndexByte(escapeChars, data[1]) < 0 {
			return 0
		}

		p.r.NormalText(out, data[1:2])
	}

	return 2
}

func unescapeText(ob *bytes.Buffer, src []byte) {
	i := 0
	for i < len(src) {
		org := i
		for i < len(src) && src[i] != '\\' {
			i++
		}

		if i > org {
			ob.Write(src[org:i])
		}

		if i+1 >= len(src) {
			break
		}

		ob.WriteByte(src[i+1])
		i += 2
	}
}

// '&' escaped when it doesn't belong to an entity
// valid entities are assumed to be anything matching &#?[A-Za-z0-9]+;
func entity(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	data = data[offset:]

	end := 1

	if end < len(data) && data[end] == '#' {
		end++
	}

	for end < len(data) && isalnum(data[end]) {
		end++
	}

	if end < len(data) && data[end] == ';' {
		end++ // real entity
		p.r.Entity(out, decodeEntity(data[:end]))
		return end
	}

	return 0 // lone '&'
}

// decode decimal entity to the UTF8 code point
// we receive the whole entity: &#10232;
func decodeEntity(entity []byte) []byte {
	base := 10
	i := 2
	if entity[2] == 'x' || entity[2] == 'X' {
		i++
		base = 16
	}
	r, e := strconv.ParseInt(string(entity[i:len(entity)-1]), base, 32)
	if e != nil {
		return entity
	}

	l := utf8.RuneLen(rune(r))
	if l == -1 {
		return []byte("0xFFFD")
	}
	u := make([]byte, l)
	l = utf8.EncodeRune(u, rune(r))
	u = u[:l]

	return u
}

func linkEndsWithEntity(data []byte, linkEnd int) bool {
	entityRanges := htmlEntity.FindAllIndex(data[:linkEnd], -1)
	if entityRanges != nil && entityRanges[len(entityRanges)-1][1] == linkEnd {
		return true
	}
	return false
}

func autoLink(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	// quick check to rule out most false hits on ':'
	if p.insideLink || len(data) < offset+3 || data[offset+1] != '/' || data[offset+2] != '/' {
		return 0
	}

	// Now a more expensive check to see if we're not inside an anchor element
	anchorStart := offset
	offsetFromAnchor := 0
	for anchorStart > 0 && data[anchorStart] != '<' {
		anchorStart--
		offsetFromAnchor++
	}

	anchorStr := anchorRe.Find(data[anchorStart:])
	if anchorStr != nil {
		out.Write(anchorStr[offsetFromAnchor:])
		return len(anchorStr) - offsetFromAnchor
	}

	// scan backward for a word boundary
	rewind := 0
	for offset-rewind > 0 && rewind <= 7 && isletter(data[offset-rewind-1]) {
		rewind++
	}
	if rewind > 6 { // longest supported protocol is "mailto" which has 6 letters
		return 0
	}

	origData := data
	data = data[offset-rewind:]

	if !isSafeLink(data) {
		return 0
	}

	linkEnd := 0
	for linkEnd < len(data) && !isEndOfLink(data[linkEnd]) {
		linkEnd++
	}

	// Skip punctuation at the end of the link
	if (data[linkEnd-1] == '.' || data[linkEnd-1] == ',') && data[linkEnd-2] != '\\' {
		linkEnd--
	}

	// But don't skip semicolon if it's a part of escaped entity:
	if data[linkEnd-1] == ';' && data[linkEnd-2] != '\\' && !linkEndsWithEntity(data, linkEnd) {
		linkEnd--
	}

	// See if the link finishes with a punctuation sign that can be closed.
	var copen byte
	switch data[linkEnd-1] {
	case '"':
		copen = '"'
	case '\'':
		copen = '\''
	case ')':
		copen = '('
	case ']':
		copen = '['
	case '}':
		copen = '{'
	default:
		copen = 0
	}

	if copen != 0 {
		bufEnd := offset - rewind + linkEnd - 2

		openDelim := 1

		// Try to close the final punctuation sign in this same line;
		// if we managed to close it outside of the URL, that means that it's
		// not part of the URL. If it closes inside the URL, that means it
		// is part of the URL.

		//	 Examples:
		//
		//	      foo http://www.pokemon.com/Pikachu_(Electric) bar
		//	              => http://www.pokemon.com/Pikachu_(Electric)
		//
		//	      foo (http://www.pokemon.com/Pikachu_(Electric)) bar
		//	              => http://www.pokemon.com/Pikachu_(Electric)
		//
		//	      foo http://www.pokemon.com/Pikachu_(Electric)) bar
		//	              => http://www.pokemon.com/Pikachu_(Electric))
		//
		//	      (foo http://www.pokemon.com/Pikachu_(Electric)) bar
		//	              => foo http://www.pokemon.com/Pikachu_(Electric)
		//

		for bufEnd >= 0 && origData[bufEnd] != '\n' && openDelim != 0 {
			if origData[bufEnd] == data[linkEnd-1] {
				openDelim++
			}

			if origData[bufEnd] == copen {
				openDelim--
			}

			bufEnd--
		}

		if openDelim == 0 {
			linkEnd--
		}
	}

	// we were triggered on the ':', so we need to rewind the output a bit
	if out.Len() >= rewind {
		out.Truncate(len(out.Bytes()) - rewind)
	}

	var uLink bytes.Buffer
	unescapeText(&uLink, data[:linkEnd])

	if uLink.Len() > 0 {
		p.r.AutoLink(out, uLink.Bytes(), _LINK_TYPE_NORMAL)
	}

	return linkEnd - rewind
}

func isEndOfLink(char byte) bool {
	return isspace(char) || char == '<'
}

var validUris = [][]byte{[]byte("http://"), []byte("https://"), []byte("ftp://"), []byte("mailto://"), []byte("/")}

func isSafeLink(link []byte) bool {
	for _, prefix := range validUris {
		// TODO: handle unicode here
		// case-insensitive prefix test
		if len(link) > len(prefix) && bytes.Equal(bytes.ToLower(link[:len(prefix)]), prefix) && isalnum(link[len(prefix)]) {
			return true
		}
	}

	return false
}

// return the length of the given tag, or 0 is it's not valid
func tagLength(data []byte, autolink *int) int {
	var i, j int

	// a valid tag can't be shorter than 3 chars
	if len(data) < 3 {
		return 0
	}

	// begins with a '<' optionally followed by '/', followed by letter or number
	if data[0] != '<' {
		return 0
	}
	if data[1] == '/' {
		i = 2
	} else {
		i = 1
	}

	if !isalnum(data[i]) {
		return 0
	}

	// scheme test
	*autolink = _LINK_TYPE_NOT_AUTOLINK

	// try to find the beginning of an URI
	for i < len(data) && (isalnum(data[i]) || data[i] == '.' || data[i] == '+' || data[i] == '-') {
		i++
	}

	if i > 1 && i < len(data) && data[i] == '@' {
		if j = isMailtoAutoLink(data[i:]); j != 0 {
			*autolink = _LINK_TYPE_EMAIL
			return i + j
		}
	}

	if i > 2 && i < len(data) && data[i] == ':' {
		*autolink = _LINK_TYPE_NORMAL
		i++
	}

	// complete autolink test: no whitespace or ' or "
	switch {
	case i >= len(data):
		*autolink = _LINK_TYPE_NOT_AUTOLINK
	case *autolink != 0:
		j = i

		for i < len(data) {
			if data[i] == '\\' {
				i += 2
			} else if data[i] == '>' || data[i] == '\'' || data[i] == '"' || isspace(data[i]) {
				break
			} else {
				i++
			}

		}

		if i >= len(data) {
			return 0
		}
		if i > j && data[i] == '>' {
			return i + 1
		}

		// one of the forbidden chars has been found
		*autolink = _LINK_TYPE_NOT_AUTOLINK
	}

	// look for something looking like a tag end
	for i < len(data) && data[i] != '>' {
		i++
	}
	if i >= len(data) {
		return 0
	}
	return i + 1
}

// look for the address part of a mail autolink and '>'
// this is less strict than the original markdown e-mail address matching
func isMailtoAutoLink(data []byte) int {
	nb := 0

	// address is assumed to be: [-@._a-zA-Z0-9]+ with exactly one '@'
	for i := 0; i < len(data); i++ {
		if isalnum(data[i]) {
			continue
		}

		switch data[i] {
		case '@':
			nb++

		case '-', '.', '_':
			break

		case '>':
			if nb == 1 {
				return i + 1
			} else {
				return 0
			}
		default:
			return 0
		}
	}

	return 0
}

func index(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	data = data[offset:]
	c := data[0]
	ret := 0
	if len(data) > 3 && data[1] != c && data[2] != c {
		// might be example list reference
		if data[1] == '@' {
			ret = exampleReference(p, out, data, 0)
			if ret > 0 {
				return ret
			}
		}
		if data[1] == '#' && p.flags&EXTENSION_SHORT_REF != 0 {
			ret = crossReference(p, out, data, 0)
			if ret > 0 {
				return ret
			}
		}
		// no three (((
		return 0
	}
	// find closing delimeter, count commas while at it
	// if more than 1 is found, it is a proper index.
	i, end := 0, 0
	comma := 0
	for end = 3; end < len(data) && i < 3; end++ {
		if data[end] == ')' {
			i++
		} else {
			i = 0
		}
		if data[end] == ',' {
			if comma != 0 {
				// already seen comma
				return 0
			}
			comma = end
		}
	}
	if comma == 3 { // just (((,
		return 0
	}
	if i < 3 && end >= len(data) {
		return 0
	}
	ret = end
	if comma == 0 {
		comma = end
	}

	// may be surrounded by whitespace, strip it
	primary := comma - 1
	for i := comma - 1; i >= 0; i-- {
		if !isspace(data[i]) {
			break
		}
		primary = i
	}

	secondary := comma + 1
	for i := comma + 1; i < end; i++ {
		if !isspace(data[i]) {
			secondary = i
			break
		}
	}

	i = 3
	prim := false
	if data[i] == '!' {
		// mark is primary
		prim = true
		i++
	}

	if secondary > end-3 {
		p.r.Index(out, data[i:primary-2], nil, prim)
		return ret
	}
	p.r.Index(out, data[i:primary+1], data[secondary:end-3], prim)
	return ret
}

// look for the next emph char, skipping other constructs
func helperFindEmphChar(data []byte, c byte) int {
	i := 1

	for i < len(data) {
		for i < len(data) && data[i] != c && data[i] != '`' && data[i] != '[' {
			i++
		}
		if i >= len(data) {
			return 0
		}
		if data[i] == c {
			return i
		}

		// do not count escaped chars
		if i != 0 && data[i-1] == '\\' {
			i++
			continue
		}

		if data[i] == '`' {
			// skip a code span
			tmpI := 0
			i++
			for i < len(data) && data[i] != '`' {
				if tmpI == 0 && data[i] == c {
					tmpI = i
				}
				i++
			}
			if i >= len(data) {
				return tmpI
			}
			i++
		} else if data[i] == '[' {
			// skip a link
			tmpI := 0
			i++
			for i < len(data) && data[i] != ']' {
				if tmpI == 0 && data[i] == c {
					tmpI = i
				}
				i++
			}
			i++
			for i < len(data) && isspace(data[i]) {
				i++
			}
			if i >= len(data) {
				return tmpI
			}
			if data[i] != '[' && data[i] != '(' { // not a link
				if tmpI > 0 {
					return tmpI
				} else {
					continue
				}
			}
			cc := data[i]
			i++
			for i < len(data) && data[i] != cc {
				if tmpI == 0 && data[i] == c {
					return i
				}
				i++
			}
			if i >= len(data) {
				return tmpI
			}
			i++
		}
	}
	return 0
}

func helperEmphasis(p *parser, out *bytes.Buffer, data []byte, c byte) int {
	i := 0

	// skip one symbol if coming from emph3
	if len(data) > 1 && data[0] == c && data[1] == c {
		i = 1
	}

	for i < len(data) {
		length := helperFindEmphChar(data[i:], c)
		if length == 0 {
			return 0
		}
		i += length
		if i >= len(data) {
			return 0
		}

		if i+1 < len(data) && data[i+1] == c {
			i++
			continue
		}

		if data[i] == c && !isspace(data[i-1]) {
			if c != '*' {
				if !(i+1 == len(data) || isspace(data[i+1]) || ispunct(data[i+1])) {
					continue
				}
			}

			var work bytes.Buffer
			p.inline(&work, data[:i])
			p.r.Emphasis(out, work.Bytes())
			return i + 1
		}
	}

	return 0
}

func helperDoubleEmphasis(p *parser, out *bytes.Buffer, data []byte, c byte) int {
	i := 0

	for i < len(data) {
		length := helperFindEmphChar(data[i:], c)
		if length == 0 {
			return 0
		}
		i += length

		if i+1 < len(data) && data[i] == c && data[i+1] == c && i > 0 && !isspace(data[i-1]) {
			var work bytes.Buffer
			p.inline(&work, data[:i])

			if work.Len() > 0 {
				// pick the right renderer
				if c == '~' {
					p.r.StrikeThrough(out, work.Bytes())
				} else {
					p.r.DoubleEmphasis(out, work.Bytes())
				}
			}
			return i + 2
		}
		i++
	}
	return 0
}

func helperTripleEmphasis(p *parser, out *bytes.Buffer, data []byte, offset int, c byte) int {
	i := 0
	origData := data
	data = data[offset:]

	for i < len(data) {
		length := helperFindEmphChar(data[i:], c)
		if length == 0 {
			return 0
		}
		i += length

		// skip whitespace preceded symbols
		if data[i] != c || isspace(data[i-1]) {
			continue
		}

		switch {
		case i+2 < len(data) && data[i+1] == c && data[i+2] == c:
			// triple symbol found
			var work bytes.Buffer

			p.inline(&work, data[:i])
			if work.Len() > 0 {
				p.r.TripleEmphasis(out, work.Bytes())
			}
			return i + 3
		case (i+1 < len(data) && data[i+1] == c):
			// double symbol found, hand over to emph1
			length = helperEmphasis(p, out, origData[offset-2:], c)
			if length == 0 {
				return 0
			} else {
				return length - 2
			}
		default:
			// single symbol found, hand over to emph2
			length = helperDoubleEmphasis(p, out, origData[offset-1:], c)
			if length == 0 {
				return 0
			} else {
				return length - 1
			}
		}
	}
	return 0
}

func subscript(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	ret := helperScript(p, out, data[offset:], '~')
	return ret
}

func superscript(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	// superscript is called from the inline call, so we are positioned
	// on the '^'
	ret := helperScript(p, out, data[offset+1:], '^')
	return ret
}

func helperScript(p *parser, out *bytes.Buffer, data []byte, c byte) int {
	i := 0
	// write too much
	var raw bytes.Buffer
	for i < len(data) {
		if isspace(data[i]) {
			if i > 0 && data[i-1] == '\\' {
				// just written the '\', truncate to length-1 and write the space
				raw.Truncate(raw.Len() - 1)
				raw.WriteByte(data[i])
				i++
				continue
			}
			return 0
		}
		if data[i] == c {
			var work bytes.Buffer
			p.inline(&work, raw.Bytes())
			switch c {
			case '~':
				p.r.Subscript(out, work.Bytes())
				return i + 1 // differences in how subscript is called ...
			case '^':
				p.r.Superscript(out, work.Bytes())
				return i + 2 // ... compated to superscript
			}
		}
		raw.WriteByte(data[i])
		i++
	}
	return 0
}

// (@r), ref is alfanumeric, underscores or hyphens
func exampleReference(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	data = data[offset:]
	i := 0
	if len(data) < 4 {
		return 0
	}
	i++
	if data[i] != '@' {
		return 0
	}
	i++
	for i < len(data) && data[i] != ')' {
		if isalnum(data[i]) {
			i++
			continue
		}
		if data[i] == '_' || data[i] == '-' {
			i++
			continue
		}
		return 0
	}
	if e, ok := p.examples[string(data[2:i])]; ok {
		p.r.Example(out, e)
		return i + 1
	}
	return 0
}

// (#r), ref is alfanumeric, underscores or hyphens
func crossReference(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	data = data[offset:]
	i := 0
	if len(data) < 4 {
		return 0
	}
	i++
	if data[i] != '#' {
		return 0
	}
	i++
	for i < len(data) && data[i] != ')' {
		if isalnum(data[i]) {
			i++
			continue
		}
		if data[i] == '_' || data[i] == '-' || data[i] == ':' {
			i++
			continue
		}
		return 0
	}
	p.r.Link(out, data[1:i], nil, nil)
	return i + 1
}

// @r, ref is known reference anchor: alfanumeric, underscores or hyphens
func citationReference(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	data = data[offset:]
	i := 1
	for i < len(data) && !isspace(data[i]) && !ispunct(data[i]) {
		if isalnum(data[i]) {
			i++
			continue
		}
		if data[i] == '_' || data[i] == '-' {
			i++
			continue
		}
		return 0
	}
	if c, ok := p.citations[string(data[1:i])]; ok {
		p.r.Citation(out, data[1:i], c.title)
		return i
	}
	return 0
}

func math(p *parser, out *bytes.Buffer, data []byte, offset int) int {
	if len(data[offset:]) < 5 {
		return 0
	}
	i := offset + 1
	if data[i] != '$' {
		return 0
	}

	// find end delimiter
	end, j := i+1, 0
	for ; end < len(data) && j < 2; end++ {
		if data[end] == '$' {
			j++
		} else {
			j = 0
		}
	}

	// no matching delimiter?
	if j < 2 && end >= len(data) {
		return 0
	}
	if p.displayMath {
		p.r.SetInlineAttr(p.ial)
		p.ial = nil
	}
	p.r.Math(out, data[i+1:end-2], p.displayMath)
	return end - offset
}
