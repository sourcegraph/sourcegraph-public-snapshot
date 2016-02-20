// HTML rendering backend

package mmark

import (
	"bytes"
	xmllib "encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"
)

// Html renderer configuration options.
const (
	HTML_SKIP_HTML             = 1 << iota // skip preformatted HTML blocks
	HTML_SKIP_STYLE                        // skip embedded <style> elements
	HTML_SKIP_IMAGES                       // skip embedded images
	HTML_SKIP_LINKS                        // skip all links
	HTML_SAFELINK                          // only link to trusted protocols
	HTML_NOFOLLOW_LINKS                    // only link with rel="nofollow"
	HTML_HREF_TARGET_BLANK                 // add a blank target
	HTML_OMIT_CONTENTS                     // skip the main contents (for a standalone table of contents)
	HTML_COMPLETE_PAGE                     // generate a complete HTML page
	HTML_FOOTNOTE_RETURN_LINKS             // generate a link at the end of a footnote to return to the source
)

var (
	alignments = []string{
		"left",
		"right",
		"center",
	}
)

type HtmlRendererParameters struct {
	// Prepend this text to each relative URL.
	AbsolutePrefix string
	// Add this text to each footnote anchor, to ensure uniqueness.
	FootnoteAnchorPrefix string
	// Show this text inside the <a> tag for a footnote return link, if the
	// HTML_FOOTNOTE_RETURN_LINKS flag is enabled. If blank, the string
	// <sup>[return]</sup> is used.
	FootnoteReturnLinkContents string
}

// Html is a type that implements the Renderer interface for HTML output.
//
// Do not create this directly, instead use the HtmlRenderer function.
type html struct {
	flags    int    // HTML_* options
	closeTag string // how to end singleton tags: either " />" or ">"
	css      string // optional css file url (used with HTML_COMPLETE_PAGE)
	head     string // option html file to be included

	// store the IAL we see for this block element
	ial *inlineAttr

	// titleBlock in TOML
	titleBlock *title

	parameters HtmlRendererParameters

	// table of contents data
	headerCount  int
	currentLevel int
	toc          *bytes.Buffer

	appendix bool

	// index, map idx to id
	index      map[idx][]string
	indexCount int

	// (@good) example list group counter
	group map[string]int
}

type idx struct {
	primary, secondary string
}

const htmlClose = ">"

// HtmlRenderer creates and configures an Html object, which
// satisfies the Renderer interface.
//
// flags is a set of HTML_* options ORed together.
// css is a URL for the document's stylesheet.
func HtmlRenderer(flags int, css, head string) Renderer {
	return HtmlRendererWithParameters(flags, css, head, HtmlRendererParameters{})
}

func HtmlRendererWithParameters(flags int, css, head string, renderParameters HtmlRendererParameters) Renderer {
	// configure the rendering engine
	closeTag := htmlClose

	if renderParameters.FootnoteReturnLinkContents == "" {
		renderParameters.FootnoteReturnLinkContents = `<sup>[return]</sup>`
	}

	anchorOrID = "id" // use id= when seeing #id. Also see ial.go

	return &html{
		flags:      flags,
		closeTag:   closeTag,
		css:        css,
		head:       head,
		parameters: renderParameters,

		headerCount:  0,
		currentLevel: 0,
		toc:          new(bytes.Buffer),

		index: make(map[idx][]string),
	}
}

// Using if statements is a bit faster than a switch statement. As the compiler
// improves, this should be unnecessary this is only worthwhile because
// attrEscape is the single largest CPU user in normal use.
// Also tried using map, but that gave a ~3x slowdown.
func escapeSingleChar(char byte) (string, bool) {
	if char == '"' {
		return "&quot;", true
	}
	if char == '&' {
		return "&amp;", true
	}
	if char == '<' {
		return "&lt;", true
	}
	if char == '>' {
		return "&gt;", true
	}
	return "", false
}

func attrEscape(out *bytes.Buffer, src []byte) {
	org := 0
	for i, ch := range src {
		if entity, ok := escapeSingleChar(ch); ok {
			if i > org {
				// copy all the normal characters since the last escape
				out.Write(src[org:i])
			}
			org = i + 1
			out.WriteString(entity)
		}
	}
	if org < len(src) {
		out.Write(src[org:])
	}
}

