package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/sourcegraph/annotate"

	"gopkg.in/inconshreveable/log15.v2"
	"gopkg.in/olebedev/go-duktape.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui/reactbridge"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

// A blob represents the React component props of
// sourcegraph/blob/Blob.js.
type blob struct {
	// Props
	Contents               string                      `json:"contents"`
	Annotations            *sourcegraph.AnnotationList `json:"annotations"`
	ActiveDef              string                      `json:"activeDef"`
	StartLine              int                         `json:"startLine"`
	EndLine                int                         `json:"endLine"`
	LineNumbers            bool                        `json:"lineNumbers"`
	HighlightSelectedLines bool                        `json:"highlightSelectedLines"`

	// State
	VisibleLinesCount int `json:"visibleLinesCount"`

	reactID int
}

func (b *blob) render() (string, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `<table class="line-numbered-code" data-reactid="%s">`, b.nextReactID())
	fmt.Fprintf(&buf, `<tbody data-reactid="%s">`, b.nextReactID())

	lines, err := b.lines()
	if err != nil {
		return "", err
	}
	buf.Write(lines)
	fmt.Fprintf(&buf, `</tbody></table>`)
	return buf.String(), nil
}

func (b *blob) lines() ([]byte, error) {
	lines := bytes.Split([]byte(b.Contents), []byte("\n"))
	lineAnns := annotationsByLine(b.Annotations, lines)

	var buf bytes.Buffer
	for i, line := range lines {
		lineNum := i + 1

		classes := []string{"line"}
		if b.HighlightSelectedLines && b.StartLine <= lineNum && b.EndLine >= lineNum {
			classes = append(classes, "main-byte-range")
		}

		fmt.Fprintf(&buf, `<tr class="%s" data-line="%d" data-reactid="%s">`, strings.Join(classes, " "), i+1, b.nextReactID())
		fmt.Fprintf(&buf, `<td class="line-number" data-line="%d" data-reactid="%s"></td>`, i+1, b.nextReactID())
		fmt.Fprintf(&buf, `<td class="line-content" data-reactid="%s">`, b.nextReactID())

		if i >= b.VisibleLinesCount {
			reactCompatibleHTMLEscape(&buf, line)
		} else {
			lineAnns2 := make([]*annotate.Annotation, len(lineAnns[i]))
			lineStartByte := b.Annotations.LineStartBytes[i]
			for j, ann := range lineAnns[i] {
				var left, right string
				var classes []string
				if ann.Class != "" {
					classes = append(classes, ann.Class)
				}
				if ann.URL != "" || len(ann.URLs) > 0 {
					classes = append(classes, "ref")

					var href string
					if len(ann.URLs) > 0 {
						href = ann.URLs[0]
						for _, u := range ann.URLs {
							if u == b.ActiveDef {
								classes = append(classes, "active-def")
							}
						}
					} else if ann.URL != "" {
						href = ann.URL
						if ann.URL == b.ActiveDef {
							classes = append(classes, "active-def")
						}
					}

					left = fmt.Sprintf(`<a class="%s" href="%s">`, strings.Join(classes, " "), href)
					right = "</a>"
				} else if len(classes) > 0 {
					left = fmt.Sprintf(`<span class="%s">`, strings.Join(classes, " "))
					right = "</span>"
				}

				var startByte, endByte int
				if ann.StartByte < lineStartByte {
					startByte = 0
				} else {
					startByte = int(ann.StartByte - lineStartByte)
				}
				if lineEndByte := lineStartByte + uint32(len(line)); ann.EndByte > lineEndByte {
					endByte = len(line)
				} else {
					endByte = int(ann.EndByte - lineStartByte)
				}

				lineAnns2[j] = &annotate.Annotation{
					Start:     startByte,
					End:       endByte,
					Left:      []byte(left),
					Right:     []byte(right),
					WantInner: int(ann.WantInner),
				}
			}

			lineHTML, err := annotate.Annotate(line, lineAnns2, reactCompatibleHTMLEscape)
			if annotate.IsOutOfBounds(err) {
				err = nil
			}
			if err != nil {
				return nil, err
			}

			lineHTML, err = addReactIDs(b.nextReactID, lineHTML)
			if err != nil {
				return nil, err
			}
			buf.Write(lineHTML)
		}

		fmt.Fprintf(&buf, `</td></tr>`)
	}
	return buf.Bytes(), nil
}

