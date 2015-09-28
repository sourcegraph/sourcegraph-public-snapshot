// XML2RFC v2 rendering backend

package mmark

import (
	"bytes"
	"log"
	"strconv"
	"time"
)

// References code in Xml2rfcv3.go

// XML renderer configuration options.
const (
	XML2_STANDALONE = 1 << iota // create standalone document
)

// <meta name="GENERATOR" content="Blackfriday Markdown Processor v1.0" />

// Xml2 is a type that implements the Renderer interface for XML2RFV3 output.
//
// Do not create this directly, instead use the Xml2Renderer function.
type xml2 struct {
	flags          int  // XML2_* options
	sectionLevel   int  // current section level
	docLevel       int  // frontmatter/mainmatter or backmatter
	part           bool // parts cannot nest, if true a part has been opened
	specialSection int  // are we in a special section
	paraInList     bool // subsequent paras in lists are faked with vspace

	// store the IAL we see for this block element
	ial *inlineAttr

	// titleBlock in TOML
	titleBlock *title

	// (@good) example list group counter
	group map[string]int
}

// Xml2Renderer creates and configures a Xml object, which
// satisfies the Renderer interface.
//
// flags is a set of XML2_* options ORed together
func Xml2Renderer(flags int) Renderer {
	anchorOrID = "anchor"
	return &xml2{flags: flags, group: make(map[string]int)}
}
func (options *xml2) Flags() int { return options.flags }
func (options *xml2) State() int { return 0 }

func (options *xml2) SetInlineAttr(i *inlineAttr) {
	options.ial = i
}

func (options *xml2) inlineAttr() *inlineAttr {
	if options.ial == nil {
		return newInlineAttr()
	}
	return options.ial
}

// render code chunks using verbatim, or listings if we have a language
func (options *xml2) BlockCode(out *bytes.Buffer, text []byte, lang string, caption []byte, subfigure, callout bool) {
	ial := options.inlineAttr()
	ial.GetOrDefaultAttr("align", "center")

	prefix := ial.Value("prefix")
	ial.DropAttr("prefix")  // it's a fake attribute, so drop it
	ial.DropAttr("callout") // it's a fake attribute, so drop it
	// subfigure stuff. TODO(miek): check
	if len(caption) > 0 {
		ial.GetOrDefaultAttr("title", string(sanitizeXML(caption)))
	}
	ial.DropAttr("type")
	s := ial.String()

	out.WriteString("\n<figure" + s + "><artwork" + ial.Key("align") + ">\n")
	if prefix != "" {
		nl := bytes.Count(text, []byte{'\n'})
		text = bytes.Replace(text, []byte{'\n'}, []byte("\n"+prefix), nl-1)
		// add prefix at the start as well
		text = append([]byte(prefix), text...)
	}
	if callout {
		attrEscapeInCode(options, out, text)
	} else {
		writeEntity(out, text)
	}
	out.WriteString("</artwork></figure>\n")
}

func (options *xml2) CalloutCode(out *bytes.Buffer, index, id string) {
	// Should link to id
	attrEscape(out, []byte("<"))
	out.WriteString(index)
	attrEscape(out, []byte(">"))
	return
}

