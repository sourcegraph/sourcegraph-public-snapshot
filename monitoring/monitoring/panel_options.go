package monitoring

import (
	"github.com/grafana-tools/sdk"
)

// ObservablePanelOption declares an option for customizing a graph panel.
// `ObservablePanel` is responsible for collecting and applying options.
//
// You can make any customization you want to a graph panel by using `ObservablePanel.With`:
//
//   Panel: monitoring.Panel().With(func(o monitoring.Observable, g *sdk.GraphPanel) {
//     // modify 'g' with desired changes
//   }),
//
// When writing a custom `ObservablePanelOption`, keep in mind that:
//
// - There are only ever two `YAxes`: left at `YAxes[0]` and right at `YAxes[1]`.
// Target customizations at the Y-axis you want to modify, e.g. `YAxes[0].Property = Value`.
//
// - The observable being graphed is configured in `Targets[0]`.
// Customize it by editing it directly, e.g. `Targets[0].Property = Value`.
//
// If an option could be leveraged by multiple observables, a shared panel option can be
// defined in the `monitoring` package.
//
// When creating a shared `ObservablePanelOption`, it should defined as a function on the
// `panelOptionsLibrary` that returns a `ObservablePanelOption`. The function should be
// It can then be used with the `ObservablePanel.With`:
//
//   Panel: monitoring.Panel().With(monitoring.PanelOptions.MyCustomization),
//
// Using a shared prefix helps with discoverability of available options.
type ObservablePanelOption func(Observable, *sdk.GraphPanel)

// PanelOptions exports available shared `ObservablePanelOption` implementations.
//
// See `ObservablePanelOption` for more details.
var PanelOptions panelOptionsLibrary

// panelOptionsLibrary provides `ObservablePanelOption` implementations.
//
// Shared panel options should be declared as functions on this struct - see the
// `ObservablePanelOption` documentation for more details.
type panelOptionsLibrary struct{}

// basicPanel instantiates all properties of a graph that can be adjusted in an
// ObservablePanelOption, and some reasonable defaults aimed at maintaining a uniform
// look and feel.
//
// All ObservablePanelOptions start with this option.
func (panelOptionsLibrary) basicPanel() ObservablePanelOption {
	return func(o Observable, g *sdk.GraphPanel) {
		g.Legend.Show = true
		g.Fill = 1
		g.Lines = true
		g.Linewidth = 1
		g.Pointradius = 2
		g.AliasColors = map[string]string{}
		g.Xaxis = sdk.Axis{
			Show: true,
		}
		g.Targets = []sdk.Target{{
			Expr: o.Query,
		}}
		g.Yaxes = []sdk.Axis{
			{
				Decimals: 0,
				LogBase:  1,
				Show:     true,
			},
			{
				// Most graphs will not need the right Y axis, disable by default.
				Show: false,
			},
		}
	}
}

// OptionOpinionatedDefaults sets some opinionated default properties aimed at
// encouraging good dashboard practices.
//
// It is applied in the default PanelOptions().
func (panelOptionsLibrary) OpinionatedDefaults() ObservablePanelOption {
	return func(o Observable, g *sdk.GraphPanel) {
		// We use "value" as the default legend format and not, say, "{{instance}}" or
		// an empty string (Grafana defaults to all labels in that case) because:
		//
		// 1. Using "{{instance}}" is often wrong, see: https://about.sourcegraph.com/handbook/engineering/observability/monitoring_pillars#faq-why-can-t-i-create-a-graph-panel-with-more-than-5-cardinality-labels
		// 2. More often than not, you actually do want to aggregate your whole query with `sum()`, `max()` or similar.
		// 3. If "{{instance}}" or similar was the default, it would be easy for people to say "I guess that's intentional"
		//    instead of seeing multiple "value" labels on their dashboard (which immediately makes them think
		//    "how can I fix that?".)
		g.Targets[0].LegendFormat = "value"
		// Most metrics will have a minimum value of 0.
		g.Yaxes[0].Min = sdk.NewFloatString(0.0)
		// Default to treating values as simple numbers.
		g.Yaxes[0].Format = string(Number)
		// Default to showing a zero when values are null. Using 'connected' can be misleading,
		// and this looks better and less worrisome than just 'null'.
		g.NullPointMode = "null as zero"
	}
}

// AlertThresholds draws threshold lines based on the Observable's configured alerts.
//
// It is applied in the default PanelOptions().
func (panelOptionsLibrary) AlertThresholds() ObservablePanelOption {
	return func(o Observable, g *sdk.GraphPanel) {
		if o.Warning != nil && o.Warning.greaterThan {
			// Warning threshold
			g.Thresholds = append(g.Thresholds, sdk.Threshold{
				Value:     float32(o.Warning.threshold),
				Op:        "gt",
				ColorMode: "custom",
				Line:      true,
				LineColor: "rgba(255, 73, 53, 0.8)",
			})
		}
		if o.Critical != nil && o.Critical.greaterThan {
			// Critical threshold
			g.Thresholds = append(g.Thresholds, sdk.Threshold{
				Value:     float32(o.Critical.threshold),
				Op:        "gt",
				ColorMode: "custom",
				Line:      true,
				LineColor: "rgba(255, 17, 36, 0.8)",
			})
		}
		if o.Warning != nil && o.Warning.lessThan {
			// Warning threshold
			g.Thresholds = append(g.Thresholds, sdk.Threshold{
				Value:     float32(o.Warning.threshold),
				Op:        "lt",
				ColorMode: "custom",
				Line:      true,
				LineColor: "rgba(255, 73, 53, 0.8)",
			})
		}
		if o.Critical != nil && o.Critical.lessThan {
			// Critical threshold
			g.Thresholds = append(g.Thresholds, sdk.Threshold{
				Value:     float32(o.Critical.threshold),
				Op:        "lt",
				ColorMode: "custom",
				Line:      true,
				LineColor: "rgba(255, 17, 36, 0.8)",
			})
		}
	}
}
