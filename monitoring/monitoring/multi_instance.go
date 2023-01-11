package monitoring

import (
	"fmt"
	"sync"

	"github.com/grafana-tools/sdk"
	grafanasdk "github.com/grafana-tools/sdk"
	"github.com/prometheus/prometheus/model/labels"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring/internal/grafana"
)

func renderMultiInstanceDashboard(dashboards []*Dashboard, groupings []string) (*grafanasdk.Board, error) {
	board := grafana.NewBoard("multi-instance-overviews", "Multi-instance overviews",
		[]string{"multi-instance", "generated"})

	var variableMatchers []*labels.Matcher
	for _, g := range groupings {
		containerVar := ContainerVariable{
			Name:  g,
			Label: g,
			OptionsLabelValues: ContainerVariableOptionsLabelValues{
				// For now we don't support any labels that aren't present on this metric.
				Query:     "src_service_metadata",
				LabelName: g,
			},
			WildcardAllValue: true,
			Multi:            true,
		}
		grafanaVar, err := containerVar.toGrafanaTemplateVar(nil)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate template var for grouping %q", g)
		}
		board.Templating.List = append(board.Templating.List, grafanaVar)

		// generate the matcher to inject
		m, err := labels.NewMatcher(labels.MatchRegexp, g, fmt.Sprintf("${%s:regex}", g))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate template var matcher for grouping %q", g)
		}
		variableMatchers = append(variableMatchers, m)
	}

	var offsetY int
	for dashboardIndex, d := range dashboards {
		var row *sdk.Panel
		var addDashboardRow sync.Once
		for groupIndex, g := range d.Groups {
			for _, r := range g.Rows {
				for observableIndex, o := range r {
					if !o.MultiInstance {
						continue
					}

					// Only add row if this dashboard has a multi instance panel, and only
					// do it once per dashboard
					addDashboardRow.Do(func() {
						offsetY++
						row = grafana.NewRowPanel(offsetY, d.Title)
						row.Collapsed = true // avoid crazy loading times
						board.Panels = append(board.Panels, row)
					})

					// Generate the panel with groupings and variables
					offsetY++
					panel, err := o.renderPanel(d, panelManipulationOptions{
						injectGroupings:     groupings,
						injectLabelMatchers: variableMatchers,
					}, &panelRenderOptions{
						// these indexes are only used for identification
						groupIndex: dashboardIndex,
						rowIndex:   groupIndex,
						panelIndex: observableIndex,

						panelWidth:  24,      // max-width
						panelHeight: 10,      // tall dashboards!
						offsetY:     offsetY, // total index added
					})
					if err != nil {
						return nil, errors.Wrapf(err, "render panel for %q", o.Name)
					}

					row.RowPanel.Panels = append(row.RowPanel.Panels, *panel)
				}
			}
		}
	}
	return board, nil
}
