// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Godoc comment extraction and comment -> HTML formatting.

package doc

import (
	"io"
	"regexp"
	"strings"
	"text/template" // for HTMLEscape
	"unicode"
	"unicode/utf8"
)

// Escape comment text for HTML. If nice is set,
// also turn `` into &ldquo; and '' into &rdquo;.
func commentEscape(w io.Writer, text string, nice bool, opts fmtOpts) {
	last := 0
	if nice {
		for i := 0; i < len(text)-1; i++ {
			ch := text[i]
			if ch == text[i+1] && (ch == '`' || ch == '\'') {
				if opts.ShouldHTMLEscape {
					template.HTMLEscape(w, []byte(text[last:i]))
				} else {
					w.Write([]byte(text[last:i]))
				}
				last = i + 2
				switch ch {
				case '`':
					w.Write(opts.Backtick)
				case '\'':
					w.Write(opts.Singlequote)
				}
				i++ // loop will add one more
			}
		}
	}
	if opts.ShouldHTMLEscape {
		template.HTMLEscape(w, []byte(text[last:]))
	} else {
		w.Write([]byte(text[last:]))
	}
}

const (
	// Regexp for Go identifiers
	identRx = `[\pL_][\pL_0-9]*`

	// Regexp for URLs
	protocol = `https?|ftp|file|gopher|mailto|news|nntp|telnet|wais|prospero`
	hostPart = `[a-zA-Z0-9_@\-]+`
	filePart = `[a-zA-Z0-9_?%#~&/\-+=()]+` // parentheses may not be matching; see pairedParensPrefixLen
	urlRx    = `(` + protocol + `)://` +   // http://
		hostPart + `([.:]` + hostPart + `)*/?` + // //www.google.com:8080/
		filePart + `([:.,]` + filePart + `)*`
)

var matchRx = regexp.MustCompile(`(` + urlRx + `)|(` + identRx + `)`)

type fmtOpts struct {
	ShouldHTMLEscape                       bool
	URLBegin, URLBeginEnd, URLEnd          []byte
	ItalicBegin, ItalicEnd                 []byte
	ParagraphBegin, ParagraphEnd           []byte
	PreBegin, PreEnd                       []byte
	HeaderBegin, HeaderBeginEnd, HeaderEnd []byte
	Backtick, Singlequote                  []byte
}

var htmlFmtOpts = fmtOpts{
	ShouldHTMLEscape: true,
	URLBegin:         []byte(`<a href="`),
	URLBeginEnd:      []byte(`">`),
	URLEnd:           []byte("</a>"),
	ItalicBegin:      []byte("<i>"),
	ItalicEnd:        []byte("</i>"),
	ParagraphBegin:   []byte("<p>\n"),
	ParagraphEnd:     []byte("</p>\n"),
	PreBegin:         []byte("<pre>"),
	PreEnd:           []byte("</pre>\n"),
	HeaderBegin:      []byte(`<h3 id="`),
	HeaderBeginEnd:   []byte(`">`),
	HeaderEnd:        []byte("</h3>\n"),
	Backtick:         []byte("&ldquo;"),
	Singlequote:      []byte("&rdquo;"),
}

var markdownFmtOpts = fmtOpts{
	ShouldHTMLEscape: false,
	URLBegin:         []byte("["),
	URLBeginEnd:      []byte("]("),
	URLEnd:           []byte(")"),
	ItalicBegin:      []byte("_"),
	ItalicEnd:        []byte("_"),
	ParagraphBegin:   []byte(""),
	ParagraphEnd:     []byte("\n\n"),
	PreBegin:         []byte("```\n"),
	PreEnd:           []byte("\n```\n"),
	HeaderBegin:      []byte("### "),
	HeaderBeginEnd:   []byte(""),
	HeaderEnd:        []byte("\n"),
	Backtick:         []byte("`"),
	Singlequote:      []byte("'"),
}

// pairedParensPrefixLen returns the length of the longest prefix of s containing paired parentheses.
func pairedParensPrefixLen(s string) int {
	parens := 0
	l := len(s)
	for i, ch := range s {
		switch ch {
		case '(':
			if parens == 0 {
				l = i
			}
			parens++
		case ')':
			parens--
			if parens == 0 {
				l = len(s)
			} else if parens < 0 {
				return i
			}
		}
	}
	return l
}

