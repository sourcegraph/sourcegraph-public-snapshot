// Markdown parsing and processing

// Mmark markdown processor.
//
// Translates plain text with simple formatting rules into HTML or XML.
package mmark

import (
	"bytes"
	"io/ioutil"
	"log"
	"path"
	"unicode"
	"unicode/utf8"
)

const VERSION = "1.0"

// These are the supported markdown parsing extensions.
// OR these values together to select multiple extensions.
const (
	_                                    = 1 << iota
	EXTENSION_ABBREVIATIONS              // render abbreviations `*[HTML]: Hyper Text Markup Language`
	EXTENSION_AUTO_HEADER_IDS            // Create the header ID from the text
	EXTENSION_AUTOLINK                   // detect embedded URLs that are not explicitly marked
	EXTENSION_CITATION                   // Support citations via the link syntax
	EXTENSION_EXAMPLE_LISTS              // render '(@tag)  ' example lists
	EXTENSION_FENCED_CODE                // render fenced code blocks
	EXTENSION_FOOTNOTES                  // Pandoc-style footnotes
	EXTENSION_HARD_LINE_BREAK            // translate newlines into line breaks
	EXTENSION_HEADER_IDS                 // specify header IDs with {#id}
	EXTENSION_INCLUDE                    // Include file with {{ syntax
	EXTENSION_INLINE_ATTR                // detect CommonMark's IAL syntax
	EXTENSION_LAX_HTML_BLOCKS            // loosen up HTML block parsing rules
	EXTENSION_MATH                       // detect $$...$$ and parse as math
	EXTENSION_MATTER                     // use {frontmatter} {mainmatter} {backmatter}
	EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK // No need to insert an empty line to start a (code, quote, order list, unorder list)block
	EXTENSION_PARTS                      // detect part headers (-#)
	EXTENSION_QUOTES                     // Allow A> AS> and N> to be parsed as abstract, asides and notes
	EXTENSION_SHORT_REF                  // (#id) will be a cross reference.
	EXTENSION_SPACE_HEADERS              // be strict about prefix header rules
	EXTENSION_TABLES                     // render tables
	EXTENSION_TITLEBLOCK_TOML            // Titleblock in TOML
	EXTENSION_UNIQUE_HEADER_IDS          // When detecting identical anchors add a sequence number -1, -2 etc.
	EXTENSION_BACKSLASH_LINE_BREAK       // translate trailing backslashes into line breaks

	commonHtmlFlags = 0

	commonExtensions = 0 |
		EXTENSION_TABLES |
		EXTENSION_FENCED_CODE |
		EXTENSION_AUTOLINK |
		EXTENSION_SPACE_HEADERS |
		EXTENSION_HEADER_IDS |
		EXTENSION_ABBREVIATIONS |
		EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK | // CommonMark
		EXTENSION_BACKSLASH_LINE_BREAK // CommonMark

	commonXmlExtensions = commonExtensions |
		EXTENSION_UNIQUE_HEADER_IDS |
		EXTENSION_AUTO_HEADER_IDS |
		EXTENSION_INLINE_ATTR |
		EXTENSION_QUOTES |
		EXTENSION_MATTER |
		EXTENSION_CITATION |
		EXTENSION_EXAMPLE_LISTS |
		EXTENSION_SHORT_REF
)

// These are the possible flag values for the link renderer.
// Only a single one of these values will be used; they are not ORed together.
// These are mostly of interest if you are writing a new output format.
const (
	_LINK_TYPE_NOT_AUTOLINK = iota
	_LINK_TYPE_NORMAL
	_LINK_TYPE_EMAIL
)

// These are the possible flag values for the ListItem renderer.
// Multiple flag values may be ORed together.
// These are mostly of interest if you are writing a new output format.
const (
	_LIST_TYPE_ORDERED = 1 << iota
	_LIST_TYPE_ORDERED_ROMAN_UPPER
	_LIST_TYPE_ORDERED_ROMAN_LOWER
	_LIST_TYPE_ORDERED_ALPHA_UPPER
	_LIST_TYPE_ORDERED_ALPHA_LOWER
	_LIST_TYPE_ORDERED_GROUP
	_LIST_TYPE_DEFINITION
	_LIST_TYPE_TERM
	_LIST_ITEM_CONTAINS_BLOCK
	_LIST_ITEM_BEGINNING_OF_LIST
	_LIST_ITEM_END_OF_LIST
	_LIST_INSIDE_LIST
	_INSIDE_FIGURE
)

