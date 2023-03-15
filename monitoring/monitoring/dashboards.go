package monitoring

import (
	"github.com/grafana-tools/sdk"
)

// UnitType for controlling the unit type display on graphs.
type UnitType string

// short returns the short string description of the unit, for qualifying a
// number of this unit type as human-readable.
func (u UnitType) short() string {
	switch u {
	case Number, "":
		return ""
	case Milliseconds:
		return "ms"
	case Seconds:
		return "s"
	case Percentage:
		return "%"
	case Bytes:
		return "B"
	case BitsPerSecond:
		return "bps"
	case ReadsPerSecond:
		return "rps"
	case WritesPerSecond:
		return "wps"
	case RequestsPerSecond:
		return "reqps"
	default:
		panic("never here")
	}
}

// From https://sourcegraph.com/github.com/grafana/grafana@b63b82976b3708b082326c0b7d42f38d4bc261fa/-/blob/packages/grafana-data/src/valueFormats/categories.ts#L23
const (
	// Number is the default unit type.
	Number UnitType = "short"

	// Milliseconds for representing time.
	Milliseconds UnitType = "ms"

	// Seconds for representing time.
	Seconds UnitType = "s"

	// Minutes for representing time.
	Minutes UnitType = "m"

	// Percentage in the range of 0-100.
	Percentage UnitType = "percent"

	// Bytes in IEC (1024) format, e.g. for representing storage sizes.
	Bytes UnitType = "bytes"

	// BitsPerSecond, e.g. for representing network and disk IO.
	BitsPerSecond UnitType = "bps"

	// BytesPerSecond, e.g. for representing network and disk IO.
	BytesPerSecond UnitType = "Bps"

	// ReadsPerSecond, e.g for representing disk IO.
	ReadsPerSecond UnitType = "rps"

	// WritesPerSecond, e.g for representing disk IO.
	WritesPerSecond UnitType = "wps"

	// RequestsPerSecond, e.g. for representing number of HTTP requests per second
	RequestsPerSecond UnitType = "reqps"

	// PacketsPerSecond, e.g. for representing number of network packets per second
	PacketsPerSecond UnitType = "pps"
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
