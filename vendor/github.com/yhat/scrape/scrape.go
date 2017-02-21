// Package scrape provides a searching api on top of golang.org/x/net/html
package scrape

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Matcher should return true when a desired node is found.
type Matcher func(node *html.Node) bool

// FindAll returns all nodes which match the provided Matcher. After discovering a matching
// node, it will _not_ discover matching subnodes of that node.
func FindAll(node *html.Node, matcher Matcher) []*html.Node {
	return findAllInternal(node, matcher, false)
}

// FindAllNested returns all nodes which match the provided Matcher and _will_ discover
// matching subnodes of matching nodes.
func FindAllNested(node *html.Node, matcher Matcher) []*html.Node {
	return findAllInternal(node, matcher, true)
}

// Find returns the first node which matches the matcher using depth-first search.
// If no node is found, ok will be false.
//
//     root, err := html.Parse(resp.Body)
//     if err != nil {
//         // handle error
//     }
//     matcher := func(n *html.Node) bool {
//         return n.DataAtom == atom.Body
//     }
//     body, ok := scrape.Find(root, matcher)
func Find(node *html.Node, matcher Matcher) (n *html.Node, ok bool) {
	if matcher(node) {
		return node, true
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		n, ok := Find(c, matcher)
		if ok {
			return n, true
		}
	}
	return nil, false
}

// FindParent searches up HTML tree from the current node until either a
// match is found or the top is hit.
func FindParent(node *html.Node, matcher Matcher) (n *html.Node, ok bool) {
	for p := node.Parent; p != nil; p = p.Parent {
		if matcher(p) {
			return p, true
		}
	}
	return nil, false
}

// Text returns text from all descendant text nodes joined.
// For control over the join function, see TextJoin.
func Text(node *html.Node) string {
	joiner := func(s []string) string {
		n := 0
		for i := range s {
			trimmed := strings.TrimSpace(s[i])
			if trimmed != "" {
				s[n] = trimmed
				n++
			}
		}
		return strings.Join(s[:n], " ")
	}
	return TextJoin(node, joiner)
}

// TextJoin returns a string from all descendant text nodes joined by a
// caller provided join function.
func TextJoin(node *html.Node, join func([]string) string) string {
	nodes := FindAll(node, func(n *html.Node) bool { return n.Type == html.TextNode })
	parts := make([]string, len(nodes))
	for i, n := range nodes {
		parts[i] = n.Data
	}
	return join(parts)
}

// Attr returns the value of an HTML attribute.
func Attr(node *html.Node, key string) string {
	for _, a := range node.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// ByTag returns a Matcher which matches all nodes of the provided tag type.
//
//     root, err := html.Parse(resp.Body)
//     if err != nil {
//         // handle error
//     }
//     title, ok := scrape.Find(root, scrape.ByTag(atom.Title))
func ByTag(a atom.Atom) Matcher {
	return func(node *html.Node) bool { return node.DataAtom == a }
}

// ById returns a Matcher which matches all nodes with the provided id.
func ById(id string) Matcher {
	return func(node *html.Node) bool { return Attr(node, "id") == id }
}

// ByClass returns a Matcher which matches all nodes with the provided class.
func ByClass(class string) Matcher {
	return func(node *html.Node) bool {
		classes := strings.Fields(Attr(node, "class"))
		for _, c := range classes {
			if c == class {
				return true
			}
		}
		return false
	}
}

// findAllInternal encapsulates the node tree traversal
func findAllInternal(node *html.Node, matcher Matcher, searchNested bool) []*html.Node {
	matched := []*html.Node{}

	if matcher(node) {
		matched = append(matched, node)

		if !searchNested {
			return matched
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		found := findAllInternal(c, matcher, searchNested)
		if len(found) > 0 {
			matched = append(matched, found...)
		}
	}
	return matched
}

// Find returns the first node which matches the matcher using next sibling search.
// If no node is found, ok will be false.
//
//     root, err := html.Parse(resp.Body)
//     if err != nil {
//         // handle error
//     }
//     matcher := func(n *html.Node) bool {
//         return n.DataAtom == atom.Body
//     }
//     body, ok := scrape.FindNextSibling(root, matcher)
func FindNextSibling(node *html.Node, matcher Matcher) (n *html.Node, ok bool) {

	for s := node.NextSibling; s != nil; s = s.NextSibling {
		if matcher(s) {
			return s, true
		}
	}
	return nil, false
}

// Find returns the first node which matches the matcher using previous sibling search.
// If no node is found, ok will be false.
//
//     root, err := html.Parse(resp.Body)
//     if err != nil {
//         // handle error
//     }
//     matcher := func(n *html.Node) bool {
//         return n.DataAtom == atom.Body
//     }
//     body, ok := scrape.FindPrevSibling(root, matcher)
func FindPrevSibling(node *html.Node, matcher Matcher) (n *html.Node, ok bool) {
	for s := node.PrevSibling; s != nil; s = s.PrevSibling {
		if matcher(s) {
			return s, true
		}
	}
	return nil, false
}