// These are the possible flag values for the table cell renderer.
// Only a single one of these values will be used; they are not ORed together.
// These are mostly of interest if you are writing a new output format.
const (
	_TABLE_ALIGNMENT_LEFT = 1 << iota
	_TABLE_ALIGNMENT_RIGHT
	_TABLE_ALIGNMENT_CENTER = (_TABLE_ALIGNMENT_LEFT | _TABLE_ALIGNMENT_RIGHT)
)

// The size of a tab stop.
const _TAB_SIZE_DEFAULT = 4

const (
	_DOC_FRONT_MATTER = iota // Different divisions of the document
	_DOC_MAIN_MATTER
	_DOC_BACK_MATTER
	_ABSTRACT // Special headers, keep track if there are open
	_PREFACE
	_COLOPHON
)

// These are the tags that are recognized as HTML block tags.
// Any of these can be included in markdown text without special escaping.
var blockTags = map[string]bool{
	"p":          true,
	"dl":         true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"ol":         true,
	"ul":         true,
	"del":        true,
	"div":        true,
	"ins":        true,
	"pre":        true,
	"form":       true,
	"math":       true,
	"table":      true,
	"iframe":     true,
	"script":     true,
	"fieldset":   true,
	"noscript":   true,
	"blockquote": true,

	// HTML5
	"video":      true,
	"aside":      true,
	"canvas":     true,
	"figure":     true,
	"footer":     true,
	"header":     true,
	"hgroup":     true,
	"output":     true,
	"article":    true,
	"section":    true,
	"progress":   true,
	"figcaption": true,
}

// Renderer is the rendering interface.
// This is mostly of interest if you are implementing a new rendering format.
//
// When a byte slice is provided, it contains the (rendered) contents of the
// element.
//
// When a callback is provided instead, it will write the contents of the
// respective element directly to the output buffer and return true on success.
// If the callback returns false, the rendering function should reset the
// output buffer as though it had never been called.
//
// Currently Html, XML2RFCv3 and XML2RFC v2 implementations are provided.
type Renderer interface {
	// block-level callbacks
	BlockCode(out *bytes.Buffer, text []byte, lang string, caption []byte, subfigure bool, callouts bool)
	BlockQuote(out *bytes.Buffer, text []byte, attribution []byte)
	BlockHtml(out *bytes.Buffer, text []byte)
	CommentHtml(out *bytes.Buffer, text []byte)
	// SpecialHeader is used for Abstract and Preface. The what string contains abstract or preface.
	SpecialHeader(out *bytes.Buffer, what []byte, text func() bool, id string)
	Part(out *bytes.Buffer, text func() bool, id string)
	Header(out *bytes.Buffer, text func() bool, level int, id string)
	HRule(out *bytes.Buffer)
	List(out *bytes.Buffer, text func() bool, flags, start int, group []byte)
	ListItem(out *bytes.Buffer, text []byte, flags int)
	Paragraph(out *bytes.Buffer, text func() bool, flags int)
	Table(out *bytes.Buffer, header []byte, body []byte, footer []byte, columnData []int, caption []byte)
	TableRow(out *bytes.Buffer, text []byte)
	TableHeaderCell(out *bytes.Buffer, text []byte, flags int)
	TableCell(out *bytes.Buffer, text []byte, flags int)
	Footnotes(out *bytes.Buffer, text func() bool)
	FootnoteItem(out *bytes.Buffer, name, text []byte, flags int)
	TitleBlockTOML(out *bytes.Buffer, data *title)
	Aside(out *bytes.Buffer, text []byte)
	Note(out *bytes.Buffer, text []byte)
	Figure(out *bytes.Buffer, text []byte, caption []byte)

	// Span-level callbacks
	AutoLink(out *bytes.Buffer, link []byte, kind int)
	CodeSpan(out *bytes.Buffer, text []byte)
	// CalloutText is called when a callout is seen in the text. Id is the text
	// seen between < and > and ids references the callout counter(s) in the code.
	CalloutText(out *bytes.Buffer, id string, ids []string)
	// Called when a callout is seen in a code block. Index is the callout counter, id
	// is the number seen between < and >.
	CalloutCode(out *bytes.Buffer, index, id string)
	DoubleEmphasis(out *bytes.Buffer, text []byte)
	Emphasis(out *bytes.Buffer, text []byte)
	Subscript(out *bytes.Buffer, text []byte)
	Superscript(out *bytes.Buffer, text []byte)
	Image(out *bytes.Buffer, link []byte, title []byte, alt []byte, subfigure bool)
	LineBreak(out *bytes.Buffer)
	Link(out *bytes.Buffer, link []byte, title []byte, content []byte)
	RawHtmlTag(out *bytes.Buffer, tag []byte)
	TripleEmphasis(out *bytes.Buffer, text []byte)
	StrikeThrough(out *bytes.Buffer, text []byte)
	FootnoteRef(out *bytes.Buffer, ref []byte, id int)
	Index(out *bytes.Buffer, primary, secondary []byte, prim bool)
	Citation(out *bytes.Buffer, link, title []byte)
	Abbreviation(out *bytes.Buffer, abbr, title []byte)
	Example(out *bytes.Buffer, index int)
	Math(out *bytes.Buffer, text []byte, display bool)

	// Low-level callbacks
	Entity(out *bytes.Buffer, entity []byte)
	NormalText(out *bytes.Buffer, text []byte)

	// Header and footer
	DocumentHeader(out *bytes.Buffer, start bool)
	DocumentFooter(out *bytes.Buffer, start bool)

	// Frontmatter, mainmatter or backmatter
	DocumentMatter(out *bytes.Buffer, matter int)
	References(out *bytes.Buffer, citations map[string]*citation)

	// Helper functions
	Flags() int

	SetInlineAttr(*inlineAttr)
	inlineAttr() *inlineAttr
}