func (options *xml2) CalloutText(out *bytes.Buffer, index string, id []string) {
	out.WriteByte('(')
	for i, k := range id {
		out.WriteString(k)
		if i < len(id)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteByte(')')
}

func (options *xml2) TitleBlockTOML(out *bytes.Buffer, block *title) {
	if options.flags&XML2_STANDALONE == 0 {
		return
	}
	options.titleBlock = block
	out.WriteString("<rfc ipr=\"" +
		options.titleBlock.Ipr + "\" category=\"" +
		options.titleBlock.Category + "\" docName=\"" + options.titleBlock.DocName + "\">\n")

	// Default processing instructions
	out.WriteString("<?rfc toc=\"yes\"?>\n")
	out.WriteString("<?rfc symrefs=\"yes\"?>\n")
	out.WriteString("<?rfc sortrefs=\"yes\"?>\n")
	out.WriteString("<?rfc compact=\"yes\"?>\n")
	out.WriteString("<?rfc subcompact=\"no\"?>\n")

	out.WriteString("<front>\n")
	out.WriteString("<title abbrev=\"" + options.titleBlock.Abbrev + "\">")
	out.WriteString(options.titleBlock.Title + "</title>\n\n")

	for _, a := range options.titleBlock.Author {
		out.WriteString("<author")
		out.WriteString(" initials=\"" + a.Initials + "\"")
		out.WriteString(" surname=\"" + a.Surname + "\"")
		out.WriteString(" fullname=\"" + a.Fullname + "\">\n")

		out.WriteString("<organization>" + a.Organization + "</organization>\n")
		out.WriteString("<address>\n")

		out.WriteString("<postal>\n")
		out.WriteString("<street>" + a.Address.Postal.Street + "</street>\n") // street is a list?
		out.WriteString("<city>" + a.Address.Postal.City + "</city>\n")
		out.WriteString("<code>" + a.Address.Postal.Code + "</code>\n")
		out.WriteString("<country>" + a.Address.Postal.Country + "</country>\n")
		out.WriteString("</postal>\n")

		out.WriteString("<email>" + a.Address.Email + "</email>\n")
		out.WriteString("<uri>" + a.Address.Uri + "</uri>\n")

		out.WriteString("</address>\n")
		out.WriteString("</author>\n")
	}

	year := ""
	if options.titleBlock.Date.Year() > 0 {
		year = " year=\"" + strconv.Itoa(options.titleBlock.Date.Year()) + "\""
	}
	month := ""
	if options.titleBlock.Date.Month() > 0 {
		month = " month=\"" + time.Month(options.titleBlock.Date.Month()).String() + "\""
	}
	day := ""
	if options.titleBlock.Date.Day() > 0 {
		day = " day=\"" + strconv.Itoa(options.titleBlock.Date.Day()) + "\""
	}
	out.WriteString("<date" + year + month + day + "/>\n\n")

	out.WriteString("<area>" + options.titleBlock.Area + "</area>\n")
	out.WriteString("<workgroup>" + options.titleBlock.Workgroup + "</workgroup>\n")
	for _, k := range options.titleBlock.Keyword {
		out.WriteString("<keyword>" + k + "</keyword>\n")
	}
	out.WriteString("\n")
}

func (options *xml2) BlockQuote(out *bytes.Buffer, text []byte, attribution []byte) {
	// Fake a list paragraph

	// TODO(miek): IAL, clear them for now
	options.inlineAttr()

	out.WriteString("<t><list style=\"empty\">\n")
	out.Write(text)

	// check for "person -- URI" syntax use those if found
	// need to strip tags because these are attributes
	// TODO(miek): better
	if len(attribution) != 0 {
		parts := bytes.Split(attribution, []byte("-- "))
		// TODO(miek): 1 part
		if len(parts) > 0 {
			cite := bytes.TrimSpace(parts[0])
			var quotedFrom []byte
			if len(parts) == 2 {
				quotedFrom = sanitizeXML(bytes.TrimSpace(parts[1]))
			}
			out.WriteString("<t>--\n")
			out.Write(cite)
			if len(parts) == 2 {
				out.WriteString(", ")
				out.Write(quotedFrom)
			}
			out.WriteString("</t>\n")
		}
	}
	out.WriteString("</list></t>\n")
}

func (options *xml2) Aside(out *bytes.Buffer, text []byte) {
	options.BlockQuote(out, text, nil)
}

func (options *xml2) Note(out *bytes.Buffer, text []byte) {
	options.BlockQuote(out, text, nil)
}

func (options *xml2) CommentHtml(out *bytes.Buffer, text []byte) {
	// nothing fancy any left of the first `:` will be used as the source="..."
	i := bytes.Index(text, []byte("-->"))
	if i > 0 {
		text = text[:i]
	}
	// strip, <!--
	text = text[4:]

	var source []byte
	l := len(text)
	if l > 20 {
		l = 20
	}
	for i := 0; i < l; i++ {
		if text[i] == '-' && text[i+1] == '-' {
			source = text[:i]
			text = text[i+2:]
			break
		}
	}
	// don't output a cref if it is not name: remark
	if len(source) != 0 {
		source = bytes.TrimSpace(source)
		text = bytes.TrimSpace(text)
		out.WriteString("<t><cref source=\"")
		out.Write(source)
		out.WriteString("\">")
		out.Write(text)
		out.WriteString("</cref></t>\n")
	}
	return
}

func (options *xml2) BlockHtml(out *bytes.Buffer, text []byte) {
	// not supported, don't know yet if this is useful
	return
}

func (options *xml2) Part(out *bytes.Buffer, text func() bool, id string) {}

func (options *xml2) SpecialHeader(out *bytes.Buffer, what []byte, text func() bool, id string) {
	if string(what) == "preface" {
		log.Printf("mmark: handling preface like abstract")
		what = []byte("abstract")
	}
	level := 1
	if level <= options.sectionLevel {
		// close previous ones
		for i := options.sectionLevel - level + 1; i > 0; i-- {
			out.WriteString("</section>\n")
		}
	}

	ial := options.inlineAttr()

	out.WriteString("\n<abstract" + ial.String() + ">\n")
	options.sectionLevel = 0
	options.specialSection = _ABSTRACT
	return
}

func (options *xml2) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	switch options.specialSection {
	case _ABSTRACT:
		out.WriteString("</abstract>\n\n")
	}

	if level > options.sectionLevel+1 {
		log.Printf("mmark: section jump from H%d to H%d, id: \"%s\"", options.sectionLevel, level, id)
	}

	if level <= options.sectionLevel {
		// close previous ones
		for i := options.sectionLevel - level + 1; i > 0; i-- {
			out.WriteString("</section>\n")
		}
	}

	ial := options.inlineAttr()
	ial.GetOrDefaultId(id)
	ial.KeepAttr([]string{"title", "toc"})
	ial.KeepClass(nil)

	// new section
	out.WriteString("\n<section" + ial.String())
	out.WriteString(" title=\"")
	text()
	out.WriteString("\">\n")
	options.sectionLevel = level
	options.specialSection = 0
	return
}

