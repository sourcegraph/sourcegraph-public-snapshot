// Package gist5258650 splits a multiline string into a slice of lines.
package gist5258650

import (
	"strings"
)

func GetLines(s string) []string {
	return strings.Split(s, "\n")
}

func GetLine(s string, LineIndex int) string {
	return GetLines(s)[LineIndex]
}

func main() {
	str := `First Line,
2nd Line.

This is actually 4th line (not 3rd)!
The index of this line is 4 (aka it's the 5th line).`

	// Test GetLines() and GetLine()
	for i := range GetLines(str) {
		println(i, GetLine(str, i))
	}
}
