package tail

import (
	"fmt"
	"hash/fnv"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.HiddenBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.HiddenBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
	activeTabStyle   = lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0")).Padding(0, 1, 0)
	inactiveTabStyle = lipgloss.NewStyle().Background(lipgloss.Color("0")).Foreground(lipgloss.Color("7")).Padding(0, 1, 0)
)

func nameToColor(s string) lipgloss.Color {
	h := fnv.New32()
	h.Write([]byte(s))
	// We don't use 256 colors because some of those are too dark/bright and hard to read
	c := int(h.Sum32()) % 220
	if c == 0 {
		// 0 is black, so it's going to be the same color as the background.
		c = 1
	}
	return lipgloss.Color(fmt.Sprintf("%d", c))
}

func levelToColor(level string) lipgloss.Color {
	switch level {
	case "INFO":
		return lipgloss.Color("7") // silver
	case "WARN":
		return lipgloss.Color("11") // yellow
	case "DBUG":
		return lipgloss.Color("8") // gray
	case "EROR":
		return lipgloss.Color("9") // red
	case "HELP":
		// special case for usage message at the beginning, this isn't a real log level.
		return lipgloss.Color("6") // green
	default:
		return lipgloss.Color("0") // black
	}
}
