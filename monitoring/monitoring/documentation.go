package monitoring

import (
	"bytes"
	"fmt"
	"strings"
)

func renderDocumentation(containers []*Container) ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintf(&b, `# Alert solutions

This document contains possible solutions for when you find alerts are firing in Sourcegraph's monitoring.
If your alert isn't mentioned here, or if the solution doesn't help, [contact us](mailto:support@sourcegraph.com)
for assistance.

To learn more about Sourcegraph's alerting, see [our alerting documentation](https://docs.sourcegraph.com/admin/observability/alerting).

<!-- DO NOT EDIT: generated via: go generate ./monitoring -->

`)
	for _, c := range containers {
		for _, g := range c.Groups {
			for _, r := range g.Rows {
				for _, o := range r {
					if o.Warning == nil && o.Critical == nil {
						continue
					}

					fmt.Fprintf(&b, "## %s: %s\n\n", c.Name, o.Name)
					fmt.Fprintf(&b, `<p class="subtitle">%s: %s</p>`, o.Owner, o.Description)

					// Render descriptions of various levels of this alert
					fmt.Fprintf(&b, "\n\n**Descriptions:**\n\n")
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
							return nil, err
						}
						fmt.Fprintf(&b, "- _%s_\n", desc)
						prometheusAlertNames = append(prometheusAlertNames,
							fmt.Sprintf("  \"%s\"", prometheusAlertName(alert.level, c.Name, o.Name)))
					}
					fmt.Fprint(&b, "\n")

					// Render solutions for dealing with this alert
					fmt.Fprintf(&b, "**Possible solutions:**\n\n")
					if o.PossibleSolutions != "none" {
						possibleSolutions, _ := toMarkdownList(o.PossibleSolutions)
						fmt.Fprintf(&b, "%s\n", possibleSolutions)
					}
					// add silencing configuration as another solution
					fmt.Fprintf(&b, "- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:\n\n")
					fmt.Fprintf(&b, "```json\n%s\n```\n\n", fmt.Sprintf(`"observability.silenceAlerts": [
%s
]`, strings.Join(prometheusAlertNames, ",\n")))

					// Render break for readability
					fmt.Fprint(&b, "<br />\n\n")
				}
			}
		}
	}
	return b.Bytes(), nil
}