func (options *xml2) HRule(out *bytes.Buffer) {
	// not used
}

func (options *xml2) List(out *bytes.Buffer, text func() bool, flags, start int, group []byte) {
	marker := out.Len()
	// inside lists we must drop the paragraph
	if flags&_LIST_INSIDE_LIST == 0 {
		out.WriteString("<t>\n")
	}

	ial := options.inlineAttr()
	ial.KeepAttr([]string{"style", "counter"})
	// start > 1 is not supported

	// for group, fake a numbered format (if not already given and put a
	// group -> current number in options

	switch {
	case flags&_LIST_TYPE_ORDERED != 0:
		switch {
		case flags&_LIST_TYPE_ORDERED_ALPHA_LOWER != 0:
			ial.GetOrDefaultAttr("style", "format %c")
		case flags&_LIST_TYPE_ORDERED_ALPHA_UPPER != 0:
			ial.GetOrDefaultAttr("style", "format %C")
		case flags&_LIST_TYPE_ORDERED_ROMAN_LOWER != 0:
			ial.GetOrDefaultAttr("style", "format %i")
		case flags&_LIST_TYPE_ORDERED_ROMAN_UPPER != 0:
			ial.GetOrDefaultAttr("style", "format %I")
		case flags&_LIST_TYPE_ORDERED_GROUP != 0:

			if group != nil {
				// don't think we need ++ this.
				options.group[string(group)]++
				ial.GetOrDefaultAttr("counter", string(group))
				ial.GetOrDefaultAttr("style", "format (%d)")
			}
		default:
			ial.GetOrDefaultAttr("style", "numbers")
		}
	case flags&_LIST_TYPE_DEFINITION != 0:
		ial.GetOrDefaultAttr("style", "hanging")
	default:
		ial.GetOrDefaultAttr("style", "symbols")
	}

	out.WriteString("<list" + ial.String() + ">\n")

	if !text() {
		out.Truncate(marker)
		return
	}
	switch {
	case flags&_LIST_TYPE_ORDERED != 0:
		out.WriteString("</list>\n")
	case flags&_LIST_TYPE_DEFINITION != 0:
		out.WriteString("</t>\n</list>\n")
	default:
		out.WriteString("</list>\n")
	}
	if flags&_LIST_INSIDE_LIST == 0 {
		out.WriteString("</t>\n")
	}
}