// Callback functions for inline parsing. One such function is defined
// for each character that triggers a response when parsing inline data.
type inlineParser func(p *parser, out *bytes.Buffer, data []byte, offset int) int

// Parser holds runtime state used by the parser.
// This is constructed by the Markdown function.
type parser struct {
	r                    Renderer
	refs                 map[string]*reference
	citations            map[string]*citation
	abbreviations        map[string]*abbreviation
	examples             map[string]int
	callouts             map[string][]string
	codeBlock            int // count codeblock for callout ID generation
	inlineCallback       [256]inlineParser
	flags                int
	nesting              int
	maxNesting           int
	insideLink           bool
	insideDefinitionList bool // when in def. list ... TODO(miek):doc
	insideList           int  // list in list counter
	insideFigure         bool // when inside a F> paragraph
	displayMath          bool

	// Footnotes need to be ordered as well as available to quickly check for
	// presence. If a ref is also a footnote, it's stored both in refs and here
	// in notes. Slice is nil if footnotes not enabled.
	notes []*reference

	appendix   bool // have we seen a {backmatter}?
	titleblock bool // have we seen a titleblock

	partCount    int // TODO, keep track of part counts (-#)
	chapterCount int // TODO, keep track of chapter count (#)

	// Placeholder IAL that can be added to blocklevel elements.
	ial *inlineAttr

	// Prevent identical header anchors by appending -<sequence_number> starting
	// with -1, this is the same thing that pandoc does.
	anchors map[string]int
}

// Markdown is an io.Writer. Writing a buffer with markdown text will be converted to
// the output format the renderer outputs. Note that the conversion only takes place
// when String() or Bytes() is called.
type Markdown struct {
	renderer   Renderer
	extensions int
	in         *bytes.Buffer
	out        *bytes.Buffer

	renderedSinceLastWrite bool
}

func NewMarkdown(renderer Renderer, extensions int) *Markdown {
	return &Markdown{renderer, extensions, &bytes.Buffer{}, &bytes.Buffer{}, false}
}

func (m *Markdown) Write(p []byte) (n int, err error) {
	m.renderedSinceLastWrite = false
	return m.in.Write(p)
}

func (m *Markdown) String() string { m.render(); return m.out.String() }
func (m *Markdown) Bytes() []byte  { m.render(); return m.out.Bytes() }

func (m *Markdown) render() {
	if m.renderer == nil {
		// default to Html renderer
	}
	if m.renderedSinceLastWrite {
		return
	}
	m.out = Parse(m.in.Bytes(), m.renderer, m.extensions)
	m.renderedSinceLastWrite = true
}