func attrEscapeInCode(r Renderer, out *bytes.Buffer, src []byte) {
	var prev byte
	j := 0
	for i := 0; i < len(src); i++ {
		ch := src[i]
		if ch == '<' && prev != '\\' {
			if x := leftAngleCode(src[i:]); x > 0 {
				j++
				// Call the renderer's CalloutCode
				r.CalloutCode(out, strconv.Itoa(j), string(src[i:i+x+1]))
				i += x
				prev = ch
				continue
			}
		}
		if ch == '\\' && i < len(src)-1 && src[i+1] == '<' {
			// skip \\ here
			prev = ch
			continue
		}
		if entity, ok := escapeSingleChar(ch); ok {
			out.WriteString(entity)
			prev = ch
			continue
		}
		out.WriteByte(ch)
		prev = ch
	}
}

func entityEscapeWithSkip(out *bytes.Buffer, src []byte, skipRanges [][]int) {
	end := 0
	for _, rang := range skipRanges {
		attrEscape(out, src[end:rang[0]])
		out.Write(src[rang[0]:rang[1]])
		end = rang[1]
	}
	attrEscape(out, src[end:])
}

func (options *html) Flags() int {
	return options.flags
}

func (options *html) TitleBlockTOML(out *bytes.Buffer, block *title) {
	if options.flags&HTML_COMPLETE_PAGE == 0 { // use STANDALONE
		return
	}
	options.titleBlock = block
	ending := ""
	out.WriteString("<head>\n")
	out.WriteString("  <title>")
	options.NormalText(out, []byte(options.titleBlock.Title))
	out.WriteString("</title>\n")
	out.WriteString("  <meta name=\"GENERATOR\" content=\"Mmark Markdown Processor v")
	out.WriteString(VERSION)
	out.WriteString("\"")
	out.WriteString(ending)
	out.WriteString(">\n")
	out.WriteString("  <meta charset=\"utf-8\"")
	out.WriteString(ending)
	out.WriteString(">\n")
	if options.css != "" {
		out.WriteString("  <link rel=\"stylesheet\" type=\"text/css\" href=\"")
		attrEscape(out, []byte(options.css))
		out.WriteString("\"")
		out.WriteString(ending)
		out.WriteString(">\n")
	}
	if options.head != "" {
		headBytes, err := ioutil.ReadFile(options.head)
		if err != nil {
			log.Printf("mmark: failed: `%s': %s", options.head, err)
		} else {
			out.Write(headBytes)
		}

	}
	out.WriteString("</head>\n")
	out.WriteString("<body>\n")

	// Write some elements of the TOML block in the doc as well.
}

func (options *html) Part(out *bytes.Buffer, text func() bool, id string) {
	if id != "" {
		out.WriteString(fmt.Sprintf("<h1 class=\"part\" id=\"%s\">", id))
	} else {
		out.WriteString(fmt.Sprintf("<h1 class=\"part\""))
	}
	text()
	out.WriteString(fmt.Sprintf("</h1>\n"))
}

func (options *html) SpecialHeader(out *bytes.Buffer, what []byte, text func() bool, id string) {
	options.inlineAttr() //reset the IAL
	if id != "" {
		out.WriteString(fmt.Sprintf("<h1 class=\""+string(what)+"\" id=\"%s\">", id))
	} else {
		out.WriteString(fmt.Sprintf("<h1 class=\"" + string(what) + "\""))
	}
	text()
	out.WriteString(fmt.Sprintf("</h1>\n"))
}

func (options *html) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	marker := out.Len()
	doubleSpace(out)

	ial := options.inlineAttr()
	ial.GetOrDefaultId(id)
	if options.appendix {
		ial.GetOrDefaultClass("appendix")
	}

	out.WriteString(fmt.Sprintf("<h%d%s>", level, ial.String()))

	if !text() {
		out.Truncate(marker)
		return
	}
	// special section closing etc. etc. TODO(miek)
	out.WriteString(fmt.Sprintf("</h%d>\n", level))
}

func (options *html) CommentHtml(out *bytes.Buffer, text []byte) {
	if options.flags&HTML_SKIP_HTML != 0 {
		return
	}

	doubleSpace(out)
	out.Write(text)
	out.WriteByte('\n')
}

func (options *html) BlockHtml(out *bytes.Buffer, text []byte) {
	if options.flags&HTML_SKIP_HTML != 0 {
		return
	}

	doubleSpace(out)
	out.Write(text)
	out.WriteByte('\n')
}

func (options *html) HRule(out *bytes.Buffer) {
	doubleSpace(out)
	out.WriteString("<hr")
	out.WriteString(options.closeTag)
	out.WriteByte('\n')
}

func (options *html) CalloutCode(out *bytes.Buffer, index, id string) {
	out.WriteString("<span class=\"callout\">")
	out.WriteString(index)
	out.WriteString("</span>")
	return
}

