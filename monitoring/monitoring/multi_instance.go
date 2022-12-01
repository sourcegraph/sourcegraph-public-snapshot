package monitoring

import (
	"fmt"
	"sync"

	"github.com/grafana-tools/sdk"
	grafanasdk "github.com/grafana-tools/sdk"
	"github.com/prometheus/prometheus/model/labels"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func renderMultiInstanceDashboard(dashboards []*Dashboard, groupings []string) (*grafanasdk.Board, error) {
	board := sdk.NewBoard("Multi-instance overviews")
	board.AddTags("multi-instance", "generated")
	board.UID = "multi-instance-overviews"
	board.ID = 0
	board.Timezone = "utc"
	board.Timepicker.RefreshIntervals = []string{"5s", "10s", "30s", "1m", "5m", "15m", "30m", "1h", "2h", "1d"}
	board.Time.From = "now-6h"
	board.Time.To = "now"
	board.SharedCrosshair = true
	board.Editable = false

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

	for _, d := range dashboards {
		var row *sdk.Row
		var addDashboardRow sync.Once
		for _, g := range d.Groups {
			for _, r := range g.Rows {
				for _, o := range r {
					if !o.MultiInstance {
						continue
					}

					// Only add row if this dashboard has a multi instance panel, and only
					// do it once per dashboard
					addDashboardRow.Do(func() {
						row = board.AddRow(d.Title)
						row.ShowTitle = true
						row.Collapse = true // avoid crazy loading times
					})

					// TODO make this size correctly in this context and output a valid
					// dashboard, right now it isn't quite right
					panel, err := o.renderPanel(d, panelManipulationOptions{
						injectGroupings:     groupings,
						injectLabelMatchers: variableMatchers,
					}, nil)
					if err != nil {
						return nil, errors.Wrapf(err, "render panel for %q", o.Name)
					}

					row.Add(panel)
				}
			}
		}
	}
	return board, nil
}
