package monitoring

import "github.com/grafana-tools/sdk"

// ObservablePanelOption declares an option for customizing a graph panel.
//
// When writing a custom ObservablePanelOption, keep in mind that:
//
// - There are only ever two `YAxes`: left at `YAxes[0]` and right at `YAxes[1]`.
//   Target customizations at the Y-axis you want to modify, e.g. `YAxes[0].Property = Value`.
// - There observable being graphed is configured in `Targets[0]`.
//   Customize it by editing it directly, e.g. `Targets[0].Property = Value`.
//
// `ObservablePanelOptions` is responsible for collecting and applying options.
type ObservablePanelOption func(Observable, *sdk.GraphPanel)

func optionMinimal() ObservablePanelOption {
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
				Format:  "short",
				LogBase: 1,
				Show:    true,
			},
		}
	}
}

func optionWithUnits(units UnitType) ObservablePanelOption {
	return func(o Observable, g *sdk.GraphPanel) {
		g.Yaxes[0].Format = string(units)
	}
}

func optionWithThresholds() ObservablePanelOption {
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