// Parse is the main rendering function.
// It parses and renders a block of markdown-encoded text.
// The supplied Renderer is used to format the output, and extensions dictates
// which non-standard extensions are enabled.
//
// To use the supplied Html or XML renderers, see HtmlRenderer, XmlRenderer and
// Xml2Renderer, respectively.
func Parse(input []byte, renderer Renderer, extensions int) *bytes.Buffer {
	// no point in parsing if we can't render
	if renderer == nil {
		return nil
	}

	// fill in the render structure
	p := new(parser)
	p.r = renderer
	p.flags = extensions
	p.refs = make(map[string]*reference)
	p.abbreviations = make(map[string]*abbreviation)
	p.anchors = make(map[string]int)
	p.examples = make(map[string]int)
	// newly created in 'callouts'
	p.maxNesting = 16
	p.insideLink = false

	// register inline parsers
	p.inlineCallback['*'] = emphasis
	p.inlineCallback['_'] = emphasis
	p.inlineCallback['~'] = emphasis
	p.inlineCallback['`'] = codeSpan
	p.inlineCallback['\n'] = lineBreak
	p.inlineCallback['['] = link
	p.inlineCallback['<'] = leftAngle
	p.inlineCallback['\\'] = escape
	p.inlineCallback['&'] = entity
	p.inlineCallback['{'] = leftBrace
	p.inlineCallback['^'] = superscript // subscript is handled in emphasis
	p.inlineCallback['('] = index       // also find example list references and cross references
	p.inlineCallback['$'] = math

	if extensions&EXTENSION_AUTOLINK != 0 {
		p.inlineCallback[':'] = autoLink
	}

	if extensions&EXTENSION_FOOTNOTES != 0 {
		p.notes = make([]*reference, 0)
	}

	if extensions&EXTENSION_CITATION != 0 {
		p.inlineCallback['@'] = citationReference // @ref, short form of citations
		p.citations = make(map[string]*citation)
	}

	first := firstPass(p, input, 0)
	second := secondPass(p, first.Bytes(), 0)
	return second
}

// first pass:
// - extract references
// - extract abbreviations
// - expand tabs
// - normalize newlines
// - copy everything else
// - add missing newlines before fenced code blocks
// - include includes
func firstPass(p *parser, input []byte, depth int) *bytes.Buffer {
	var out bytes.Buffer
	if depth > 8 {
		log.Printf("mmark: nested includes depth > 8")
		out.WriteByte('\n')
		return &out
	}

	tabSize := _TAB_SIZE_DEFAULT
	beg, end := 0, 0
	lastLineWasBlank := false
	lastFencedCodeBlockEnd := 0
	for beg < len(input) { // iterate over lines
		if beg >= lastFencedCodeBlockEnd { // don't parse inside fenced code blocks
			if end = isReference(p, input[beg:], tabSize); end > 0 {
				beg += end
				continue
			}
		}
		// skip to the next line
		end = beg
		for end < len(input) && input[end] != '\n' && input[end] != '\r' {
			end++
		}

		if p.flags&EXTENSION_FENCED_CODE != 0 {
			// when last line was none blank and a fenced code block comes after
			if beg >= lastFencedCodeBlockEnd {
				// Keep the apppend '\n', other the tests fails.
				// The original PR (149) didn't need this. I do need this
				// prolly because of the CommonMark changes I made.
				//if i := p.fencedCode(&out, input[beg:], false); i > 0 { // does not pass tests
				if i := p.fencedCode(&out, append(input[beg:], '\n'), false); i > 0 {
					if !lastLineWasBlank {
						out.WriteByte('\n') // need to inject additional linebreak
					}
					lastFencedCodeBlockEnd = beg + i
				}
			}
			lastLineWasBlank = end == beg
		}

		// add the line body if present
		if end > beg {
			if end < lastFencedCodeBlockEnd { // Do not expand tabs while inside fenced code blocks.
				out.Write(input[beg:end])
			} else {
				if p.flags&EXTENSION_INCLUDE != 0 && input[beg] == '{' {
					if beg == 0 || (beg > 0 && input[beg-1] == '\n') {
						if j := p.include(&out, input[beg:end], depth); j > 0 {
							beg += j
						}
					}
				}
				expandTabs(&out, input[beg:end], tabSize)
			}
		}
		out.WriteByte('\n')

		if end < len(input) && input[end] == '\r' {
			end++
		}
		if end < len(input) && input[end] == '\n' {
			end++
		}
		beg = end
	}

	// empty input?
	if out.Len() == 0 {
		out.WriteByte('\n')
	}

	return &out
}

