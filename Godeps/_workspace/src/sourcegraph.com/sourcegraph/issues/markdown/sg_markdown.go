package markdown

import (
	"bytes"
	"html/template"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

// Sourcegraph flavored markdown
func Parse(md string) template.HTML {
	//l := link([]byte(md))
	b := markdown([]byte(md))
	s := sanitize(b)
	return template.HTML(string(s))
}

func linkIssues(text []byte, baseUri string) []byte {
	num := string(text[1:])
	a := "<a href=\"" + baseUri + "/" + num + "\">" + "#" + num + "</a>"
	return []byte(a)
}

// This is quadratic (in commitList*size of md) and kinda terrible
func limkCommits(maybeCommit []byte, commitList [][]byte) []byte {
	for _, commit := range commitList {
		if bytes.HasPrefix(commit, maybeCommit[1:]) {
			return []byte(string(maybeCommit[0]) + "COMMIT#" + string(maybeCommit[1:]))
		}
	}
	return maybeCommit
}

/*func link(md []byte, baseUri string) []byte {
	issueID := regexp.MustCompile("#[0-9]+")
	issueLinked := issueID.ReplaceAllFunc(md, func(t []byte) []byte {
		return linkIssues(t, baseUri)
	})
	commitID := regexp.MustCompile(`(\s|^)([0-9]|[a-f]){7,40}`)
	commits := make([][]byte, 0)
	commits = append(commits, []byte("1234567890abcdef"))
	linked := commitID.ReplaceAllFunc(issueLinked, func(mc []byte) []byte {
		return limkCommits(mc, commits)
	})
	return linked
}*/

func markdown(m []byte) []byte {
	r := blackfriday.HtmlRenderer(0, "", "")
	var extensions int
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS
	extensions |= blackfriday.EXTENSION_HARD_LINE_BREAK
	b := blackfriday.Markdown(m, r, extensions)
	return b
}

func sanitize(b []byte) []byte {
	p := bluemonday.UGCPolicy()
	s := p.SanitizeBytes(b)
	return s
}
