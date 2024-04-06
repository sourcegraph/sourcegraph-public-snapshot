package search

import "strings"

// truncateLines truncates the lines in content to be at most maxLineLen and
// includes a truncation suffix.
func truncateLines(content string, maxLineLen int) (_ string, truncated bool) {
	// Skip if disabled or impossible to have a line longer than maxLineLen.
	if maxLineLen <= 0 || len(content) < maxLineLen {
		return content, false
	}

	// Before doing allocations, check if every line is short enough.
	if allLineLenLessThanEqual(content, maxLineLen) {
		return content, false
	}

	truncateSuffix := "...truncated"
	if len(truncateSuffix) > maxLineLen {
		truncateSuffix = ""
	}

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if len(line) <= maxLineLen {
			continue
		}

		lines[i] = line[:maxLineLen-len(truncateSuffix)] + truncateSuffix
	}

	return strings.Join(lines, "\n"), true
}

func allLineLenLessThanEqual(s string, l int) bool {
	for len(s) > 0 {
		idx := strings.IndexByte(s, '\n')
		if idx < 0 {
			return len(s) <= l
		} else if idx > l {
			return false
		}
		s = s[idx+1:]
	}
	return true
}
