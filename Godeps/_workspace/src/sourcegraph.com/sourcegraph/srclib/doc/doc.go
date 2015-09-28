package doc

import (
	"bytes"
	"html"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

type Formatter string

const (
	Text             Formatter = "text"
	Markdown         Formatter = "markdown"
	ReStructuredText Formatter = "rst"
)

// ExtensionFormat maps a lowercase file extension (beginning with ".", such as
// ".md") to the formatter that should be used.
var ExtensionFormat = map[string]Formatter{
	".md":       Markdown,
	".markdown": Markdown,
	".mdown":    Markdown,
	".rdoc":     Markdown, // TODO(sqs): actually implement RDoc
	".txt":      Text,
	".text":     Text,
	"":          Text,
	".ascii":    Text,
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
		out = blackfriday.MarkdownCommon([]byte(src))

	case ReStructuredText:
		out, err = ReStructuredTextToHTML(src)

	case Text:
		// wrap in <pre> below
	}
	if err != nil || len(out) == 0 {
		out = []byte("<pre>" + strings.TrimSpace(html.EscapeString(string(src))) + "</pre>")
	}
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
