package monitoring

import (
	"strings"
	"unicode"

	"github.com/grafana-tools/sdk"
)

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

// observablePanelID generates a panel ID unique per dashboard for an observable at a
// given group and row.
func observablePanelID(groupIndex, rowIndex, observableIndex int) uint {
	// by default, Grafana generates panel IDs starting at 0 for panels not assigned an ID.
	// to avoid conflicts, we start generated panel IDs at large number.
	const baseGeneratedPanelID = 100000
	return uint(baseGeneratedPanelID +
		(groupIndex * 100) +
		(rowIndex * 10) +
		(observableIndex * 1))
}

// isValidGrafanaUID checks if the given string is a valid UID for entry into a Grafana dashboard. This is
// primarily used in the URL, e.g. /-/debug/grafana/d/syntect-server/<UID> and allows us to have
// static URLs we can document like:
//
// 	Go to https://sourcegraph.example.com/-/debug/grafana/d/syntect-server/syntect-server
//
// Instead of having to describe all the steps to navigate there because the UID is random.
func isValidGrafanaUID(s string) bool {
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
