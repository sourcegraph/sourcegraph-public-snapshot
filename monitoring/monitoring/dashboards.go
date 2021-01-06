package monitoring

import (
	"fmt"
	"strings"
	"unicode"

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
	default:
		panic("never here")
	}
}

// From https://sourcegraph.com/github.com/grafana/grafana@b63b82976b3708b082326c0b7d42f38d4bc261fa/-/blob/packages/grafana-data/src/valueFormats/categories.ts#L23
const (
	// Number is the default unit type.
	Number UnitType = "short"

	// Milliseconds for representing time.
	Milliseconds UnitType = "dtdurationms"

	// Seconds for representing time.
	Seconds UnitType = "dtdurations"

	// Percentage in the range of 0-100.
	Percentage UnitType = "percent"

	// Bytes in IEC (1024) format, e.g. for representing storage sizes.
	Bytes UnitType = "bytes"

	// BitsPerSecond, e.g. for representing network and disk IO.
	BitsPerSecond UnitType = "bps"
)

// ObservablePanelOptions declares options for visualizing an Observable.
type ObservablePanelOptions struct {
	options                []ObservablePanelOption
	disableAlertThresholds bool

	// parameters used by other parts of the generator
	unitType UnitType
}

// PanelOptions provides a builder for customizing an Observable visualization, starting
// with recommended defaults.
func PanelOptions() ObservablePanelOptions {
	return ObservablePanelOptions{
		options: []ObservablePanelOption{
			optionMinimal(),
			optionWithThresholds(),
			optionWithUnits(Number),
			// misc defaults
			func(o Observable, g *sdk.GraphPanel) {
				g.Targets[0].LegendFormat = "value"
				g.Yaxes[0].Min = sdk.NewFloatString(0.0)
			},
		},
	}
}

// MinimalPanelOptions provides a builder for customizing an Observable visualization
// starting with an extremely minimal graph panel.
func MinimalPanelOptions() ObservablePanelOptions {
	return ObservablePanelOptions{
		options: []ObservablePanelOption{
			optionMinimal(),
		},
	}
}

// Min sets the minimum value of the Y axis on the panel. The default is zero.
func (p ObservablePanelOptions) Min(min float64) ObservablePanelOptions {
	p.options = append(p.options, func(o Observable, g *sdk.GraphPanel) {
		g.Yaxes[0].Min = sdk.NewFloatString(min)
	})
	return p
}

// Min sets the minimum value of the Y axis on the panel to auto, instead of
// the default zero.
//
// This is generally only useful if trying to show negative numbers.
func (p ObservablePanelOptions) MinAuto() ObservablePanelOptions {
	p.options = append(p.options, func(o Observable, g *sdk.GraphPanel) {
		g.Yaxes[0].Min = nil
	})
	return p
}

// Max sets the maximum value of the Y axis on the panel. The default is auto.
func (p ObservablePanelOptions) Max(max float64) ObservablePanelOptions {
	p.options = append(p.options, func(o Observable, g *sdk.GraphPanel) {
		g.Yaxes[0].Max = sdk.NewFloatString(max)
	})
	return p
}

// LegendFormat sets the panel's legend format, which may use Go template strings to select
// labels from the Prometheus query.
func (p ObservablePanelOptions) LegendFormat(format string) ObservablePanelOptions {
	p.options = append(p.options, func(o Observable, g *sdk.GraphPanel) {
		g.Targets[0].LegendFormat = format
	})
	return p
}

// Unit sets the panel's Y axis unit type.
func (p ObservablePanelOptions) Unit(t UnitType) ObservablePanelOptions {
	p.unitType = t
	p.options = append(p.options, optionWithUnits(t))
	return p
}

// Interval declares the panel's interval in milliseconds.
func (p ObservablePanelOptions) Interval(ms int) ObservablePanelOptions {
	p.options = append(p.options, func(o Observable, g *sdk.GraphPanel) {
		g.Targets[0].Interval = fmt.Sprintf("%dms", ms)
	})
	return p
}

func (p ObservablePanelOptions) Thresholds(enabled bool) {
	p.disableAlertThresholds = !enabled
}

// WithOption renders an additional option.
func (p ObservablePanelOptions) WithOption(op ObservablePanelOption) ObservablePanelOptions {
	p.options = append(p.options, op)
	return p
}

func (p ObservablePanelOptions) generatePipeline(o Observable) []ObservablePanelOption {
	if p.disableAlertThresholds {
		return p.options
	}
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
