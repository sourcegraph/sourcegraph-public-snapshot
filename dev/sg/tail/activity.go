package tail

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/grafana/regexp"
)

type activityMsg struct {
	name  string
	ts    string
	level string
	data  string
}

func (a *activityMsg) render(width int, search string) string {
	name := lipgloss.NewStyle().Width(20).Align(lipgloss.Right).Foreground(nameToColor(a.name)).Render(a.name)
	level := lipgloss.NewStyle().Width(6).Align(lipgloss.Center).Background(levelToColor(a.level)).Foreground(lipgloss.Color("0")).Render(a.level)
	wrapped := lipgloss.NewStyle().Width(width - 20 - 6).Render(a.data)
	if search != "" && strings.Contains(wrapped, search) {
		wrapped = lipgloss.NewStyle().Background(lipgloss.Color("3")).Render(wrapped)
	}
	return name + " " + level + " " + wrapped
}

var activityRe = regexp.MustCompile(`^(?P<name>[\w-]+):\s+(?P<ts>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)?\s*(?P<level>\w{4})\s+(?P<data>.*)`)
var tsRe = regexp.MustCompile(`(?:\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)|(?:[\d\/]+ [\d:]+)`)
var levelAndContentRe = regexp.MustCompile(`\s*(\w{4} )?\s*(.*)`) // space after \w{4} is here to disambiguate.

func parseActivity(s string) activityMsg {
	var name, ts, level, data string
	parts := strings.SplitAfterN(s, ":", 2)
	name = strings.TrimSuffix(parts[0], ":")
	rest := strings.TrimSpace(parts[1])

	for _, c := range rest {
		if unicode.IsSpace(c) {
			continue
		}
		if unicode.IsDigit(c) {
			// Ignore the TS for now
			rest = tsRe.ReplaceAllString(rest, "")
		}
		break
	}
	matches := levelAndContentRe.FindStringSubmatch(rest)
	if len(matches) == 2 {
		// We got the content, but not the level
		data = matches[1]
	} else if len(matches) == 3 {
		level = matches[1]
		data = matches[2]
	} else {
		data = rest
	}

	return activityMsg{
		name:  name,
		level: strings.ToUpper(strings.TrimSpace(level)),
		ts:    ts,
		data:  data,
	}
}

type activityPred func(a *activityMsg) *activityMsg

type tab struct {
	title string
	preds activityPreds
}

type activityPreds []activityPred

func (p activityPreds) Apply(a *activityMsg) *activityMsg {
	if p == nil {
		return a
	}
	for _, pred := range p {
		a = pred(a)
		if a == nil {
			return nil
		}
	}
	return a
}