func (options *xml2) ListItem(out *bytes.Buffer, text []byte, flags int) {
	if flags&_LIST_TYPE_DEFINITION != 0 && flags&_LIST_TYPE_TERM == 0 {
		//out.WriteString("<dd>")
		out.Write(text)
		//out.WriteString("</dd>\n")
		return
	}
	if flags&_LIST_TYPE_TERM != 0 {
		if flags&_LIST_ITEM_BEGINNING_OF_LIST == 0 {
			out.WriteString("</t>\n")
		}
		// close previous one?/
		out.WriteString("<t hangText=\"")
		writeSanitizeXML(out, text)
		out.WriteString("\">\n")
		return
	}
	out.WriteString("<t>")
	out.Write(text)
	out.WriteString("</t>\n")
	options.paraInList = false
}

func (options *xml2) Example(out *bytes.Buffer, index int) {
	out.WriteByte('(')
	out.WriteString(strconv.Itoa(index))
	out.WriteByte(')')
}

func (options *xml2) Paragraph(out *bytes.Buffer, text func() bool, flags int) {
	marker := out.Len()
	if flags&_LIST_TYPE_DEFINITION == 0 && flags&_LIST_INSIDE_LIST == 0 {
		out.WriteString("<t>")
	} else {
		if options.paraInList && flags&_LIST_ITEM_BEGINNING_OF_LIST != 0 {
			out.WriteString("<vspace blankLines=\"1\" />\n")
		}
	}
	if !text() {
		out.Truncate(marker)
		return
	}
	if marker+3 == out.Len() { // empty paragraph, suppress
		out.Truncate(marker)
		return
	}
	out.WriteByte('\n')
	if flags&_LIST_TYPE_DEFINITION == 0 && flags&_LIST_INSIDE_LIST == 0 {
		out.WriteString("</t>\n")
	} else {
		options.paraInList = true
	}
}

func (options *xml2) Math(out *bytes.Buffer, text []byte, display bool) {

}

func (options *xml2) Table(out *bytes.Buffer, header []byte, body []byte, footer []byte, columnData []int, caption []byte) {
	ial := options.inlineAttr()
	if caption != nil {
		ial.GetOrDefaultAttr("title", string(sanitizeXML(caption)))
	}

	s := ial.String()
	out.WriteString("<texttable" + s + ">\n")
	out.Write(header)
	out.Write(body)
	out.Write(footer)
	out.WriteString("</texttable>\n")
}

func (options *xml2) TableRow(out *bytes.Buffer, text []byte) {
	out.Write(text)
	out.WriteString("\n")
}

func (options *xml2) TableHeaderCell(out *bytes.Buffer, text []byte, align int) {
	a := ""
	switch align {
	case _TABLE_ALIGNMENT_LEFT:
		a = " align=\"left\""
	case _TABLE_ALIGNMENT_RIGHT:
		a = " align=\"right\""
	default:
		a = " align=\"center\""
	}
	out.WriteString("<ttcol" + a + ">")
	writeSanitizeXML(out, text)
	out.WriteString("</ttcol>\n")
}

func (options *xml2) TableCell(out *bytes.Buffer, text []byte, align int) {
	out.WriteString("<c>")
	out.Write(text)
	out.WriteString("</c>")
}

func (options *xml2) Footnotes(out *bytes.Buffer, text func() bool) {
	// not used
}

func (options *xml2) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	// not used
}

