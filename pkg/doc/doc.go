// Package doc contains utilities for formatting code documentation
// strings.
//
// TODO: It should probably be moved to util/, or made
// server-side-only (i.e., moved underneath server/internal).
package doc

import (
	"bytes"
	"html"
	"path/filepath"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

type Formatter string

const (
	Markdown         Formatter = "markdown"
	ReStructuredText Formatter = "rst"
	Text             Formatter = "text"
)

// ExtensionFormat maps a lowercase file extension (beginning with ".", such as
// ".md") to the formatter that should be used.
var ExtensionFormat = map[string]Formatter{
	".md":       Markdown,
	".markdown": Markdown,
	".mdown":    Markdown,
	".rdoc":     Markdown, // TODO(sqs): actually implement RDoc
	".rst":      ReStructuredText,
}

// Format returns the doc formatter to use for the given filename. It determines
// it based on the file extension and defaults to plain text.
func Format(filename string) Formatter {
	ext := strings.ToLower(filepath.Ext(filename))
	if fmt, present := ExtensionFormat[ext]; present {
		return fmt
	}
	return Text
}

// IsFormattableDocFile determines whether the supplied file name is a doc file
// format that can be rendered into formatted HTML (e.g. Markdown).
func IsFormattableDocFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, present := ExtensionFormat[ext]
	if !present {
		return false
	}
	return true
}

// ToHTML converts a source document in format to HTML. If conversion fails, it
// returns a failsafe plaintext-to-HTML conversion and a non-nil error.
func ToHTML(formatter Formatter, src []byte) ([]byte, error) {
	var out []byte
	var err error

	switch formatter {
	case Markdown:
		// Some README.md files use "~~~" instead of "```" for delimiting code
		// blocks. But "~~~" is not supported by blackfriday, so hackily replace
		// the former with the latter. See, e.g., the code blocks at
		// https://raw.githubusercontent.com/go-martini/martini/de643861770082784ad14cba4557ad68568dcc7b/README.md.
		src = bytes.Replace(src, []byte("\n~~~"), []byte("\n```"), -1)

		// Blackfriday requires a blank line before a ``` code block. :(
		src = bytes.Replace(src, []byte("\n```"), []byte("\n\n```"), -1)

		htmlFlags := blackfriday.HTML_USE_XHTML |
			blackfriday.HTML_USE_SMARTYPANTS |
			blackfriday.HTML_SMARTYPANTS_FRACTIONS |
			blackfriday.HTML_SMARTYPANTS_LATEX_DASHES |
			blackfriday.HTML_HREF_TARGET_BLANK |
			blackfriday.HTML_NOFOLLOW_LINKS |
			blackfriday.HTML_NOREFERRER_LINKS
		commonExtensions := blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
			blackfriday.EXTENSION_TABLES |
			blackfriday.EXTENSION_FENCED_CODE |
			blackfriday.EXTENSION_AUTOLINK |
			blackfriday.EXTENSION_STRIKETHROUGH |
			blackfriday.EXTENSION_HEADER_IDS |
			blackfriday.EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK
		renderer := addBootstrapClassMarkdownRenderer{blackfriday.HtmlRenderer(htmlFlags, "", "")}
		out = blackfriday.Markdown([]byte(src), renderer, commonExtensions)

	case ReStructuredText:
		// Not implemented, handled below.

	case Text:
		// Wrap in <pre> below.
	}
	if err != nil || len(out) == 0 {
		out = []byte("<pre>" + strings.TrimSpace(html.EscapeString(string(src))) + "</pre>")
	}

	// Hackily escape HTML IDs. Fixes
	// https://github.com/sourcegraph/sourcegraph.com/issues/104.
	//
	// TODO(sqs): do this correctly by changing blackfriday or by
	// parsing the resulting HTML.
	out = bytes.Replace(out, []byte(`<div class="section" id="`), []byte(`<div class="section" id="doc-`), -1)

	// Sanitize the HTML, always.
	out = bluemonday.UGCPolicy().SanitizeBytes(out)

	return out, err
}

func StripNulls(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\x00' {
			return -1
		}
		return r
	}, s)
}

// ChooseReadme chooses the best file to use as the "main" readme
// file. It prefers shorter names (like README and README.md) over
// longer names (which are often localized readmes, like
// README.en.md). If no suitable README file is found, the empty
// string is returned.
func ChooseReadme(files []string) string {
	var shortest string
	for _, file := range files {
		if strings.HasPrefix(strings.ToLower(file), "readme") && (shortest == "" || len(shortest) > len(file)) {
			shortest = file
		}
	}
	return shortest
}

type addBootstrapClassMarkdownRenderer struct {
	blackfriday.Renderer
}

func (r addBootstrapClassMarkdownRenderer) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {
	out.WriteString("<table class=\"table table-striped table-bordered\">\n<thead>\n")
	out.Write(header)
	out.WriteString("</thead>\n\n<tbody>\n")
	out.Write(body)
	out.WriteString("</tbody>\n</table>\n")
}