// second pass: actual rendering
func secondPass(p *parser, input []byte, depth int) *bytes.Buffer {
	var output bytes.Buffer

	p.r.DocumentHeader(&output, depth == 0)
	p.block(&output, input)

	if p.flags&EXTENSION_FOOTNOTES != 0 && len(p.notes) > 0 {
		p.r.Footnotes(&output, func() bool {
			flags := _LIST_ITEM_BEGINNING_OF_LIST
			for _, ref := range p.notes {
				var buf bytes.Buffer
				if ref.hasBlock {
					flags |= _LIST_ITEM_CONTAINS_BLOCK
					p.block(&buf, ref.title)
				} else {
					p.inline(&buf, ref.title)
				}
				p.r.FootnoteItem(&output, ref.link, buf.Bytes(), flags)
				flags &^= _LIST_ITEM_BEGINNING_OF_LIST | _LIST_ITEM_CONTAINS_BLOCK
			}

			return true
		})
	}
	if !p.appendix {
		// appendix not started in doc, start it now and output references
		p.r.DocumentMatter(&output, _DOC_BACK_MATTER)
		p.r.References(&output, p.citations)
		p.appendix = true
	}
	p.r.DocumentFooter(&output, depth == 0)

	if p.nesting != 0 {
		panic("Nesting level did not end at zero")
	}

	return &output
}

//
// Link references
//
// This section implements support for references that (usually) appear
// as footnotes in a document, and can be referenced anywhere in the document.
// The basic format is:
//
//    [1]: http://www.google.com/ "Google"
//    [2]: http://www.github.com/ "Github"
//
// Anywhere in the document, the reference can be linked by referring to its
// label, i.e., 1 and 2 in this example, as in:
//
//    This library is hosted on [Github][2], a git hosting site.
//
// Actual footnotes as specified in Pandoc and supported by some other Markdown
// libraries such as php-markdown are also taken care of. They look like this:
//
//    This sentence needs a bit of further explanation.[^note]
//
//    [^note]: This is the explanation.
//
// Footnotes should be placed at the end of the document in an ordered list.
// Inline footnotes such as:
//
//    Inline footnotes^[Not supported.] also exist.
//
// are not yet supported.

// References are parsed and stored in this struct.
type reference struct {
	link     []byte
	title    []byte
	noteId   int // 0 if not a footnote ref
	hasBlock bool
}

// abbreviations are parsed and stored in this struct.
type abbreviation struct {
	title []byte
}

// citations are parsed and stored in this struct.
type citation struct {
	link  []byte
	title []byte
	xml   []byte // raw include of reference XML
	typ   byte   // 'i' for informal, 'n' normative (default = 'i')
	seq   int    // sequence number for I-Ds
}

// Check whether or not data starts with a reference link.
// If so, it is parsed and stored in the list of references
// (in the render struct).
// Returns the number of bytes to skip to move past it,
// or zero if the first line is not a reference.
func isReference(p *parser, data []byte, tabSize int) int {
	// up to 3 optional leading spaces
	if len(data) < 4 {
		return 0
	}
	i := 0
	for i < 3 && data[i] == ' ' { // break tests if this is 'iswhitespace'
		i++
	}

	noteId := 0
	abbrId := ""

	// id part: anything but a newline between brackets
	// abbreviations start with *[
	if data[i] != '[' && data[i] != '*' {
		return 0
	}
	if data[i] == '*' && (i < len(data)-1 && data[i+1] != '[') {
		return 0
	}
	if data[i] == '*' && p.flags&EXTENSION_ABBREVIATIONS != 0 {
		abbrId = "yes" // any non empty
	}
	i++
	if p.flags&EXTENSION_FOOTNOTES != 0 {
		if data[i] == '^' {
			// we can set it to anything here because the proper noteIds will
			// be assigned later during the second pass. It just has to be != 0
			noteId = 1
			i++
		}
	}
	idOffset := i
	for i < len(data) && data[i] != '\n' && data[i] != '\r' && data[i] != ']' {
		if data[i] == '\\' {
			i++
		}
		i++
	}
	if i >= len(data) || data[i] != ']' {
		return 0
	}
	idEnd := i
	if abbrId != "" {
		abbrId = string(data[idOffset+1 : idEnd])
	}

	// spacer: colon (space | tab)* newline? (space | tab)*
	i++
	if i >= len(data) || data[i] != ':' {
		return 0
	}
	i++
	for i < len(data) && (data[i] == ' ' || data[i] == '\t') {
		i++
	}
	if i < len(data) && (data[i] == '\n' || data[i] == '\r') {
		i++
		if i < len(data) && data[i] == '\n' && data[i-1] == '\r' {
			i++
		}
	}
	for i < len(data) && (data[i] == ' ' || data[i] == '\t') {
		i++
	}
	if i >= len(data) {
		return 0
	}

	var (
		linkOffset, linkEnd   int
		titleOffset, titleEnd int
		lineEnd               int
		raw                   []byte
		hasBlock              bool
	)

	if p.flags&EXTENSION_FOOTNOTES != 0 && noteId != 0 {
		linkOffset, linkEnd, raw, hasBlock = scanFootnote(p, data, i, tabSize)
		lineEnd = linkEnd
	} else if abbrId != "" {
		titleOffset, titleEnd, lineEnd = scanAbbreviation(p, data, idEnd)
		p.abbreviations[abbrId] = &abbreviation{title: data[titleOffset:titleEnd]}
		return lineEnd
	} else {
		linkOffset, linkEnd, titleOffset, titleEnd, lineEnd = scanLinkRef(p, data, i)
	}
	if lineEnd == 0 {
		return 0
	}

	// a valid ref has been found
	ref := &reference{
		noteId:   noteId,
		hasBlock: hasBlock,
	}

	if noteId > 0 {
		// reusing the link field for the id since footnotes don't have links
		ref.link = data[idOffset:idEnd]
		// if footnote, it's not really a title, it's the contained text
		ref.title = raw
	} else {
		ref.link = data[linkOffset:linkEnd]
		ref.title = data[titleOffset:titleEnd]
	}

	// id matches are case-insensitive
	id := string(bytes.ToLower(data[idOffset:idEnd]))

	// CommonMark don't overwrite newly found references
	if _, ok := p.refs[id]; !ok {
		p.refs[id] = ref
	}

	return lineEnd
}

