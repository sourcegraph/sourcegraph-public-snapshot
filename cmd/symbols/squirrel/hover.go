package squirrel

import (
	"math"
	"strings"
)

// Returns the markdown hover message for the given node if it exists.
func findHover(node Node) string {
	style := node.LangSpec.commentStyle

	hover := ""
	hover += "```" + style.codeFenceName + "\n"
	hover += strings.Split(string(node.Contents), "\n")[node.StartPoint().Row] + "\n"
	hover += "```"

	for cur := node.Node; cur != nil && cur.StartPoint().Row == node.StartPoint().Row; cur = cur.Parent() {
		prev := cur.PrevNamedSibling()

		// Skip over Java annotations and the like.
		for ; prev != nil; prev = prev.PrevNamedSibling() {
			if !contains(style.skipNodeTypes, prev.Type()) {
				break
			}
		}

		// Collect comments backwards.
		comments := []string{}
		lastStartRow := -1
		for ; prev != nil && contains(style.nodeTypes, prev.Type()); prev = prev.PrevNamedSibling() {
			if lastStartRow == -1 {
				lastStartRow = int(prev.StartPoint().Row)
			} else if lastStartRow != int(prev.EndPoint().Row+1) {
				break
			} else {
				lastStartRow = int(prev.StartPoint().Row)
			}

			comment := prev.Content(node.Contents)

			// Strip line noise and delete garbage lines.
			lines := []string{}
			allLines := strings.Split(comment, "\n")
			for _, line := range allLines {
				if style.ignoreRegex != nil && style.ignoreRegex.MatchString(line) {
					continue
				}

				if style.stripRegex != nil {
					line = style.stripRegex.ReplaceAllString(line, "")
				}

				lines = append(lines, line)
			}

			// Remove shared leading spaces.
			spaces := math.MaxInt32
			for _, line := range lines {
				spaces = min(spaces, len(line)-len(strings.TrimLeft(line, " ")))
			}
			for i := range lines {
				lines[i] = strings.TrimLeft(lines[i], " ")
			}

			// Join lines.
			comments = append(comments, strings.Join(lines, "\n"))
		}

		if len(comments) == 0 {
			continue
		}

		// Reverse comments
		for i, j := 0, len(comments)-1; i < j; i, j = i+1, j-1 {
			comments[i], comments[j] = comments[j], comments[i]
		}

		hover = hover + "\n\n---\n\n" + strings.Join(comments, "\n") + "\n"
	}

	return strings.ToValidUTF8(hover, "�")
}