// Emphasize and escape a line of text for HTML. URLs are converted into links;
// if the URL also appears in the words map, the link is taken from the map (if
// the corresponding map value is the empty string, the URL is not converted
// into a link). Go identifiers that appear in the words map are italicized; if
// the corresponding map value is not the empty string, it is considered a URL
// and the word is converted into a link. If nice is set, the remaining text's
// appearance is improved where it makes sense (e.g., `` is turned into &ldquo;
// and '' into &rdquo;).
func emphasize(w io.Writer, line string, words map[string]string, nice bool, opts fmtOpts) {
	for {
		m := matchRx.FindStringSubmatchIndex(line)
		if m == nil {
			break
		}
		// m >= 6 (two parenthesized sub-regexps in matchRx, 1st one is urlRx)

		// write text before match
		commentEscape(w, line[0:m[0]], nice, opts)

		// adjust match if necessary
		match := line[m[0]:m[1]]
		if n := pairedParensPrefixLen(match); n < len(match) {
			// match contains unpaired parentheses (rare);
			// redo matching with shortened line for correct indices
			m = matchRx.FindStringSubmatchIndex(line[:m[0]+n])
			match = match[:n]
		}

		// analyze match
		url := ""
		italics := false
		if words != nil {
			url, italics = words[match]
		}
		if m[2] >= 0 {
			// match against first parenthesized sub-regexp; must be match against urlRx
			if !italics {
				// no alternative URL in words list, use match instead
				url = match
			}
			italics = false // don't italicize URLs
		}

		// write match
		if len(url) > 0 {
			w.Write(opts.URLBegin)
			if opts.ShouldHTMLEscape {
				template.HTMLEscape(w, []byte(url))
			} else {
				w.Write([]byte(url))
			}
			w.Write(opts.URLBeginEnd)
		}
		if italics {
			w.Write(opts.ItalicBegin)
		}
		commentEscape(w, match, nice, opts)
		if italics {
			w.Write(opts.ItalicEnd)
		}
		if len(url) > 0 {
			w.Write(opts.URLEnd)
		}

		// advance
		line = line[m[1]:]
	}
	commentEscape(w, line, nice, opts)
}

func indentLen(s string) int {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return i
}

func isBlank(s string) bool {
	return len(s) == 0 || (len(s) == 1 && s[0] == '\n')
}

func commonPrefix(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	return a[0:i]
}

func unindent(block []string) {
	if len(block) == 0 {
		return
	}

	// compute maximum common white prefix
	prefix := block[0][0:indentLen(block[0])]
	for _, line := range block {
		if !isBlank(line) {
			prefix = commonPrefix(prefix, line[0:indentLen(line)])
		}
	}
	n := len(prefix)

	// remove
	for i, line := range block {
		if !isBlank(line) {
			block[i] = line[n:]
		}
	}
}

// heading returns the trimmed line if it passes as a section heading;
// otherwise it returns the empty string.
func heading(line string) string {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return ""
	}

	// a heading must start with an uppercase letter
	r, _ := utf8.DecodeRuneInString(line)
	if !unicode.IsLetter(r) || !unicode.IsUpper(r) {
		return ""
	}

	// it must end in a letter or digit:
	r, _ = utf8.DecodeLastRuneInString(line)
	if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
		return ""
	}

	// exclude lines with illegal characters
	if strings.ContainsAny(line, ",.;:!?+*/=()[]{}_^°&§~%#@<\">\\") {
		return ""
	}

	// allow "'" for possessive "'s" only
	for b := line; ; {
		i := strings.IndexRune(b, '\'')
		if i < 0 {
			break
		}
		if i+1 >= len(b) || b[i+1] != 's' || (i+2 < len(b) && b[i+2] != ' ') {
			return "" // not followed by "s "
		}
		b = b[i+2:]
	}

	return line
}

type op int

const (
	opPara op = iota
	opHead
	opPre
)

type block struct {
	op    op
	lines []string
}

var nonAlphaNumRx = regexp.MustCompile(`[^a-zA-Z0-9]`)

func anchorID(line string) string {
	// Add a "hdr-" prefix to avoid conflicting with IDs used for package symbols.
	return "hdr-" + nonAlphaNumRx.ReplaceAllString(line, "_")
}

// ToHTML converts comment text to formatted HTML.
// The comment was prepared by DocReader,
// so it is known not to have leading, trailing blank lines
// nor to have trailing spaces at the end of lines.
// The comment markers have already been removed.
//
// Each span of unindented non-blank lines is converted into
// a single paragraph. There is one exception to the rule: a span that
// consists of a single line, is followed by another paragraph span,
// begins with a capital letter, and contains no punctuation
// is formatted as a heading.
//
// A span of indented lines is converted into a <pre> block,
// with the common indent prefix removed.
//
// URLs in the comment text are converted into links; if the URL also appears
// in the words map, the link is taken from the map (if the corresponding map
// value is the empty string, the URL is not converted into a link).
//
// Go identifiers that appear in the words map are italicized; if the corresponding
// map value is not the empty string, it is considered a URL and the word is converted
// into a link.
func ToHTML(w io.Writer, text string, words map[string]string) {
	toFormatted(w, text, words, htmlFmtOpts)
}

