package markdown

import (
	"bytes"
	"fmt"
	"regexp" //nolint:depguard // bluemonday requires this pkg
	"sync"

	chroma "github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	once     sync.Once
	policy   *bluemonday.Policy
	renderer goldmark.Markdown
)

// Render renders Markdown content into sanitized HTML that is safe to render anywhere.
func Render(content string) (string, error) {
	once.Do(func() {
		policy = bluemonday.UGCPolicy()
		policy.AllowAttrs("name").Matching(bluemonday.SpaceSeparatedTokens).OnElements("a")
		policy.AllowAttrs("rel").Matching(regexp.MustCompile(`^nofollow$`)).OnElements("a")
		policy.AllowAttrs("class").Matching(regexp.MustCompile(`^anchor$`)).OnElements("a")
		policy.AllowAttrs("aria-hidden").Matching(regexp.MustCompile(`^true$`)).OnElements("a")
		policy.AllowAttrs("type").Matching(regexp.MustCompile(`^checkbox$`)).OnElements("input")
		policy.AllowAttrs("checked", "disabled").Matching(regexp.MustCompile(`^$`)).OnElements("input")
		policy.AllowAttrs("class").Matching(regexp.MustCompile("^(?:chroma-[a-zA-Z0-9\\-]+)|chroma$")).OnElements("pre", "code", "span")

		origTypes := chroma.StandardTypes
		sourcegraphTypes := map[chroma.TokenType]string{}
		for k, v := range origTypes {
			if k == chroma.PreWrapper {
				sourcegraphTypes[k] = v
			} else {
				sourcegraphTypes[k] = fmt.Sprintf("chroma-%s", v)
			}
		}
		chroma.StandardTypes = sourcegraphTypes

		renderer = goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				highlighting.NewHighlighting(
					highlighting.WithFormatOptions(
						chromahtml.WithClasses(true),
						chromahtml.WithLineNumbers(false),
					),
				),
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				// HTML sanitization is handled by bluemonday
				html.WithUnsafe(),
			),
		)
	})

	var buf bytes.Buffer
	if err := renderer.Convert([]byte(content), &buf); err != nil {
		return "", err
	}
	return string(policy.SanitizeBytes(buf.Bytes())), nil
}