func scanLinkRef(p *parser, data []byte, i int) (linkOffset, linkEnd, titleOffset, titleEnd, lineEnd int) {
	// link: whitespace-free sequence, optionally between angle brackets
	if data[i] == '<' {
		i++
	}
	linkOffset = i
	for i < len(data) && data[i] != ' ' && data[i] != '\t' && data[i] != '\n' && data[i] != '\r' {
		i++
	}
	linkEnd = i
	if data[linkOffset] == '<' && data[linkEnd-1] == '>' {
		linkOffset++
		linkEnd--
	}

	// optional spacer: (space | tab)* (newline | '\'' | '"' | '(' )
	for i < len(data) && (data[i] == ' ' || data[i] == '\t') {
		i++
	}
	if i < len(data) && data[i] != '\n' && data[i] != '\r' && data[i] != '\'' && data[i] != '"' && data[i] != '(' {
		return
	}

	// compute end-of-line
	if i >= len(data) || data[i] == '\r' || data[i] == '\n' {
		lineEnd = i
	}
	if i+1 < len(data) && data[i] == '\r' && data[i+1] == '\n' {
		lineEnd++
	}

	// optional (space|tab)* spacer after a newline
	if lineEnd > 0 {
		i = lineEnd + 1
		for i < len(data) && (data[i] == ' ' || data[i] == '\t') {
			i++
		}
	}

	// optional title: any non-newline sequence enclosed in '"() alone on its line
	if i+1 < len(data) && (data[i] == '\'' || data[i] == '"' || data[i] == '(') {
		i++
		titleOffset = i

		// look for EOL
		for i < len(data) && data[i] != '\n' && data[i] != '\r' {
			i++
		}
		if i+1 < len(data) && data[i] == '\n' && data[i+1] == '\r' {
			titleEnd = i + 1
		} else {
			titleEnd = i
		}

		// step back
		i--
		for i > titleOffset && (data[i] == ' ' || data[i] == '\t') {
			i--
		}
		if i > titleOffset && (data[i] == '\'' || data[i] == '"' || data[i] == ')') {
			lineEnd = titleEnd
			titleEnd = i
		}
	}

	return
}

