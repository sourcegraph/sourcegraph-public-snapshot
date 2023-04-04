package markdown

import (
	"fmt"
	"regexp" //nolint:depguard // bluemonday requires this pkg
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
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
		policy.AllowAttrs("class").Matching(regexp.MustCompile(`^(?:chroma-[a-zA-Z0-9\-]+)|chroma$`)).OnElements("pre", "code", "span")
		policy.AllowAttrs("align").OnElements("img", "p")
		policy.AllowElements("picture", "video", "track", "source")
		policy.AllowAttrs("srcset", "src", "type", "media", "width", "height", "sizes").OnElements("source")
		policy.AllowAttrs("playsinline", "muted", "autoplay", "loop", "controls", "width", "height", "poster", "src").OnElements("video")
		policy.AllowAttrs("src", "kind", "srclang", "default", "label").OnElements("track")
		policy.AddTargetBlankToFullyQualifiedLinks(true)

		html.LinkAttributeFilter.Add([]byte("aria-hidden"))
		html.LinkAttributeFilter.Add([]byte("name"))

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
				parser.WithASTTransformers(util.Prioritized(mdTransformFunc(mdLinkHeaders), 1)),
			),
			goldmark.WithRendererOptions(
				// HTML sanitization is handled by bluemonday
				html.WithUnsafe(),
			),
		)
	})

	var buf strings.Builder
	if err := renderer.Convert([]byte(content), &buf); err != nil {
		return "", err
	}
	return policy.Sanitize(buf.String()), nil
}

type mdTransformFunc func(*ast.Document, text.Reader, parser.Context)

func (f mdTransformFunc) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	f(node, reader, pc)
}

func mdLinkHeaders(doc *ast.Document, _ text.Reader, _ parser.Context) {
	mdWalk(doc)
}

func mdWalk(n ast.Node) {
	switch n := n.(type) {
	case *ast.Heading:
		id, ok := n.AttributeString("id")
		if !ok {
			return
		}

		var idStr string
		switch id := id.(type) {
		case []byte:
			idStr = string(id)
		case string:
			idStr = id
		default:
			return
		}

		anchorLink := ast.NewLink()
		anchorLink.Destination = []byte("#" + idStr)
		anchorLink.SetAttributeString("class", []byte("anchor"))
		anchorLink.SetAttributeString("rel", []byte("nofollow"))
		anchorLink.SetAttributeString("aria-hidden", []byte("true"))
		anchorLink.SetAttributeString("name", id)

		n.InsertBefore(n, n.FirstChild(), anchorLink)
		return
	}
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		mdWalk(child)
	}
}
