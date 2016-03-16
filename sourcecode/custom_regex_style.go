package sourcecode

import (
	"html/template"
	"log"
	"sort"
	"strings"
	"unicode"

	"github.com/sourcegraph/annotate"
	"golang.org/x/net/html"
	"sourcegraph.com/sourcegraph/sourcegraph/ui/uiconf"
	"sourcegraph.com/sourcegraph/sourcegraph/util"
)

// Ability to override def rendering style via regexp flags.

var styles = []struct {
	flag     *util.RegexpFlag
	cssClass string
}{
	{&uiconf.Flags.DefQualifiedNameHide, "def-qualified-name-hide"},
	{&uiconf.Flags.DefQualifiedNameDim, "def-qualified-name-dim"},
	{&uiconf.Flags.DefQualifiedNameBold, "def-qualified-name-bold"},
}

// OverrideStyleViaRegexpFlags accepts the original HTML segment for
// a Definition Qualified Name, and optionally overrides the style
// as specified via regexp flags. If no such flags are set, then original
// HTML is returned unmodified. If an error is encountered during processing,
// original HTML is returned and the error is logged.
func OverrideStyleViaRegexpFlags(originalHTML template.HTML) template.HTML {
	if !anyStyleRegexpFlagsSet() {
		return originalHTML
	}
	qualifiedName, err := processStyleRegexp(originalHTML)
	if err != nil {
		log.Println("overriding style via regexp flags failed:", err)
		return originalHTML
	}
	return qualifiedName
}

// anyStyleRegexpFlagsSet returns true if any style regexp flag is set.
func anyStyleRegexpFlagsSet() bool {
	for _, s := range styles {
		r := s.flag.Regexp
		if r != nil {
			return true
		}
	}
	return false
}

// processStyleRegexp accepts the original HTML segment for
// a Definition Qualified Name, and overrides the style as specified via regexp flags.
//
// The approach is as follows.
// First, it extracts the text content from original HTML, dropping all markup.
// Second, for each style regexp, it finds matching parts of the text and annotates it.
func processStyleRegexp(originalHTML template.HTML) (template.HTML, error) {
	text, err := extractTextContentFromHTML(originalHTML)
	if err != nil {
		return "", err
	}

	var anns annotate.Annotations
	for _, s := range styles {
		r := s.flag.Regexp
		if r == nil {
			continue
		}

		for _, m := range r.FindAllStringSubmatchIndex(text, -1) {
			switch {
			case len(m) == 2:
				// If there are no capturing groups, use entire expression.
				anns = append(anns, &annotate.Annotation{
					Start: m[0], End: m[1],
					Left: []byte(`<span class="` + s.cssClass + `">`), Right: []byte(`</span>`),
				})
			case len(m) >= 4:
				// If there are capturing groups, skip the entire expression and use all capturing groups.
				for i := 1; 2*i+1 < len(m); i++ {
					anns = append(anns, &annotate.Annotation{
						Start: m[2*i+0], End: m[2*i+1],
						Left: []byte(`<span class="` + s.cssClass + `">`), Right: []byte(`</span>`),
					})
				}
			}
		}
	}
	sort.Sort(anns)

	html, err := annotate.Annotate([]byte(text), anns, template.HTMLEscape)
	if err != nil {
		return "", err
	}
	return template.HTML(html), nil
}

// extractTextContentFromHTML returns the recursive concatenation of the text content of an HTML segment.
//
// Multiple consecutive whitespace are consolidated into a single space.
func extractTextContentFromHTML(segment template.HTML) (string, error) {
	node, err := html.Parse(strings.NewReader(string(segment)))
	if err != nil {
		return "", err
	}
	var text string
	// Consolidate multiple consecutive whitespace into a single space.
	for _, r := range extractText(node) {
		if !unicode.IsSpace(r) {
			text += string(r)
		} else if !strings.HasSuffix(text, " ") {
			text += " "
		}
	}
	return strings.TrimSpace(text), nil
}

// extractText returns the recursive concatenation of the text content of an html node.
func extractText(n *html.Node) string {
	var out string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			out += c.Data
		} else {
			out += extractText(c)
		}
	}
	return out
}