func (options *xml2) Index(out *bytes.Buffer, primary, secondary []byte, prim bool) {
	p := ""
	if prim {
		p = " primary=\"true\""
	}
	out.WriteString("<iref item=\"" + string(primary) + "\"" + p)
	out.WriteString(" subitem=\"" + string(secondary) + "\"" + "/>")
}

func (options *xml2) Citation(out *bytes.Buffer, link, title []byte) {
	if len(title) == 0 {
		out.WriteString("<xref target=\"" + string(link) + "\"/>")
		return
	}
	out.WriteString("<xref target=\"" + string(link) + "\"/>")
}

func (options *xml2) References(out *bytes.Buffer, citations map[string]*citation) {
	if options.flags&XML2_STANDALONE == 0 {
		return
	}
	// close any option section tags
	for i := options.sectionLevel; i > 0; i-- {
		out.WriteString("</section>\n")
		options.sectionLevel--
	}
	switch options.docLevel {
	case _DOC_FRONT_MATTER:
		out.WriteString("</front>\n")
		out.WriteString("<back>\n")
	case _DOC_MAIN_MATTER:
		out.WriteString("</middle>\n")
		out.WriteString("<back>\n")
	case _DOC_BACK_MATTER:
		// nothing to do
	}
	options.docLevel = _DOC_BACK_MATTER

	refi, refn := 0, 0
	for _, c := range citations {
		if c.typ == 'i' {
			refi++
		}
		if c.typ == 'n' {
			refn++
		}
	}
	// output <xi:include href="<references file>.xml"/>, we use file it its not empty, otherwise
	// we construct one for RFCNNNN and I-D.something something.
	if refi+refn > 0 {
		if refn > 0 {
			out.WriteString("<references title=\"Normative References\">\n")
			for _, c := range citations {
				if c.typ == 'n' {
					if c.xml != nil {
						out.Write(c.xml)
						out.WriteByte('\n')
						continue
					}
					f := referenceFile(c)
					out.WriteString("<?rfc include=\"" + f + "\"?>\n")
				}
			}
			out.WriteString("</references>\n")
		}
		if refi > 0 {
			out.WriteString("<references title=\"Informative References\">\n")
			for _, c := range citations {
				if c.typ == 'i' {
					// if we have raw xml, output that
					if c.xml != nil {
						out.Write(c.xml)
						out.WriteByte('\n')
						continue
					}
					f := referenceFile(c)
					out.WriteString("<?rfc include=\"" + f + "\"?>\n")
				}
			}
			out.WriteString("</references>\n")
		}
	}
}

func (options *xml2) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	out.WriteString("<eref target=\"")
	if kind == _LINK_TYPE_EMAIL {
		out.WriteString("mailto:")
	}
	out.Write(link)
	out.WriteString("\"/>")
}

func (options *xml2) CodeSpan(out *bytes.Buffer, text []byte) {
	out.WriteString("<spanx style=\"verb\">")
	writeEntity(out, text)
	out.WriteString("</spanx>")
}

func (options *xml2) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	// Check for 2119 Keywords, strip emphasis from them.
	s := string(text)
	if _, ok := words2119[s]; ok {
		out.Write(text)
		return
	}
	out.WriteString("<spanx style=\"strong\">")
	out.Write(text)
	out.WriteString("</spanx>")
}

func (options *xml2) Emphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<spanx style=\"emph\">")
	out.Write(text)
	out.WriteString("</spanx>")
}

func (options *xml2) Subscript(out *bytes.Buffer, text []byte) {
	// There is no subscript
	out.WriteByte('~')
	out.Write(text)
	out.WriteByte('~')
}

func (options *xml2) Superscript(out *bytes.Buffer, text []byte) {
	// There is no superscript
	out.WriteByte('^')
	out.Write(text)
	out.WriteByte('^')
}

func (options *xml2) Figure(out *bytes.Buffer, text []byte, caption []byte) {
	// the caption is discarded here.
	out.Write(text)
}