func (options *html) CalloutText(out *bytes.Buffer, id string, ids []string) {
	for i, k := range ids {
		out.WriteString("<span class=\"callout\">")
		out.WriteString(k)
		out.WriteString("</span>")
		if i < len(ids)-1 {
			out.WriteString(" ")
		}
	}
}

func (options *html) BlockCode(out *bytes.Buffer, text []byte, lang string, caption []byte, subfigure bool, callout bool) {
	doubleSpace(out)

	ial := options.inlineAttr()
	// TODO(miek): subfigure

	// if theere is a caption we wrap the thing in the figure
	if len(caption) > 0 {
		out.WriteString("<figure" + ial.String() + ">\n")
	}

	// parse out the language names/classes
	count := 0
	for _, elt := range strings.Fields(lang) {
		if elt[0] == '.' {
			elt = elt[1:]
		}
		if len(elt) == 0 {
			continue
		}
		if count == 0 {
			out.WriteString("<pre><code class=\"language-")
		} else {
			out.WriteByte(' ')
		}
		attrEscape(out, []byte(elt))
		count++
	}

	if count == 0 {
		out.WriteString("<pre><code>")
	} else {
		out.WriteString("\">")
	}
	if callout {
		attrEscapeInCode(options, out, text)
	} else {
		attrEscape(out, text)
	}
	out.WriteString("</code></pre>\n")
	if len(caption) > 0 {
		out.WriteString("<figcaption>\n")
		out.Write(caption)
		out.WriteString("</figcaption>\n")
		out.WriteString("</figure>\n")
	}
}

func (options *html) BlockQuote(out *bytes.Buffer, text []byte, attribution []byte) {
	// attribution can potentially be split on --: meta -- who
	ial := options.inlineAttr()
	parts := bytes.Split(attribution, []byte("--"))
	for _, p := range parts {
		bytes.TrimSpace(p)
	}
	doubleSpace(out)
	out.WriteString("<blockquote" + ial.String() + ">\n")
	out.Write(text)
	if len(parts) == 2 {
		out.WriteString("<footer>")
		if len(parts[0]) > 0 {
			// could be left empty
			out.WriteString("&mdash; ")
		}
		out.Write(parts[0])
		out.WriteString("<span class=\"quote-who\">")
		out.Write(parts[1])
		out.WriteString("</span>")
		out.WriteString("</footer>")
	}
	out.WriteString("</blockquote>\n")
}

func (options *html) Aside(out *bytes.Buffer, text []byte) {
	doubleSpace(out)
	out.WriteString("<aside>\n")
	out.Write(text)
	out.WriteString("</aside>\n")
}

func (options *html) Note(out *bytes.Buffer, text []byte) {
	doubleSpace(out)
	out.WriteString("<aside class=\"note\">\n")
	out.Write(text)
	out.WriteString("</aside>\n")
}

func (options *html) Table(out *bytes.Buffer, header []byte, body []byte, footer []byte, columnData []int, caption []byte) {
	ial := options.inlineAttr()

	doubleSpace(out)
	out.WriteString("<table" + ial.String() + ">\n")
	if len(caption) > 0 {
		out.WriteString("<caption>\n")
		out.Write(caption)
		out.WriteString("\n</caption>\n")
	}
	out.WriteString("<thead>\n")
	out.Write(header)
	out.WriteString("</thead>\n\n<tbody>\n")
	out.Write(body)
	out.WriteString("</tbody>\n")
	if len(footer) > 0 {
		out.WriteString("<tfoot>\n")
		out.Write(footer)
		out.WriteString("</tfoot>\n")
	}
	out.WriteString("</table>\n")
}

func (options *html) TableRow(out *bytes.Buffer, text []byte) {
	doubleSpace(out)
	out.WriteString("<tr>\n")
	out.Write(text)
	out.WriteString("\n</tr>\n")
}

func (options *html) TableHeaderCell(out *bytes.Buffer, text []byte, align int) {
	doubleSpace(out)
	switch align {
	case _TABLE_ALIGNMENT_LEFT:
		out.WriteString("<th align=\"left\">")
	case _TABLE_ALIGNMENT_RIGHT:
		out.WriteString("<th align=\"right\">")
	case _TABLE_ALIGNMENT_CENTER:
		out.WriteString("<th align=\"center\">")
	default:
		out.WriteString("<th>")
	}

	out.Write(text)
	out.WriteString("</th>")
}

func (options *html) TableCell(out *bytes.Buffer, text []byte, align int) {
	doubleSpace(out)
	switch align {
	case _TABLE_ALIGNMENT_LEFT:
		out.WriteString("<td align=\"left\">")
	case _TABLE_ALIGNMENT_RIGHT:
		out.WriteString("<td align=\"right\">")
	case _TABLE_ALIGNMENT_CENTER:
		out.WriteString("<td align=\"center\">")
	default:
		out.WriteString("<td>")
	}

	out.Write(text)
	out.WriteString("</td>")
}

