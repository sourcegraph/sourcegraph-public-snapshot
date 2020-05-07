package graphqlbackend

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type highlightedRange struct {
	line      int32
	character int32
	length    int32
}

func (h *highlightedRange) Line() int32      { return h.line }
func (h *highlightedRange) Character() int32 { return h.character }
func (h *highlightedRange) Length() int32    { return h.length }

type highlightedString struct {
	value      string
	highlights []*highlightedRange
}

func (s *highlightedString) Value() string                   { return s.value }
func (s *highlightedString) Highlights() []*highlightedRange { return s.highlights }

func fromVCSHighlights(vcsHighlights []git.Highlight) []*highlightedRange {
	highlights := make([]*highlightedRange, len(vcsHighlights))
	for i, vh := range vcsHighlights {
		highlights[i] = &highlightedRange{
			line:      int32(vh.Line),
			character: int32(vh.Character),
			length:    int32(vh.Length),
		}
	}
	return highlights
}

func ParseLinesFromHighlight(input string) (map[int32]string, error) {
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return nil, err
	}

	lines := make(map[int32]string)

	body := doc.FirstChild.LastChild // html->body
	pre := body.FirstChild
	if pre == nil || pre.Type != html.ElementNode || pre.DataAtom != atom.Pre {
		return nil, fmt.Errorf("expected html->body->pre, found %+v", pre)
	}

	var next = pre.FirstChild
	var line int32 = 1
	var codeCell *html.Node = &html.Node{Type: html.ElementNode, DataAtom: atom.Div, Data: atom.Div.String()}

	newRow := func() error {
		var buf bytes.Buffer
		err = html.Render(&buf, codeCell)
		if err != nil {
			return err
		}
		lines[line] = buf.String()
		line++
		codeCell = &html.Node{Type: html.ElementNode, DataAtom: atom.Div, Data: atom.Div.String()}
		return nil
	}
	for next != nil {
		nextSibling := next.NextSibling
		switch {
		case next.Type == html.ElementNode && next.DataAtom == atom.Span:
			// Found a span, so add it to our current code cell td.
			next.Parent = nil
			next.PrevSibling = nil
			next.NextSibling = nil
			codeCell.AppendChild(next)

			// Scan the children for text nodes containing new lines so that we
			// can create new table rows.
			if next.FirstChild != nil {
				nextChild := next.FirstChild
				for nextChild != nil {
					switch {
					case nextChild.Type == html.TextNode:
						// Text node, create a new table row for each newline.
						newlines := strings.Count(nextChild.Data, "\n")
						for i := 0; i < newlines; i++ {
							err = newRow()
							if err != nil {
								return nil, err
							}
						}
					default:
						return nil, fmt.Errorf("unexpected HTML child structure (encountered %+v)", nextChild)
					}
					nextChild = nextChild.NextSibling
				}
			}
		case next.Type == html.TextNode:
			// Text node, create a new table row for each newline.
			newlines := strings.Count(next.Data, "\n")
			for i := 0; i < newlines; i++ {
				err = newRow()
				if err != nil {
					return nil, err
				}
			}
		default:
			return nil, fmt.Errorf("unexpected HTML structure (encountered %+v)", next)
		}
		// If the last element in the tree was no text node, need to create one more line from the existing content.
		if nextSibling == nil && next.Type != html.TextNode {
			err = newRow()
			if err != nil {
				return nil, err
			}
		}
		next = nextSibling
	}
	return lines, nil
}
