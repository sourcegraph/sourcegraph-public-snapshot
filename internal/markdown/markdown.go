package markdown

import (
	"bytes"
	"regexp"
	"sync"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	once   sync.Once
	md     goldmark.Markdown
	policy *bluemonday.Policy
)

// Render renders Markdown content into sanitized HTML that is safe to render anywhere.
func Render(content string) (string, error) {
	once.Do(func() {
		md = goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
				html.WithXHTML(),
			),
		)
		policy = bluemonday.UGCPolicy()
		policy.AllowAttrs("name").Matching(bluemonday.SpaceSeparatedTokens).OnElements("a")
		policy.AllowAttrs("rel").Matching(regexp.MustCompile(`^nofollow$`)).OnElements("a")
		policy.AllowAttrs("class").Matching(regexp.MustCompile(`^anchor$`)).OnElements("a")
		policy.AllowAttrs("aria-hidden").Matching(regexp.MustCompile(`^true$`)).OnElements("a")
		policy.AllowAttrs("type").Matching(regexp.MustCompile(`^checkbox$`)).OnElements("input")
		policy.AllowAttrs("checked", "disabled").Matching(regexp.MustCompile(`^$`)).OnElements("input")
		policy.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
	})

	var buf bytes.Buffer
	err := md.Convert([]byte(content), &buf)
	if err != nil {
		return "", err
	}
	return string(policy.SanitizeBytes(buf.Bytes())), nil
}