func (options *html) Footnotes(out *bytes.Buffer, text func() bool) {
	if options.flags&HTML_COMPLETE_PAGE != 0 {
		options.ial = &inlineAttr{class: map[string]bool{"footnotes": true}}
		options.Header(out, func() bool { out.WriteString("Footnotes"); return true }, 1, "footnotes")
	}
	// reset now that the header is out
	options.ial = nil
	out.WriteString("<div class=\"footnotes\">\n")
	if options.flags&HTML_COMPLETE_PAGE == 0 {
		options.HRule(out)
	}
	options.List(out, text, _LIST_TYPE_ORDERED, 0, nil)
	out.WriteString("</div>\n")
}

func (options *html) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	if flags&_LIST_ITEM_CONTAINS_BLOCK != 0 || flags&_LIST_ITEM_BEGINNING_OF_LIST != 0 {
		doubleSpace(out)
	}
	slug := slugify(name)
	out.WriteString(`<li id="`)
	out.WriteString(`fn:`)
	out.WriteString(options.parameters.FootnoteAnchorPrefix)
	out.Write(slug)
	out.WriteString(`">`)
	out.Write(text)
	if options.flags&HTML_FOOTNOTE_RETURN_LINKS != 0 {
		out.WriteString(` <a class="footnote-return" href="#`)
		out.WriteString(`fnref:`)
		out.WriteString(options.parameters.FootnoteAnchorPrefix)
		out.Write(slug)
		out.WriteString(`">`)
		out.WriteString(options.parameters.FootnoteReturnLinkContents)
		out.WriteString(`</a>`)
	}
	out.WriteString("</li>\n")
}

func (options *html) List(out *bytes.Buffer, text func() bool, flags, start int, group []byte) {
	marker := out.Len()
	doubleSpace(out)

	ial := options.inlineAttr()
	if start > 1 {
		ial.GetOrDefaultAttr("start", strconv.Itoa(start))
	}

	switch {
	case flags&_LIST_TYPE_ORDERED != 0:
		switch {
		case flags&_LIST_TYPE_ORDERED_ALPHA_LOWER != 0:
			ial.GetOrDefaultAttr("type", "a")
		case flags&_LIST_TYPE_ORDERED_ALPHA_UPPER != 0:
			ial.GetOrDefaultAttr("type", "A")
		case flags&_LIST_TYPE_ORDERED_ROMAN_LOWER != 0:
			ial.GetOrDefaultAttr("type", "i")
		case flags&_LIST_TYPE_ORDERED_ROMAN_UPPER != 0:
			ial.GetOrDefaultAttr("type", "I")
		case flags&_LIST_TYPE_ORDERED_GROUP != 0:
			// check start as well
			if group != nil {
				options.group[string(group)]++
				start := options.group[string(group)]
				ial.GetOrDefaultAttr("start", strconv.Itoa(start))
				ial.GetOrDefaultAttr("type", "I")
			}
		}
		out.WriteString("<ol" + ial.String() + ">")
	case flags&_LIST_TYPE_DEFINITION != 0:
		out.WriteString("<dl" + ial.String() + ">")
	default:
		out.WriteString("<ul" + ial.String() + ">")
	}
	if !text() {
		out.Truncate(marker)
		return
	}
	switch {
	case flags&_LIST_TYPE_ORDERED != 0:
		out.WriteString("</ol>\n")
	case flags&_LIST_TYPE_DEFINITION != 0:
		out.WriteString("</dl>\n")
	default:
		out.WriteString("</ul>\n")
	}
}

func (options *html) ListItem(out *bytes.Buffer, text []byte, flags int) {
	if flags&_LIST_ITEM_CONTAINS_BLOCK != 0 || flags&_LIST_ITEM_BEGINNING_OF_LIST != 0 {
		doubleSpace(out)
	}
	if flags&_LIST_TYPE_DEFINITION != 0 && flags&_LIST_TYPE_TERM == 0 {
		out.WriteString("<dd>")
		out.Write(text)
		out.WriteString("</dd>\n")
		return
	}
	if flags&_LIST_TYPE_TERM != 0 {
		out.WriteString("<dt>")
		out.Write(text)
		out.WriteString("</dt>")
		return
	}
	out.WriteString("<li>")
	out.Write(text)
	out.WriteString("</li>\n")
}