// ToMarkdown is just like ToHTML, except it renders Markdown output instead
// of HTML output.
func ToMarkdown(w io.Writer, text string, words map[string]string) {
	toFormatted(w, text, words, markdownFmtOpts)
}

func toFormatted(w io.Writer, text string, words map[string]string, opts fmtOpts) {
	for _, b := range blocks(text) {
		switch b.op {
		case opPara:
			w.Write(opts.ParagraphBegin)
			for _, line := range b.lines {
				emphasize(w, strings.Replace(line, "\n", " ", -1), words, true, opts)
			}
			w.Write(opts.ParagraphEnd)
		case opHead:
			w.Write(opts.HeaderBegin)
			id := ""
			for _, line := range b.lines {
				if id == "" {
					id = anchorID(line)
					w.Write([]byte(id))
					w.Write(opts.HeaderBeginEnd)
				}
				commentEscape(w, line, true, opts)
			}
			if id == "" {
				w.Write(opts.HeaderBeginEnd)
			}
			w.Write(opts.HeaderEnd)
		case opPre:
			w.Write(opts.PreBegin)
			for _, line := range b.lines {
				emphasize(w, line, nil, false, opts)
			}
			w.Write(opts.PreEnd)
		}
	}
}

func blocks(text string) []block {
	var (
		out  []block
		para []string

		lastWasBlank   = false
		lastWasHeading = false
	)

	close := func() {
		if para != nil {
			out = append(out, block{opPara, para})
			para = nil
		}
	}

	lines := strings.SplitAfter(text, "\n")
	unindent(lines)
	for i := 0; i < len(lines); {
		line := lines[i]
		if isBlank(line) {
			// close paragraph
			close()
			i++
			lastWasBlank = true
			continue
		}
		if indentLen(line) > 0 {
			// close paragraph
			close()

			// count indented or blank lines
			j := i + 1
			for j < len(lines) && (isBlank(lines[j]) || indentLen(lines[j]) > 0) {
				j++
			}
			// but not trailing blank lines
			for j > i && isBlank(lines[j-1]) {
				j--
			}
			pre := lines[i:j]
			i = j

			unindent(pre)

			// put those lines in a pre block
			out = append(out, block{opPre, pre})
			lastWasHeading = false
			continue
		}

		if lastWasBlank && !lastWasHeading && i+2 < len(lines) &&
			isBlank(lines[i+1]) && !isBlank(lines[i+2]) && indentLen(lines[i+2]) == 0 {
			// current line is non-blank, surrounded by blank lines
			// and the next non-blank line is not indented: this
			// might be a heading.
			if head := heading(line); head != "" {
				close()
				out = append(out, block{opHead, []string{head}})
				i += 2
				lastWasHeading = true
				continue
			}
		}

		// open paragraph
		lastWasBlank = false
		lastWasHeading = false
		para = append(para, lines[i])
		i++
	}
	close()

	return out
}

// ToText prepares comment text for presentation in textual output.
// It wraps paragraphs of text to width or fewer Unicode code points
// and then prefixes each line with the indent. In preformatted sections
// (such as program text), it prefixes each non-blank line with preIndent.
func ToText(w io.Writer, text string, indent, preIndent string, width int) {
	l := lineWrapper{
		out:    w,
		width:  width,
		indent: indent,
	}
	for _, b := range blocks(text) {
		switch b.op {
		case opPara:
			// l.write will add leading newline if required
			for _, line := range b.lines {
				l.write(line)
			}
			l.flush()
		case opHead:
			w.Write(nl)
			for _, line := range b.lines {
				l.write(line + "\n")
			}
			l.flush()
		case opPre:
			w.Write(nl)
			for _, line := range b.lines {
				if isBlank(line) {
					w.Write([]byte("\n"))
				} else {
					w.Write([]byte(preIndent))
					w.Write([]byte(line))
				}
			}
		}
	}
}

type lineWrapper struct {
	out       io.Writer
	printed   bool
	width     int
	indent    string
	n         int
	pendSpace int
}

var nl = []byte("\n")
var space = []byte(" ")

func (l *lineWrapper) write(text string) {
	if l.n == 0 && l.printed {
		l.out.Write(nl) // blank line before new paragraph
	}
	l.printed = true

	for _, f := range strings.Fields(text) {
		w := utf8.RuneCountInString(f)
		// wrap if line is too long
		if l.n > 0 && l.n+l.pendSpace+w > l.width {
			l.out.Write(nl)
			l.n = 0
			l.pendSpace = 0
		}
		if l.n == 0 {
			l.out.Write([]byte(l.indent))
		}
		l.out.Write(space[:l.pendSpace])
		l.out.Write([]byte(f))
		l.n += l.pendSpace + w
		l.pendSpace = 1
	}
}

func (l *lineWrapper) flush() {
	if l.n == 0 {
		return
	}
	l.out.Write(nl)
	l.pendSpace = 0
	l.n = 0
}
