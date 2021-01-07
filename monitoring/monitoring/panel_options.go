package monitoring

import (
	"fmt"

	"github.com/grafana-tools/sdk"
)

// ObservablePanelOptions declares options for visualizing an Observable. A default set
// of options can be instantiated with `PanelOptions()`, and further customized using
// `ObservablePanelOptions.With(ObservablePanelOption)`.
type ObservablePanelOptions struct {
	options []ObservablePanelOption

	// unitType is used by other parts of the generator
	unitType UnitType
}

// PanelOptions provides a builder for customizing an Observable visualization, starting
// with recommended defaults.
func PanelOptions() ObservablePanelOptions {
	return ObservablePanelOptions{
		options: []ObservablePanelOption{
			optionBasicPanel(), // required basic values
			OptionOpinionatedDefaults(),
			OptionAlertThresholds(),
		},
	}
}

// PanelOptionsMinimal provides a builder for customizing an Observable visualization
// starting with an extremely minimal graph panel.
//
// In general, we advise using PanelOptions() instead to start with recommended defaults.
func PanelOptionsMinimal() ObservablePanelOptions {
	return ObservablePanelOptions{
		options: []ObservablePanelOption{
			optionBasicPanel(), // required basic values
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
	p.options = append(p.options, func(o Observable, g *sdk.GraphPanel) {
		g.Yaxes[0].Format = string(t)
	})
	return p
}

// Interval declares the panel's interval in milliseconds.
func (p ObservablePanelOptions) Interval(ms int) ObservablePanelOptions {
	p.options = append(p.options, func(o Observable, g *sdk.GraphPanel) {
		g.Targets[0].Interval = fmt.Sprintf("%dms", ms)
	})
	return p
}

// With will add the provided option to be applied when building this panel.
func (p ObservablePanelOptions) With(op ObservablePanelOption) ObservablePanelOptions {
	p.options = append(p.options, op)
	return p
}