func (options *html) Example(out *bytes.Buffer, index int) {
	out.WriteByte('(')
	out.WriteString(strconv.Itoa(index))
	out.WriteByte(')')
}

func (options *html) Paragraph(out *bytes.Buffer, text func() bool, flags int) {
	marker := out.Len()
	doubleSpace(out)

	out.WriteString("<p>")
	if !text() {
		out.Truncate(marker)
		return
	}
	out.WriteString("</p>\n")
}

func (options *html) Math(out *bytes.Buffer, text []byte, display bool) {
	ial := options.inlineAttr()
	s := ial.String()
	oTag := "\\("
	cTag := "\\)"
	if display {
		oTag = "\\["
		cTag = "\\]"
	}
	out.WriteString("<span " + s + " class=\"math\">")
	out.WriteString(oTag)
	out.Write(text)
	out.WriteString(cTag)
	out.WriteString("</span>")
}

func (options *html) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	skipRanges := htmlEntity.FindAllIndex(link, -1)
	if options.flags&HTML_SAFELINK != 0 && !isSafeLink(link) && kind != _LINK_TYPE_EMAIL {
		// mark it but don't link it if it is not a safe link
		out.WriteString("<tt>")
		entityEscapeWithSkip(out, link, skipRanges)
		out.WriteString("</tt>")
		return
	}

	out.WriteString("<a href=\"")
	if kind == _LINK_TYPE_EMAIL {
		out.WriteString("mailto:")
	} else {
		options.maybeWriteAbsolutePrefix(out, link)
	}

	entityEscapeWithSkip(out, link, skipRanges)

	if options.flags&HTML_NOFOLLOW_LINKS != 0 && !isRelativeLink(link) {
		out.WriteString("\" rel=\"nofollow")
	}
	// blank target only add to external link
	if options.flags&HTML_HREF_TARGET_BLANK != 0 && !isRelativeLink(link) {
		out.WriteString("\" target=\"_blank")
	}

	out.WriteString("\">")

	// Pretty print: if we get an email address as
	// an actual URI, e.g. `mailto:foo@bar.com`, we don't
	// want to print the `mailto:` prefix
	switch {
	case bytes.HasPrefix(link, []byte("mailto://")):
		attrEscape(out, link[len("mailto://"):])
	case bytes.HasPrefix(link, []byte("mailto:")):
		attrEscape(out, link[len("mailto:"):])
	default:
		entityEscapeWithSkip(out, link, skipRanges)
	}

	out.WriteString("</a>")
}

func (options *html) CodeSpan(out *bytes.Buffer, text []byte) {
	out.WriteString("<code>")
	attrEscape(out, text)
	out.WriteString("</code>")
}

func (options *html) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<strong>")
	out.Write(text)
	out.WriteString("</strong>")
}

func (options *html) Emphasis(out *bytes.Buffer, text []byte) {
	// TODO(miek): why is this check here?
	if len(text) == 0 {
		return
	}
	out.WriteString("<em>")
	out.Write(text)
	out.WriteString("</em>")
}

func (options *html) Subscript(out *bytes.Buffer, text []byte) {
	out.WriteString("<sub>")
	out.Write(text)
	out.WriteString("</sub>")
}

func (options *html) Superscript(out *bytes.Buffer, text []byte) {
	out.WriteString("<sup>")
	out.Write(text)
	out.WriteString("</sup>")
}

func (options *html) maybeWriteAbsolutePrefix(out *bytes.Buffer, link []byte) {
	if options.parameters.AbsolutePrefix != "" && isRelativeLink(link) {
		out.WriteString(options.parameters.AbsolutePrefix)
		if link[0] != '/' {
			out.WriteByte('/')
		}
	}
}

func (options *html) Figure(out *bytes.Buffer, text []byte, caption []byte) {

}

func (options *html) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte, subfigure bool) {
	if options.flags&HTML_SKIP_IMAGES != 0 {
		return
	}
	ial := options.inlineAttr()
	out.WriteString("<figure" + ial.String() + ">")
	out.WriteString("<img src=\"")
	options.maybeWriteAbsolutePrefix(out, link)
	attrEscape(out, link)
	out.WriteString("\" alt=\"")
	if len(alt) > 0 {
		attrEscape(out, alt)
	}
	if len(title) > 0 {
		out.WriteString("\" title=\"")
		attrEscape(out, title)
	}
	out.WriteByte('"')
	out.WriteString(options.closeTag)
	if len(title) > 0 {
		out.WriteString("<figcaption>")
		out.Write(title)
		out.WriteString("</figcaption>")
	}
	out.WriteString("</figure>")
	return
}

