package monitoring

import (
	"bytes"
	"fmt"
	"strings"
)

const alertSolutionsHeader = `# Sourcegraph alert solutions

This document contains possible solutions for when you find alerts are firing in Sourcegraph's monitoring.
If your alert isn't mentioned here, or if the solution doesn't help, [contact us](mailto:support@sourcegraph.com) for assistance.

To learn more about Sourcegraph's alerting and how to set up alerts, see [our alerting documentation](https://docs.sourcegraph.com/admin/observability/alerting).

<!-- DO NOT EDIT: generated via: go generate ./monitoring -->

`

const dashboardsHeader = `# Sourcegraph monitoring dashboards

This document contains details on how to intepret panels and metrics in Sourcegraph's monitoring dashboards.

To learn more about Sourcegraph's metrics and how to view these dashboards, see [our metrics documentation](https://docs.sourcegraph.com/admin/observability/metrics).

<!-- DO NOT EDIT: generated via: go generate ./monitoring -->`

type documentation struct {
	alertSolutions bytes.Buffer
	dashboards     bytes.Buffer
}

func renderDocumentation(containers []*Container) (*documentation, error) {
	var docs documentation

	fmt.Fprint(&docs.alertSolutions, alertSolutionsHeader)
	fmt.Fprint(&docs.dashboards, dashboardsHeader)

	for _, c := range containers {
		for _, g := range c.Groups {
			for _, r := range g.Rows {
				for _, o := range r {
					if err := docs.renderAlertSolutionEntry(c, o); err != nil {
						return nil, fmt.Errorf("error rendering alert solution entry %q %q: %w",
							c.Name, o.Name, err)
					}
				}
			}
		}
	}

	return &docs, nil
}

func (d *documentation) renderAlertSolutionEntry(c *Container, o Observable) error {
	if o.Warning == nil && o.Critical == nil {
		return nil
	}

	fmt.Fprintf(&d.alertSolutions, "## %s: %s\n\n", c.Name, o.Name)
	fmt.Fprintf(&d.alertSolutions, `<p class="subtitle">%s: %s</p>`, o.Owner, o.Description)

	// Render descriptions of various levels of this alert
	fmt.Fprintf(&d.alertSolutions, "\n\n**Descriptions:**\n\n")
	var prometheusAlertNames []string
	for _, alert := range []struct {
		level     string
		threshold *ObservableAlertDefinition
	}{
		{level: "warning", threshold: o.Warning},
		{level: "critical", threshold: o.Critical},
	} {
		if alert.threshold.isEmpty() {
			continue
		}
		desc, err := c.alertDescription(o, alert.threshold)
		if err != nil {
			return err
		}
		fmt.Fprintf(&d.alertSolutions, "- _%s_\n", desc)
		prometheusAlertNames = append(prometheusAlertNames,
			fmt.Sprintf("  \"%s\"", prometheusAlertName(alert.level, c.Name, o.Name)))
	}
	fmt.Fprint(&d.alertSolutions, "\n")

	// Render solutions for dealing with this alert
	fmt.Fprintf(&d.alertSolutions, "**Possible solutions:**\n\n")
	if o.PossibleSolutions != "none" {
		possibleSolutions, _ := toMarkdownList(o.PossibleSolutions)
		fmt.Fprintf(&d.alertSolutions, "%s\n", possibleSolutions)
	}
	// add silencing configuration as another solution
	fmt.Fprintf(&d.alertSolutions, "- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:\n\n")
	fmt.Fprintf(&d.alertSolutions, "```json\n%s\n```\n\n", fmt.Sprintf(`"observability.silenceAlerts": [
%s
]`, strings.Join(prometheusAlertNames, ",\n")))

	// Render break for readability
	fmt.Fprint(&d.alertSolutions, "<br />\n\n")
	return nil
}
