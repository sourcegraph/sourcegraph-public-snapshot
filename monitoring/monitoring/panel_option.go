package monitoring

import (
	"github.com/grafana-tools/sdk"
)

// ObservablePanelOption declares an option for customizing a graph panel.
//
// When writing a custom ObservablePanelOption, keep in mind that:
//
// - There are only ever two `YAxes`: left at `YAxes[0]` and right at `YAxes[1]`.
//   Target customizations at the Y-axis you want to modify, e.g. `YAxes[0].Property = Value`.
// - There observable being graphed is configured in `Targets[0]`.
//   Customize it by editing it directly, e.g. `Targets[0].Property = Value`.
//
// ObservablePanelOptions is responsible for collecting and applying options.
type ObservablePanelOption func(Observable, *sdk.GraphPanel)

// basicPanel instantiates all properties of a graph that can be adjusted in an
// ObservablePanelOption, and some reasonable defaults aimed at maintaining a uniform
// look and feel.
//
// All ObservablePanelOptions start with this option.
func basicPanel() ObservablePanelOption {
	return func(o Observable, g *sdk.GraphPanel) {
		g.Legend.Show = true
		g.Fill = 1
		g.Lines = true
		g.Linewidth = 1
		g.NullPointMode = "connected"
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

// PanelWithOpinionatedDefaults sets some opinionated default properties aimed at
// encouraging good dashboard practices.
//
// It is applied in the default PanelOptions().
func PanelWithOpinionatedDefaults() ObservablePanelOption {
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
	}
}

// PanelWithThresholds draws threshold lines based on the Observable's configured alerts.
//
// It is applied in the default PanelOptions().
func PanelWithThresholds() ObservablePanelOption {
	return func(o Observable, g *sdk.GraphPanel) {
		if o.Warning != nil && o.Warning.greaterOrEqual != nil {
			// Warning threshold
			g.Thresholds = append(g.Thresholds, sdk.Threshold{
				Value:     float32(*o.Warning.greaterOrEqual),
				Op:        "gt",
				ColorMode: "custom",
				Line:      true,
				LineColor: "rgba(255, 73, 53, 0.8)",
			})
		}
		if o.Critical != nil && o.Critical.greaterOrEqual != nil {
			// Critical threshold
			g.Thresholds = append(g.Thresholds, sdk.Threshold{
				Value:     float32(*o.Critical.greaterOrEqual),
				Op:        "gt",
				ColorMode: "custom",
				Line:      true,
				LineColor: "rgba(255, 17, 36, 0.8)",
			})
		}
		if o.Warning != nil && o.Warning.lessOrEqual != nil {
			// Warning threshold
			g.Thresholds = append(g.Thresholds, sdk.Threshold{
				Value:     float32(*o.Warning.lessOrEqual),
				Op:        "lt",
				ColorMode: "custom",
				Line:      true,
				LineColor: "rgba(255, 73, 53, 0.8)",
			})
		}
		if o.Critical != nil && o.Critical.lessOrEqual != nil {
			// Critical threshold
			g.Thresholds = append(g.Thresholds, sdk.Threshold{
				Value:     float32(*o.Critical.lessOrEqual),
				Op:        "lt",
				ColorMode: "custom",
				Line:      true,
				LineColor: "rgba(255, 17, 36, 0.8)",
			})
		}
	}
}