func (options *html) LineBreak(out *bytes.Buffer) {
	out.WriteString("<br")
	out.WriteString(options.closeTag)
	out.WriteByte('\n')
}

func (options *html) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	if options.flags&HTML_SKIP_LINKS != 0 {
		// write the link text out but don't link it, just mark it with typewriter font
		out.WriteString("<tt>")
		attrEscape(out, content)
		out.WriteString("</tt>")
		return
	}

	if options.flags&HTML_SAFELINK != 0 && !isSafeLink(link) {
		// write the link text out but don't link it, just mark it with typewriter font
		out.WriteString("<tt>")
		attrEscape(out, content)
		out.WriteString("</tt>")
		return
	}

	out.WriteString("<a href=\"")
	options.maybeWriteAbsolutePrefix(out, link)
	attrEscape(out, link)
	if len(title) > 0 {
		out.WriteString("\" title=\"")
		attrEscape(out, title)
	}
	if options.flags&HTML_NOFOLLOW_LINKS != 0 && !isRelativeLink(link) {
		out.WriteString("\" rel=\"nofollow")
	}
	// blank target only add to external link
	if options.flags&HTML_HREF_TARGET_BLANK != 0 && !isRelativeLink(link) {
		out.WriteString("\" target=\"_blank")
	}

	out.WriteString("\">")
	out.Write(content)
	out.WriteString("</a>")
	return
}

func (options *html) Abbreviation(out *bytes.Buffer, abbr, title []byte) {
	if len(title) == 0 {
		out.WriteString("<abbr>")
	} else {
		out.WriteString("<abbr title=\"")
		out.Write(title)
		out.WriteString("\">")
	}
	out.Write(abbr)
	out.WriteString("</abbr>")
}

func (options *html) RawHtmlTag(out *bytes.Buffer, text []byte) {
	if options.flags&HTML_SKIP_HTML != 0 {
		return
	}
	if options.flags&HTML_SKIP_STYLE != 0 && isHtmlTag(text, "style") {
		return
	}
	if options.flags&HTML_SKIP_LINKS != 0 && isHtmlTag(text, "a") {
		return
	}
	if options.flags&HTML_SKIP_IMAGES != 0 && isHtmlTag(text, "img") {
		return
	}
	out.Write(text)
}

func (options *html) TripleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<strong><em>")
	out.Write(text)
	out.WriteString("</em></strong>")
}

func (options *html) StrikeThrough(out *bytes.Buffer, text []byte) {
	out.WriteString("<del>")
	out.Write(text)
	out.WriteString("</del>")
}

func (options *html) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	slug := slugify(ref)
	out.WriteString(`<sup class="footnote-ref" id="`)
	out.WriteString(`fnref:`)
	out.WriteString(options.parameters.FootnoteAnchorPrefix)
	out.Write(slug)
	out.WriteString(`"><a class="footnote" href="#`)
	out.WriteString(`fn:`)
	out.WriteString(options.parameters.FootnoteAnchorPrefix)
	out.Write(slug)
	out.WriteString(`">`)
	out.WriteString(strconv.Itoa(id))
	out.WriteString(`</a></sup>`)
}

func (options *html) Index(out *bytes.Buffer, primary, secondary []byte, prim bool) {
	idx := idx{string(primary), string(secondary)}
	id := ""
	if ids, ok := options.index[idx]; ok {
		// write id out and add it to the list
		id = fmt.Sprintf("#idxref:%d-%d", options.indexCount, len(ids))
		options.index[idx] = append(options.index[idx], id)
	} else {
		id = fmt.Sprintf("#idxref:%d-0", options.indexCount)
		options.index[idx] = []string{id}
	}
	out.WriteString("<span class=\"index-ref\" id=\"" + id[1:] + "\"></span>")

	options.indexCount++
}

func (options *html) Entity(out *bytes.Buffer, entity []byte) { out.Write(entity) }

func (options *html) Citation(out *bytes.Buffer, link, title []byte) {
	out.WriteString("<a class=\"cite\" href=\"#")
	out.Write(bytes.ToLower(link))
	out.WriteString("\">")
	out.Write(title)
	out.WriteString("</a>")
}

// RefAuthor is the reference author, exported because we need to be able to parse
// raw XML references when included in the document.
type RefAuthor struct {
	Fullname string `xml:"fullname,attr"`
	Initials string `xml:"initials,attr"`
	Surname  string `xml:"surname,attr"`
}

// RefDate is the reference date. See RefAuthor.
type RefDate struct {
	Year  string `xml:"year,attr,omitempty"`
	Month string `xml:"month,attr,omitempty"`
	Day   string `xml:"day,attr,omitempty"`
}