func (b *blob) nextReactID() string {
	b.reactID++
	return strconv.Itoa(b.reactID)
}

func addReactIDs(nextReactID func() string, src []byte) ([]byte, error) {
	container := &html.Node{
		Type:     html.ElementNode,
		Data:     "div",
		DataAtom: atom.Div,
	}
	nodes, err := html.ParseFragment(bytes.NewReader(src), container)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		container.AppendChild(n)
	}

	var nodeFn func(*html.Node)
	nodeFn = func(n *html.Node) {
		switch n.Type {
		case html.TextNode:
			// If a text element is by itself, leave as-is. Otherwise
			// surround it with comments.
			if n.PrevSibling != nil || n.NextSibling != nil {
				n.Parent.InsertBefore(&html.Node{
					Type: html.CommentNode,
					Data: fmt.Sprintf(" react-text: %s ", nextReactID()),
				}, n)
				n.Parent.InsertBefore(&html.Node{
					Type: html.CommentNode,
					Data: " /react-text ",
				}, n.NextSibling)
			}

		case html.ElementNode:
			if n != container {
				n.Attr = append(n.Attr, html.Attribute{
					Key: "data-reactid",
					Val: nextReactID(),
				})
			}
			if n.FirstChild != nil {
				nodeFn(n.FirstChild)
			}
		}
		if n.NextSibling != nil {
			nodeFn(n.NextSibling)
		}
	}
	nodeFn(container)

	var buf bytes.Buffer
	for n := container.FirstChild; n != nil; n = n.NextSibling {
		if err := html.Render(&buf, n); err != nil {
			return nil, err
		}
	}

	return convertToReactHTMLEscapeStyle(buf.Bytes()), nil
}

// annotationsByLine returns an array with one entry per line. Each line's entry
// is the array of annotations that intersect that line.
//
// Assumes annotations has been sorted by sortAnns.
//
// NOTE: This must stay in sync with annotationsByLine.js.
func annotationsByLine(anns *sourcegraph.AnnotationList, lines [][]byte) [][]*sourcegraph.Annotation {
	lineAnns := make([][]*sourcegraph.Annotation, len(lines))

	lineEndBytes := make([]uint32, len(anns.LineStartBytes))
	for i, lineStartByte := range anns.LineStartBytes {
		if i == len(anns.LineStartBytes)-1 {
			lineEndBytes[i] = lineStartByte + uint32(len(lines[i]))
		} else {
			lineEndBytes[i] = anns.LineStartBytes[i+1]
		}
	}

	var line int
	for _, ann := range anns.Annotations {
		// Advance (if necessary) to the first line that ann intersects.
		if ann.StartByte >= lineEndBytes[line] {
			for line < len(lines) && ann.StartByte >= lineEndBytes[line] {
				line++
			}
		}
		if line == len(lines) {
			break
		}

		// Optimization: add the ann to this line (if it intersects);
		if ann.StartByte < lineEndBytes[line] && ann.EndByte >= anns.LineStartBytes[line] {
			lineAnns[line] = append(lineAnns[line], ann)
		}

		// Add the ann to all lines (current and subsequent) it intersects.
		if ann.EndByte <= lineEndBytes[line] {
			continue
		}
		for line2 := line + 1; line2 < len(lines); line2++ {
			if ann.StartByte >= lineEndBytes[line2] {
				break
			}
			if ann.StartByte < lineEndBytes[line2] && ann.EndByte >= anns.LineStartBytes[line2] {
				lineAnns[line2] = append(lineAnns[line2], ann)
			}
		}
	}

	return lineAnns
}

func init() {
	reactbridge.GlobalFuncs["__goRenderBlob__"] = func(d *duktape.Context) int {
		nextReactID := d.ToNumber(0)
		propsJSONStr := d.SafeToString(1)

		var blob blob
		blob.reactID = int(nextReactID)
		if err := json.Unmarshal([]byte(propsJSONStr), &blob); err != nil {
			log15.Error("Error unmarshaling Blob props JSON", "error", err)
			return duktape.ErrRetAPI
		}

		componentHTML, err := blob.render()
		if err != nil {
			log15.Error("Error rendering Blob", "error", err)
			return duktape.ErrRetAPI
		}

		d.PushString(componentHTML)
		return 1
	}
}
