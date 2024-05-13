package headings

import "unicode"

// SanitizeHeadingID returns a sanitized anchor name for the given text.
//
// Copied from https://sourcegraph.com/github.com/gomarkdown/markdown@663e2500819c19ed2d3f4bf955931b16fa9adf63/-/blob/parser/block.go?L83-104
func SanitizeHeadingID(text string) string {
	var anchorName []rune
	var futureDash = false
	for _, r := range text {
		switch {
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			if futureDash && len(anchorName) > 0 {
				anchorName = append(anchorName, '-')
			}
			futureDash = false
			anchorName = append(anchorName, unicode.ToLower(r))
		default:
			futureDash = true
		}
	}
	if len(anchorName) == 0 {
		return "empty"
	}
	return string(anchorName)
}