// RefFront the reference <front>. See RefAuthor.
type RefFront struct {
	Title  string    `xml:"title"`
	Author RefAuthor `xml:"author"`
	Date   RefDate   `xml:"date"`
}

// RefFormat is the reference format. See RefAuthor.
type RefFormat struct {
	Typ    string `xml:"type,attr,omitempty"`
	Target string `xml:"target,attr"`
}

// RefXML is the entire structure. See RefAuthor.
type RefXML struct {
	Anchor string    `xml:"anchor,attr"`
	Front  RefFront  `xml:"front"`
	Format RefFormat `xml:"format"`
}

func (options *html) References(out *bytes.Buffer, citations map[string]*citation) {
	if options.flags&HTML_COMPLETE_PAGE == 0 {
		return
	}
	if len(citations) == 0 {
		return
	}
	options.ial = &inlineAttr{class: map[string]bool{"bibliography": true}}
	options.Header(out, func() bool { out.WriteString("Bibliography"); return true }, 1, "bibliography")
	out.WriteString("<ol class=\"bibliography\">\n")

	// [1] Haskell Authors. Haskell.  http://www.haskell.org/ , 1990
	// <span id=anchor>[x]</span>
	for anchor, cite := range citations {
		if len(cite.xml) > 0 {
			var ref RefXML
			if e := xmllib.Unmarshal(cite.xml, &ref); e != nil {
				log.Printf("mmark: failed to unmarshal reference: `%s': %s", anchor, e)
				continue
			}
			out.WriteString("<li class=\"bibliography\" id=\"" + ref.Anchor + "\">\n")
			out.WriteString("  " + "<span class=\"bibliography-details\">" + ref.Front.Author.Fullname + ". ")
			out.WriteString(ref.Front.Title + ". ")
			out.WriteString("<a href=\"" + ref.Format.Target + "\">" + ref.Format.Target + "</a>\n")
			out.WriteString("  " + ref.Front.Date.Year + ".</span>\n")
			out.WriteString("</li>\n")
		}
	}
	out.WriteString("</ol>\n")
}

func (options *html) NormalText(out *bytes.Buffer, text []byte) {
	attrEscape(out, text)
}

func (options *html) DocumentHeader(out *bytes.Buffer, first bool) {
	if !first {
		return
	}
	if options.flags&HTML_COMPLETE_PAGE == 0 {
		return
	}

	out.WriteString("<!DOCTYPE html>\n")
	out.WriteString("<html>\n")
}

func (options *html) DocumentFooter(out *bytes.Buffer, first bool) {
	if !first {
		return
	}
	idx := make(map[string]*bytes.Buffer)
	idxSlice := []string{}
	if len(options.index) > 0 {
		out.WriteString("<div class=\"index\">\n")
		for k, v := range options.index {
			prim := false
			if _, ok := idx[k.primary]; !ok {
				idx[k.primary] = new(bytes.Buffer)
				idxSlice = append(idxSlice, k.primary)
				prim = true
			}
			buf := idx[k.primary]
			if prim {
				buf.WriteString("<span class=\"index-ref-primary\">" + k.primary + "</span>\n")
			}
			if len(k.secondary) == 0 {
				// if k.secondary is empty we should write the pointers here, because they are meant for
				// the primary
				buf.WriteString("<span class=\"index-ref-space\"> </span>")
				for i, r := range v {
					buf.WriteString("<a class=\"index-ref-ref\" href=\"" + r + "\">" + strconv.Itoa(i+1) + "</a>")
					if i+1 < len(v) {
						buf.WriteByte(',')
					}
				}
				buf.WriteString("\n")
				continue
			}

			buf.WriteString("<span class=\"index-ref-secondary\">" + k.secondary + "</span>")
			buf.WriteString("<span class=\"index-ref-space\"> </span>")
			for i, r := range v {
				buf.WriteString("<a class=\"index-ref-ref\" href=\"" + r + "\">" + strconv.Itoa(i+1) + "</a>")
				if i+1 < len(v) {
					buf.WriteByte(',')
				}
			}
			buf.WriteString("\n")
		}
		sort.Strings(idxSlice)
		options.ial = &inlineAttr{class: map[string]bool{"index": true}}
		options.Header(out, func() bool { out.WriteString("Index"); return true }, 1, "index-ref-index")
		char := ""
		for _, s := range idxSlice {
			if char != string(s[0]) {
				out.WriteString("<h3 class=\"index-ref-char\">" + string(s[0]) + "</h3>\n")
			}
			out.Write(idx[s].Bytes())
			char = string(s[0])
		}
		out.WriteString("</div>")
	}

	if options.flags&HTML_COMPLETE_PAGE != 0 {
		out.WriteString("\n</body>\n")
		out.WriteString("</html>\n")
	}
}