func (options *xml2) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte, subfigure bool) {
	// convert to url or image wrapped in figure
	ial := options.inlineAttr()
	ial.GetOrDefaultAttr("align", "center")
	ial.DropAttr("type") // type may be set, but is not valid in xml 2 syntax
	s := options.inlineAttr().String()
	if len(title) != 0 {
		out.WriteString("<figure" + s + " title=\"")
		title1 := sanitizeXML(title)
		out.Write(title1)
		out.WriteString("\">\n")
		// empty artwork
		out.WriteString("<artwork" + ial.Key("align") + ">" + string(link) + "</artwork>\n")
		out.WriteString("<postamble>")
	}
	out.WriteString("<eref target=\"")
	out.Write(link)
	out.WriteString("\">")
	out.Write(alt)
	out.WriteString("</eref>")
	if len(title) != 0 {
		out.WriteString("</postamble>\n")
		out.WriteString("</figure>\n")
	}
}

func (options *xml2) LineBreak(out *bytes.Buffer) {
	out.WriteString("\n<vspace/>\n")
}

func (options *xml2) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	if link[0] == '#' {
		out.WriteString("<xref target=\"")
		out.Write(link[1:])
		out.WriteString("\"/>")
		return
	}
	out.WriteString("<eref target=\"")
	out.Write(link)
	out.WriteString("\">")
	out.Write(content)
	out.WriteString("</eref>")
}

func (options *xml2) Abbreviation(out *bytes.Buffer, abbr, title []byte) {
	out.Write(abbr)
}

func (options *xml2) RawHtmlTag(out *bytes.Buffer, tag []byte) {}

func (options *xml2) TripleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<spanx style=\"strong\"><spanx style=\"emph\">")
	out.Write(text)
	out.WriteString("</spanx></spanx>")
}

func (options *xml2) StrikeThrough(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (options *xml2) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	// not used
}

func (options *xml2) Entity(out *bytes.Buffer, entity []byte) {
	out.Write(entity)
}

func (options *xml2) NormalText(out *bytes.Buffer, text []byte) {
	attrEscape(out, text)
}

// header and footer
func (options *xml2) DocumentHeader(out *bytes.Buffer, first bool) {
	if !first || options.flags&XML2_STANDALONE == 0 {
		return
	}
	out.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	out.WriteString("<!DOCTYPE rfc SYSTEM 'rfc2629.dtd' []>\n")
}

func (options *xml2) DocumentFooter(out *bytes.Buffer, first bool) {
	if !first {
		return
	}
	switch options.specialSection {
	case _ABSTRACT:
		out.WriteString("</abstract>\n\n")
	}
	// close any option section tags
	for i := options.sectionLevel; i > 0; i-- {
		out.WriteString("</section>\n")
		options.sectionLevel--
	}
	if options.flags&XML2_STANDALONE == 0 {
		return
	}
	switch options.docLevel {
	case _DOC_FRONT_MATTER:
		out.WriteString("\n</front>\n")
	case _DOC_MAIN_MATTER:
		out.WriteString("\n</middle>\n")
	case _DOC_BACK_MATTER:
		out.WriteString("\n</back>\n")
	}
	out.WriteString("</rfc>\n")
}

func (options *xml2) DocumentMatter(out *bytes.Buffer, matter int) {
	if options.flags&XML2_STANDALONE == 0 {
		return
	}
	switch options.specialSection {
	case _ABSTRACT:
		out.WriteString("</abstract>\n\n")
	}
	// we default to frontmatter already openened in the documentHeader
	for i := options.sectionLevel; i > 0; i-- {
		out.WriteString("</section>\n")
		options.sectionLevel--
	}
	switch matter {
	case _DOC_FRONT_MATTER:
		// already open
	case _DOC_MAIN_MATTER:
		out.WriteString("</front>\n")
		out.WriteString("\n<middle>\n")
	case _DOC_BACK_MATTER:
		out.WriteString("\n</middle>\n")
		out.WriteString("<back>\n")
	}
	options.docLevel = matter
	options.specialSection = 0
}
