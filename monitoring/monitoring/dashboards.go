package monitoring

import (
	"strings"
	"unicode"

	"github.com/grafana-tools/sdk"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring/internal/promql"
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

func preprocessDashboards(dashboards []*Dashboard, opts GenerateOptions) error {
	var preprocessErrs error
	if len(opts.InjectLabelMatchers) > 0 {
		for _, dashboard := range dashboards {
			for vi, v := range dashboard.Variables {
				var err error
				if v.OptionsLabel.Query != "" {
					dashboard.Variables[vi].OptionsLabel.Query, err = promql.Inject(v.OptionsLabel.Query, opts.InjectLabelMatchers, nil)
					if err != nil {
						preprocessErrs = errors.Append(preprocessErrs,
							errors.Wrapf(err, "Dashboard %q variable %q", dashboard.Name, v.Name))
						continue
					}
				}
			}

			variables := newVariableApplier(dashboard.Variables)
			for gi, g := range dashboard.Groups {
				for ri, r := range g.Rows {
					for oi, o := range r {
						var err error
						// Update observable query - we need to set the value directly in
						// the slice to update it.
						dashboard.Groups[gi].Rows[ri][oi].Query, err = promql.Inject(o.Query, opts.InjectLabelMatchers, variables)
						if err != nil {
							preprocessErrs = errors.Append(preprocessErrs, errors.Wrapf(err, "Dashboard %q observable %q", dashboard.Name, o.Name))
						}
						// Update custom alert queries if any
						for _, a := range []*ObservableAlertDefinition{
							o.Warning,
							o.Critical,
						} {
							if a != nil && a.customQuery != "" {
								a.customQuery, err = promql.Inject(a.customQuery, opts.InjectLabelMatchers, variables)
								if err != nil {
									preprocessErrs = errors.Append(preprocessErrs,
										errors.Wrapf(err, "Dashboard %q observable %q alert", dashboard.Name, o.Name))

									continue
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}