// The first bit of this logic is the same as (*parser).listItem, but the rest
// is much simpler. This function simply finds the entire block and shifts it
// over by one tab if it is indeed a block (just returns the line if it's not).
// blockEnd is the end of the section in the input buffer, and contents is the
// extracted text that was shifted over one tab. It will need to be rendered at
// the end of the document.
func scanFootnote(p *parser, data []byte, i, indentSize int) (blockStart, blockEnd int, contents []byte, hasBlock bool) {
	if i == 0 || len(data) == 0 {
		return
	}

	// skip leading whitespace on first line
	for i < len(data) && data[i] == ' ' {
		i++
	}

	blockStart = i

	// find the end of the line
	blockEnd = i
	for i < len(data) && data[i-1] != '\n' {
		i++
	}

	// get working buffer
	var raw bytes.Buffer

	// put the first line into the working buffer
	raw.Write(data[blockEnd:i])
	blockEnd = i

	// process the following lines
	containsBlankLine := false

gatherLines:
	for blockEnd < len(data) {
		i++

		// find the end of this line
		for i < len(data) && data[i-1] != '\n' {
			i++
		}

		// if it is an empty line, guess that it is part of this item
		// and move on to the next line
		if p.isEmpty(data[blockEnd:i]) > 0 {
			containsBlankLine = true
			blockEnd = i
			continue
		}

		n := 0
		if n = isIndented(data[blockEnd:i], indentSize); n == 0 {
			// this is the end of the block.
			// we don't want to include this last line in the index.
			break gatherLines
		}

		// if there were blank lines before this one, insert a new one now
		if containsBlankLine {
			raw.WriteByte('\n')
			containsBlankLine = false
		}

		// get rid of that first tab, write to buffer
		raw.Write(data[blockEnd+n : i])
		hasBlock = true

		blockEnd = i
	}

	if data[blockEnd-1] != '\n' {
		raw.WriteByte('\n')
	}

	contents = raw.Bytes()

	return
}

func scanAbbreviation(p *parser, data []byte, i int) (titleOffset, titleEnd, lineEnd int) {
	lineEnd = i
	for lineEnd < len(data) && data[lineEnd] != '\n' {
		lineEnd++
	}

	if len(data[i+2:lineEnd]) == 0 || p.isEmpty(data[i+2:lineEnd]) > 0 {
		return i + 2, i + 2, lineEnd
	}

	titleOffset = i + 2
	for data[titleOffset] == ' ' {
		titleOffset++
	}
	titleEnd = lineEnd
	for data[titleEnd-1] == ' ' {
		titleEnd--
	}

	return
}

// Miscellaneous helper functions

func ispunct(c byte) bool  { return unicode.IsPunct(rune(c)) }
func isletter(c byte) bool { return unicode.IsLetter(rune(c)) }
func isalnum(c byte) bool  { return (unicode.IsNumber(rune(c)) || unicode.IsLetter(rune(c))) }
func isnum(c byte) bool    { return unicode.IsNumber(rune(c)) }
func isspace(c byte) bool  { return unicode.IsSpace(rune(c)) }
func isupper(c byte) bool  { return unicode.IsUpper(rune(c)) }
func islower(c byte) bool  { return !unicode.IsUpper(rune(c)) }

func iswhitespace(c byte) bool { // better name?
	if c == '\n' || c == '\r' {
		return false
	}
	return unicode.IsSpace(rune(c))
}

// check if the string only contains, i, v, x, c and l. If uppercase is true, check uppercase version.
func isroman(digit byte, uppercase bool) bool {
	if !uppercase {
		if digit == 'i' || digit == 'v' || digit == 'x' || digit == 'c' || digit == 'l' {
			return true
		}
		return false
	}
	if digit == 'I' || digit == 'V' || digit == 'X' || digit == 'C' || digit == 'L' {
		return true
	}
	return false
}

// replace {{file.md}} with the contents of the file.
func (p *parser) include(out *bytes.Buffer, data []byte, depth int) int {
	i := 0
	if len(data) < 3 {
		return 0
	}
	if data[i] != '{' && data[i+1] != '{' {
		return 0
	}

	// find the end delimiter
	end, j := 0, 0
	for end = i; end < len(data) && j < 2; end++ {
		if data[end] == '}' {
			j++
		} else {
			j = 0
		}
	}
	if j < 2 && end >= len(data) {
		return 0
	}

	name := string(data[i+2 : end-2])
	input, err := ioutil.ReadFile(name)
	if err != nil {
		log.Printf("mmark: failed: `%s': %s", name, err)
		return end
	}

	if len(input) == 0 {
		input = []byte{'\n'}
	}
	if input[len(input)-1] != '\n' {
		input = append(input, '\n')
	}
	first := firstPass(p, input, depth+1)
	out.Write(first.Bytes())
	return end
}

