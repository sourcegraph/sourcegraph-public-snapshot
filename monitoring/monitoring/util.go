package monitoring

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/grafana-tools/sdk"
)

// upperFirst returns s with an uppercase first rune.
func upperFirst(s string) string {
	return strings.ToUpper(string([]rune(s)[0])) + string([]rune(s)[1:])
}

// withPeriod returns s ending with a period.
func withPeriod(s string) string {
	if !strings.HasSuffix(s, ".") {
		return s + "."
	}
	return s
}

// stringPtr converts a string value to a pointer, useful for setting fields in some APIs.
func stringPtr(s string) *string {
	return &s
}

// isValidUID checks if the given string is a valid UID for entry into a Grafana dashboard. This is
// primarily used in the URL, e.g. /-/debug/grafana/d/syntect-server/<UID> and allows us to have
// static URLs we can document like:
//
// 	Go to https://sourcegraph.example.com/-/debug/grafana/d/syntect-server/syntect-server
//
// Instead of having to describe all the steps to navigate there because the UID is random.
func isValidUID(s string) bool {
	if s != strings.ToLower(s) {
		return false
	}
	for _, r := range s {
		if !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-') {
			return false
		}
	}
	return true
}

// toMarkdownList converts a Go string into a Markdown list
func toMarkdownList(m string) (string, error) {
	m = strings.TrimPrefix(m, "\n")

	// Replace single quotes with backticks.
	// Replace escaped single quotes with single quotes.
	m = strings.Replace(m, `\'`, `$ESCAPED_SINGLE_QUOTE`, -1)
	m = strings.Replace(m, `'`, "`", -1)
	m = strings.Replace(m, `$ESCAPED_SINGLE_QUOTE`, "'", -1)

	// Unindent based on the indention of the last line.
	lines := strings.Split(m, "\n")
	baseIndention := lines[len(lines)-1]
	if strings.TrimSpace(baseIndention) == "" {
		if strings.Contains(baseIndention, " ") {
			return "", errors.New("go string literal indention must be tabs")
		}
		indentionLevel := strings.Count(baseIndention, "\t")
		removeIndention := strings.Repeat("\t", indentionLevel+1)
		for i, l := range lines[:len(lines)-1] {
			newLine := strings.TrimPrefix(l, removeIndention)
			if l == newLine {
				return "", fmt.Errorf("inconsistent indention (line %d %q expected to start with %q)", i, l, removeIndention)
			}
			lines[i] = newLine
		}
		m = strings.Join(lines[:len(lines)-1], "\n")
	}

	// If result is not a list, make it a list, so we can add items.
	if !strings.HasPrefix(m, "-") && !strings.HasPrefix(m, "*") {
		m = fmt.Sprintf("- %s", m)
	}

	return m, nil
}

// setPanelSize is a helper to set a panel's size.
func setPanelSize(p *sdk.Panel, width, height int) {
	p.GridPos.W = &width
	p.GridPos.H = &height
}

// setPanelSize is a helper to set a panel's position.
func setPanelPos(p *sdk.Panel, x, y int) {
	p.GridPos.X = &x
	p.GridPos.Y = &y
}
