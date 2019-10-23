package markdown

import (
	"github.com/microcosm-cc/bluemonday"
	gfm "github.com/shurcooL/github_flavored_markdown"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

// Render renders Markdown content into sanitized HTML that is safe to render anywhere.
func Render(content string) string {
	unsafeHTML := gfm.Markdown([]byte(content))

	p := bluemonday.UGCPolicy()
	p.AllowAttrs("name").Matching(bluemonday.SpaceSeparatedTokens).OnElements("a")
	p.AllowAttrs("rel").Matching(lazyregexp.New(`^nofollow$`)).OnElements("a")
	p.AllowAttrs("class").Matching(lazyregexp.New(`^anchor$`)).OnElements("a")
	p.AllowAttrs("aria-hidden").Matching(lazyregexp.New(`^true$`)).OnElements("a")
	p.AllowAttrs("type").Matching(lazyregexp.New(`^checkbox$`)).OnElements("input")
	p.AllowAttrs("checked", "disabled").Matching(lazyregexp.New(`^$`)).OnElements("input")
	p.AllowAttrs("class").Matching(lazyregexp.New("^language-[a-zA-Z0-9]+$")).OnElements("code")
	return string(p.SanitizeBytes(unsafeHTML))
}