// replace <{{file.go}}[address] with the contents of the file. Pay attention to the indentation of the
// include and prefix the code with that number of spaces + 4, it returns the new bytes and a boolean
// indicating we've detected a code include.
func (p *parser) codeInclude(out *bytes.Buffer, data []byte) int {
	// TODO: this is not an inline element
	i := 0
	l := len(data)
	if l < 3 {
		return 0
	}
	if data[i] != '<' && data[i+1] != '{' && data[i+2] != '{' {
		return 0
	}

	// find the end delimiter
	end, j := 0, 0
	for end = i; end < l && j < 2; end++ {
		if data[end] == '}' {
			j++
		} else {
			j = 0
		}
	}
	if j < 2 && end >= l {
		return 0
	}

	lang := ""
	// found <{{filename}}
	// this could be the end, or we could have an option [address] -block attached
	filename := data[i+3 : end-2]
	// get the extension of the filename, if it is a member of a predefined set a
	// language we use it as the lang (and we will emit <sourcecode>)
	if x := path.Ext(string(filename)); x != "" {
		// x includes the dot
		if _, ok := codes[x[1:]]; ok {
			lang = x[1:]
		}
	}

	// Now a possible address in blockquotes
	var address []byte
	if end < l && data[end] == '[' {
		j = end
		for j < l && data[j] != ']' {
			j++
		}
		if j == l {
			// assuming no address
			address = nil
			end = l
		} else {
			address = data[end+1 : j]
			end = j + 1
		}
	}

	code := parseCode(address, filename)

	if len(code) == 0 {
		code = []byte{'\n'}
	}
	if code[len(code)-1] != '\n' {
		code = append(code, '\n')
	}

	// if the next line starts with Figure: we consider that a caption
	var caption bytes.Buffer
	if end < l-1 && bytes.HasPrefix(data[end+1:], []byte("Figure: ")) {
		line := end + 1
		j := end + 1
		for line < l {
			j++
			// find the end of this line
			for j <= l && data[j-1] != '\n' {
				j++
			}
			if p.isEmpty(data[line:j]) > 0 {
				break
			}
			line = j
		}
		p.inline(&caption, data[end+1+8:j-1]) // +8 for 'Figure: '
		end = j - 1
	}

	co := ""
	if p.ial != nil {
		co = p.ial.Value("callout")
	}

	p.r.SetInlineAttr(p.ial)
	p.ial = nil

	if co != "" {
		var callout bytes.Buffer
		callouts(p, &callout, code, 0, co)
		p.r.BlockCode(out, callout.Bytes(), lang, caption.Bytes(), p.insideFigure, true)
	} else {
		p.callouts = nil
		p.r.BlockCode(out, code, lang, caption.Bytes(), p.insideFigure, false)
	}
	p.r.SetInlineAttr(nil) // reset it again. TODO(miek): double check

	return end
}

// replace tab characters with spaces, aligning to the next tab_size column.
// always ends output with a newline
func expandTabs(out *bytes.Buffer, line []byte, tabSize int) {
	// first, check for common cases: no tabs, or only tabs at beginning of line
	i, prefix := 0, 0
	slowcase := false
	for i = 0; i < len(line); i++ {
		if line[i] == '\t' {
			if prefix == i {
				prefix++
			} else {
				slowcase = true
				break
			}
		}
	}

	// no need to decode runes if all tabs are at the beginning of the line
	if !slowcase {
		for i = 0; i < prefix*tabSize; i++ {
			out.WriteByte(' ')
		}
		out.Write(line[prefix:])
		return
	}

	// the slow case: we need to count runes to figure out how
	// many spaces to insert for each tab
	column := 0
	i = 0
	for i < len(line) {
		start := i
		for i < len(line) && line[i] != '\t' {
			_, size := utf8.DecodeRune(line[i:])
			i += size
			column++
		}

		if i > start {
			out.Write(line[start:i])
		}

		if i >= len(line) {
			break
		}

		for {
			out.WriteByte(' ')
			column++
			if column%tabSize == 0 {
				break
			}
		}

		i++
	}
}

// Find if a line counts as indented or not.
// Returns number of characters the indent is (0 = not indented).
func isIndented(data []byte, indentSize int) int {
	if len(data) == 0 {
		return 0
	}
	if data[0] == '\t' {
		return 1
	}
	if len(data) < indentSize {
		return 0
	}
	for i := 0; i < indentSize; i++ {
		if data[i] != ' ' {
			return 0
		}
	}
	return indentSize
}

// Create a url-safe slug for fragments
func slugify(in []byte) []byte {
	if len(in) == 0 {
		return in
	}
	out := make([]byte, 0, len(in))
	sym := false

	for _, ch := range in {
		if isalnum(ch) {
			sym = false
			out = append(out, ch)
		} else if sym {
			continue
		} else {
			out = append(out, '-')
			sym = true
		}
	}
	var a, b int
	var ch byte
	for a, ch = range out {
		if ch != '-' {
			break
		}
	}
	for b = len(out) - 1; b > 0; b-- {
		if out[b] != '-' {
			break
		}
	}
	return out[a : b+1]
}