func (options *html) DocumentMatter(out *bytes.Buffer, matter int) {
	if matter == _DOC_BACK_MATTER {
		options.appendix = true
	}
}

func (options *html) TocHeaderWithAnchor(text []byte, level int, anchor string) {
	for level > options.currentLevel {
		switch {
		case bytes.HasSuffix(options.toc.Bytes(), []byte("</li>\n")):
			// this sublist can nest underneath a header
			size := options.toc.Len()
			options.toc.Truncate(size - len("</li>\n"))

		case options.currentLevel > 0:
			options.toc.WriteString("<li>")
		}
		if options.toc.Len() > 0 {
			options.toc.WriteByte('\n')
		}
		options.toc.WriteString("<ul>\n")
		options.currentLevel++
	}

	for level < options.currentLevel {
		options.toc.WriteString("</ul>")
		if options.currentLevel > 1 {
			options.toc.WriteString("</li>\n")
		}
		options.currentLevel--
	}

	options.toc.WriteString("<li><a href=\"#")
	if anchor != "" {
		options.toc.WriteString(anchor)
	} else {
		options.toc.WriteString("toc_")
		options.toc.WriteString(strconv.Itoa(options.headerCount))
	}
	options.toc.WriteString("\">")
	options.headerCount++

	options.toc.Write(text)

	options.toc.WriteString("</a></li>\n")
}

func (options *html) TocHeader(text []byte, level int) {
	options.TocHeaderWithAnchor(text, level, "")
}

func (options *html) TocFinalize() {
	for options.currentLevel > 1 {
		options.toc.WriteString("</ul></li>\n")
		options.currentLevel--
	}

	if options.currentLevel > 0 {
		options.toc.WriteString("</ul>\n")
	}
}

func (options *html) SetInlineAttr(i *inlineAttr) {
	options.ial = i
}

func (options *html) inlineAttr() *inlineAttr {
	if options.ial == nil {
		return newInlineAttr()
	}
	return options.ial
}

func isHtmlTag(tag []byte, tagname string) bool {
	found, _ := findHtmlTagPos(tag, tagname)
	return found
}

// Look for a character, but ignore it when it's in any kind of quotes, it
// might be JavaScript
func skipUntilCharIgnoreQuotes(html []byte, start int, char byte) int {
	inSingleQuote := false
	inDoubleQuote := false
	inGraveQuote := false
	i := start
	for i < len(html) {
		switch {
		case html[i] == char && !inSingleQuote && !inDoubleQuote && !inGraveQuote:
			return i
		case html[i] == '\'':
			inSingleQuote = !inSingleQuote
		case html[i] == '"':
			inDoubleQuote = !inDoubleQuote
		case html[i] == '`':
			inGraveQuote = !inGraveQuote
		}
		i++
	}
	return start
}

func findHtmlTagPos(tag []byte, tagname string) (bool, int) {
	i := 0
	if i < len(tag) && tag[0] != '<' {
		return false, -1
	}
	i++
	i = skipSpace(tag, i)

	if i < len(tag) && tag[i] == '/' {
		i++
	}

	i = skipSpace(tag, i)
	j := 0
	for ; i < len(tag); i, j = i+1, j+1 {
		if j >= len(tagname) {
			break
		}

		if strings.ToLower(string(tag[i]))[0] != tagname[j] {
			return false, -1
		}
	}

	if i == len(tag) {
		return false, -1
	}

	rightAngle := skipUntilCharIgnoreQuotes(tag, i, '>')
	if rightAngle > i {
		return true, rightAngle
	}

	return false, -1
}

func skipUntilChar(text []byte, start int, char byte) int {
	i := start
	for i < len(text) && text[i] != char {
		i++
	}
	return i
}

func skipSpace(tag []byte, i int) int {
	for i < len(tag) && isspace(tag[i]) {
		i++
	}
	return i
}

func doubleSpace(out *bytes.Buffer) {
	if out.Len() > 0 {
		out.WriteByte('\n')
	}
}

func isRelativeLink(link []byte) (yes bool) {
	yes = false

	// a tag begin with '#' or '.'
	if link[0] == '#' || link[0] == '.' {
		yes = true
	}

	// link begin with '/' but not '//', the second maybe a protocol relative link
	if len(link) >= 2 && link[0] == '/' && link[1] != '/' {
		yes = true
	}

	// only the root '/'
	if len(link) == 1 && link[0] == '/' {
		yes = true
	}
	return
}
