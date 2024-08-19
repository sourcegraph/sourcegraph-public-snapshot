package monitoring

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grafana-tools/sdk"
	"github.com/prometheus/prometheus/model/labels"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring/internal/grafana"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring/internal/promql"
)

// Dashboard usually describes a Service,
// and a service may contain one or more containers to be observed.
//
// It may also be used to describe a collection of services that are highly correlated, and
// it is useful to present them in a single dashboard.
//
// It may also (rarely) be used to describe aggregated infrastructure-wide metrics
// to provide operator an unified view of the system health for easier troubleshooting.
//
// These correspond to dashboards in Grafana.
type Dashboard struct {
	// Name of the Docker container, e.g. "syntect-server".
	Name string

	// Title of the Docker container, e.g. "Syntect Server".
	Title string

	// Description of the Docker container. It should describe what the container
	// is responsible for, so that the impact of issues in it is clear.
	Description string

	// Variables define the variables that can be to applied to the dashboard for this
	// container, such as instances or shards.
	Variables []ContainerVariable

	// Groups of observable information about the container.
	Groups []Group

	// NoSourcegraphDebugServer indicates if this container does not export the standard
	// Sourcegraph debug server (package `internal/debugserver`).
	//
	// This is used to configure monitoring features that depend on information exported
	// by the standard Sourcegraph debug server.
	NoSourcegraphDebugServer bool
}

func (c *Dashboard) validate() error {
	if err := grafana.ValidateUID(c.Name); err != nil {
		return errors.Wrapf(err, "Name %q is invalid", c.Name)
	}

	if c.Title != Title(c.Title) {
		return errors.Errorf("Title must be in Title Case; found \"%s\" want \"%s\"", c.Title, Title(c.Title))
	}
	if c.Description != withPeriod(c.Description) || c.Description != upperFirst(c.Description) {
		return errors.Errorf("Description must be sentence starting with an uppercase letter and ending with period; found \"%s\"", c.Description)
	}

	var errs error
	for i, v := range c.Variables {
		if err := v.validate(); err != nil {
			errs = errors.Append(errs, errors.Errorf("Variable %d %q: %v", i, c.Name, err))
		}
	}
	for i, g := range c.Groups {
		if err := g.validate(c.Variables); err != nil {
			errs = errors.Append(errs, errors.Errorf("Group %d %q: %v", i, g.Title, err))
		}
	}
	return errs
}

// noAlertsDefined indicates if a dashboard no alerts defined.
func (c *Dashboard) noAlertsDefined() bool {
	for _, g := range c.Groups {
		for _, r := range g.Rows {
			for _, o := range r {
				if !o.NoAlert {
					return false
				}
			}
		}
	}
	return true
}

// renderDashboard generates the Grafana renderDashboard for this container.
// UIDs are globally unique identifiers for a given dashboard on Grafana. For normal Sourcegraph usage,
// there is only ever a single dashboard with a given name but Cloud usage requires multiple copies
// of the same dashboards to exist for different folders so we allow the ability to inject custom
// label matchers and folder names to uniquely id the dashboards. UIDs need to be deterministic however
// to generate appropriate alerts and documentations
func (c *Dashboard) renderDashboard(injectLabelMatchers []*labels.Matcher, folder string) (*sdk.Board, error) {
	// If the folder is not specified simply use the name for the UID
	uid := c.Name
	if folder != "" {
		uid = fmt.Sprintf("%s-%s", folder, uid)
		if err := grafana.ValidateUID(uid); err != nil {
			return nil, errors.Wrapf(err, "generated UID %q is invalid", uid)
		}
	}
	board := grafana.NewBoard(uid, c.Title, []string{"builtin"})

	if !c.noAlertsDefined() {
		alertLevelVariable := ContainerVariable{
			Label: "Alert level",
			Name:  "alert_level",
			Options: ContainerVariableOptions{
				Options: []string{"critical", "warning"},
			},
		}
		templateVar, err := alertLevelVariable.toGrafanaTemplateVar(injectLabelMatchers)
		if err != nil {
			return nil, errors.Wrap(err, "Alert level")
		}
		board.Templating.List = []sdk.TemplateVar{templateVar}
	}
	for _, variable := range c.Variables {
		templateVar, err := variable.toGrafanaTemplateVar(injectLabelMatchers)
		if err != nil {
			return nil, errors.Wrap(err, variable.Name)
		}
		board.Templating.List = append(board.Templating.List, templateVar)
	}
	if !c.noAlertsDefined() {
		// Show alerts matching the selected alert_level (see template variable above)
		expr, err := promql.InjectMatchers(
			fmt.Sprintf(`ALERTS{service_name=%q,level=~"$alert_level",alertstate="firing"}`, c.Name),
			injectLabelMatchers, newVariableApplier(c.Variables))
		if err != nil {
			return nil, errors.Wrap(err, "alerts overlay query")
		}

		board.Annotations.List = []sdk.Annotation{{
			Name:        "Alert events",
			Datasource:  pointers.Ptr("Prometheus"),
			Expr:        expr,
			Step:        "60s",
			TitleFormat: "{{ description }} ({{ name }})",
			TagKeys:     "level,owner",
			IconColor:   "rgba(255, 96, 96, 1)",
			Enable:      false, // disable by default for now
			Type:        "tags",
		}}
	}
	// Annotation layers that require a service to export information required by the
	// Sourcegraph debug server - see the `NoSourcegraphDebugServer` docstring.
	if !c.NoSourcegraphDebugServer {
		// Per version, instance generate an annotation whenever labels change
		// inspired by https://github.com/grafana/grafana/issues/11948#issuecomment-403841249
		// We use `job=~.*SERVICE` because of frontend being called sourcegraph-frontend
		// in certain environments
		expr, err := promql.InjectMatchers(
			fmt.Sprintf(`group by(version, instance) (src_service_metadata{job=~".*%[1]s"} unless (src_service_metadata{job=~".*%[1]s"} offset 1m))`, c.Name),
			injectLabelMatchers,
			newVariableApplier(c.Variables))
		if err != nil {
			return nil, errors.Wrap(err, "debug server version expression")
		}

		board.Annotations.List = append(board.Annotations.List, sdk.Annotation{
			Name:        "Version changes",
			Datasource:  pointers.Ptr("Prometheus"),
			Expr:        expr,
			Step:        "60s",
			TitleFormat: "v{{ version }}",
			TagKeys:     "instance",
			IconColor:   "rgb(255, 255, 255)",
			Enable:      false, // disable by default for now
			Type:        "tags",
		})
	}

	description := sdk.NewText("")
	description.Title = "" // Removes vertical space the title would otherwise take up
	setPanelSize(description, 24, 3)
	description.TextPanel.Mode = "html"
	description.TextPanel.Content = fmt.Sprintf(`
	<div style="text-align: left;">
	  <img src="https://sourcegraphstatic.com/sourcegraph-logo-light.png" style="height:30px; margin:0.5rem"></img>
	  <div style="margin-left: 1rem; margin-top: 0.5rem; font-size: 20px;"><b>%s:</b> %s <a style="font-size: 15px" target="_blank" href="https://docs-legacy.sourcegraph.com/dev/background-information/architecture">(⧉ architecture diagram)</a></span>
	</div>
	`, c.Name, c.Description)
	board.Panels = append(board.Panels, description)

	if !c.noAlertsDefined() {
		expr, err := promql.InjectMatchers(fmt.Sprintf(`label_replace(
			sum(
				max by (level,service_name,name,description,grafana_panel_id)(alert_count{service_name="%s",name!="",level=~"$alert_level"})
			) by (
				level,description,service_name,grafana_panel_id,
			),
			"description", "$1",
			"description", ".*: (.*)"
		)`, c.Name), injectLabelMatchers, newVariableApplier(c.Variables))
		if err != nil {
			return nil, errors.Wrap(err, "alerts overview expression")
		}

		alertsDefined := grafana.NewContainerAlertsDefinedTable(sdk.Target{
			Expr:    expr,
			Format:  "table",
			Instant: true,
		})
		setPanelSize(alertsDefined, 9, 5)
		setPanelPos(alertsDefined, 0, 3)
		board.Panels = append(board.Panels, alertsDefined)

		alertsFiring := sdk.NewGraph("Alerts firing")
		setPanelSize(alertsFiring, 15, 5)
		setPanelPos(alertsFiring, 9, 3)
		alertsFiring.GraphPanel.Legend.Show = true
		alertsFiring.GraphPanel.Fill = 1
		alertsFiring.GraphPanel.Bars = true
		alertsFiring.GraphPanel.NullPointMode = "null"
		alertsFiring.GraphPanel.Pointradius = 2
		alertsFiring.GraphPanel.AliasColors = map[string]string{}
		alertsFiring.GraphPanel.Xaxis = sdk.Axis{
			Show: true,
		}
		alertsFiring.GraphPanel.Yaxes = []sdk.Axis{
			{
				Decimals: 0,
				Format:   "short",
				LogBase:  1,
				Max:      sdk.NewFloatString(1),
				Min:      sdk.NewFloatString(0),
				Show:     false,
			},
			{
				Format:  "short",
				LogBase: 1,
				Show:    true,
			},
		}
		alertsFiringExpr, err := promql.InjectMatchers(
			fmt.Sprintf(`sum by (service_name,level,name,grafana_panel_id)(max by (level,service_name,name,description,grafana_panel_id)(alert_count{service_name="%s",name!="",level=~"$alert_level"}) >= 1)`, c.Name),
			injectLabelMatchers,
			newVariableApplier(c.Variables),
		)
		if err != nil {
			return nil, errors.Wrap(err, "Alerts firing")
		}
		alertsFiring.AddTarget(&sdk.Target{
			Expr:         alertsFiringExpr,
			LegendFormat: "{{level}}: {{name}}",
		})
		alertsFiring.GraphPanel.FieldConfig = &sdk.FieldConfig{}
		alertsFiring.GraphPanel.FieldConfig.Defaults.Links = []sdk.Link{{
			Title: "Graph panel",
			URL:   pointers.Ptr("/-/debug/grafana/d/${__field.labels.service_name}/${__field.labels.service_name}?viewPanel=${__field.labels.grafana_panel_id}"),
		}}
		board.Panels = append(board.Panels, alertsFiring)
	}

	baseY := 8
	offsetY := baseY
	for groupIndex, group := range c.Groups {
		// Non-general groups are shown as collapsible panels.
		var rowPanel *sdk.Panel
		if group.Title != "General" {
			offsetY++
			rowPanel = grafana.NewRowPanel(offsetY, group.Title)
			rowPanel.Collapsed = group.Hidden
			board.Panels = append(board.Panels, rowPanel)
		}

		// Generate a panel for displaying each observable in each row.
		for rowIndex, row := range group.Rows {
			panelWidth := 24 / len(row)
			offsetY++
			for panelIndex, o := range row {
				panel, err := o.renderPanel(c, panelManipulationOptions{
					injectLabelMatchers: injectLabelMatchers,
				}, &panelRenderOptions{
					groupIndex:  groupIndex,
					rowIndex:    rowIndex,
					panelIndex:  panelIndex,
					panelWidth:  panelWidth,
					panelHeight: 5,
					offsetY:     offsetY,
				})
				if err != nil {
					return nil, errors.Wrapf(err, "render panel for %q", o.Name)
				}

				// Attach panel to board
				if rowPanel != nil && group.Hidden {
					rowPanel.RowPanel.Panels = append(rowPanel.RowPanel.Panels, *panel)
				} else {
					board.Panels = append(board.Panels, panel)
				}
			}
		}
	}
	return board, nil
}

// alertDescription generates an alert description for the specified coontainer's alert.
func (c *Dashboard) alertDescription(o Observable, alert *ObservableAlertDefinition) (string, error) {
	if alert.isEmpty() {
		return "", errors.New("cannot generate description for empty alert")
	}

	var description string

	// description based on thresholds. no special description for 'alert.strictCompare',
	// because the description is pretty ambiguous to fit different alerts.
	units := o.Panel.unitType.short()
	if alert.description != "" {
		description = fmt.Sprintf("%s: %s", c.Name, alert.description)
	} else if alert.greaterThan {
		// e.g. "zoekt-indexserver: 20+ indexed search request errors every 5m by code"
		description = fmt.Sprintf("%s: %v%s+ %s", c.Name, alert.threshold, units, o.Description)
	} else if alert.lessThan {
		// e.g. "zoekt-indexserver: less than 20 indexed search requests every 5m by code"
		description = fmt.Sprintf("%s: less than %v%s %s", c.Name, alert.threshold, units, o.Description)
	} else {
		return "", errors.Errorf("unable to generate description for observable %+v", o)
	}

	// add information about "for"
	if alert.duration > 0 {
		return fmt.Sprintf("%s for %s", description, alert.duration), nil
	}
	return description, nil
}

// RenderPrometheusRules generates the Prometheus rules file which defines our
// high-level alerting metrics for the container. For more information about
// how these work, see:
//
// https://sourcegraph.com/docs/admin/observability/metrics#high-level-alerting-metrics
func (c *Dashboard) RenderPrometheusRules(injectLabelMatchers []*labels.Matcher) (*PrometheusRules, error) {
	group := newPrometheusRuleGroup(c.Name)
	for groupIndex, g := range c.Groups {
		for rowIndex, r := range g.Rows {
			for observableIndex, o := range r {
				for level, a := range map[string]*ObservableAlertDefinition{
					"warning":  o.Warning,
					"critical": o.Critical,
				} {
					if a.isEmpty() {
						continue
					}

					alertQuery, err := a.generateAlertQuery(o, injectLabelMatchers,
						// Alert queries cannot use variable intervals
						newVariableApplierWith(c.Variables, false))
					if err != nil {
						return nil, errors.Errorf("%s.%s.%s: unable to generate query: %+v",
							c.Name, o.Name, level, err)
					}

					// Build the rule with appropriate labels. Labels are leveraged in various integrations, such as with prom-wrapper.
					description, err := c.alertDescription(o, a)
					if err != nil {
						return nil, errors.Errorf("%s.%s.%s: unable to generate labels: %+v",
							c.Name, o.Name, level, err)
					}

					labelMap := map[string]string{
						"name":         o.Name,
						"level":        level,
						"service_name": c.Name,
						"description":  description,
						"owner":        o.Owner.opsgenieTeam,

						// in the corresponding dashboard, this label should indicate
						// the panel associated with this rule
						"grafana_panel_id": strconv.Itoa(int(observablePanelID(groupIndex, rowIndex, observableIndex))),
					}
					// Inject labels as fixed values for alert rules
					for _, l := range injectLabelMatchers {
						labelMap[l.Name] = l.Value
					}
					group.appendRow(alertQuery, labelMap, a.duration)
				}
			}
		}
	}
	if err := group.validate(); err != nil {
		return nil, err
	}
	return &PrometheusRules{
		Groups: []PrometheusRuleGroup{group},
	}, nil
}

// Group describes a group of observable information about a container.
//
// These correspond to collapsible sections in a Grafana dashboard.
type Group struct {
	// Title of the group, briefly summarizing what this group is about, or
	// "General" if the group is just about the container in general.
	Title string

	// Hidden indicates whether or not the group should be hidden by default.
	//
	// This should only be used when the dashboard is already full of information
	// and the information presented in this group is unlikely to be the cause of
	// issues and should generally only be inspected in the event that an alert
	// for that information is firing.
	Hidden bool

	// Rows of observable metrics.
	Rows []Row
}

func (g Group) validate(variables []ContainerVariable) error {
	if g.Title != upperFirst(g.Title) || g.Title == withPeriod(g.Title) {
		return errors.Errorf("Title must start with an uppercase letter and not end with a period; found \"%s\"", g.Title)
	}
	var errs error
	for i, r := range g.Rows {
		if err := r.validate(variables); err != nil {
			errs = errors.Append(errs, errors.Errorf("Row %d: %v", i, err))
		}
	}
	return errs
}

// Row of observable metrics.
//
// These correspond to a row of Grafana graphs.
type Row []Observable

func (r Row) validate(variables []ContainerVariable) error {
	if len(r) < 1 || len(r) > 4 {
		return errors.Errorf("row must have 1 to 4 observables only, found %v", len(r))
	}

	var errs error
	for i, o := range r {
		if err := o.validate(variables); err != nil {
			errs = errors.Append(errs, errors.Errorf("Observable %d %q: %v", i, o.Name, err))
		}
	}
	return errs
}

// Observable describes a metric about a container that can be observed. For example, memory usage.
//
// These correspond to Grafana graphs.
type Observable struct {
	// Name is a short and human-readable lower_snake_case name describing what is being observed.
	//
	// It must be unique relative to the service name.
	//
	// Good examples:
	//
	//  github_rate_limit_remaining
	// 	search_error_rate
	//
	// Bad examples:
	//
	//  repo_updater_github_rate_limit
	// 	search_error_rate_over_5m
	//
	Name string

	// Description is a human-readable description of exactly what is being observed.
	// If a query groups by a label (such as with a `sum by(...)`), ensure that this is
	// reflected in the description by noting that this observable is grouped "by ...".
	//
	// Good examples:
	//
	// 	"remaining GitHub API rate limit quota"
	// 	"number of search errors every 5m"
	//  "90th percentile search request duration over 5m"
	//  "internal API error responses every 5m by route"
	//
	// Bad examples:
	//
	// 	"GitHub rate limit"
	// 	"search errors[5m]"
	// 	"P90 search latency"
	//
	Description string

	// Owner indicates the team that owns this Observable (including its alerts and maintainence).
	Owner ObservableOwner

	// Query is the actual Prometheus query that should be observed.
	Query string

	// DataMustExist indicates if the query must return data.
	//
	// For example, repo_updater_memory_usage should always have data present and an alert should
	// fire if for some reason that query is not returning any data, so this would be set to true.
	// In contrast, search_error_rate would depend on users actually performing searches and we
	// would not want an alert to fire if no data was present, so this will not need to be set.
	DataMustExist bool

	// Warning alerts indicate that something *could* be wrong with Sourcegraph. We
	// suggest checking in on these periodically, or using a notification channel that
	// will not bother anyone if it is spammed.
	//
	// Learn more about how alerting is used: https://sourcegraph.com/docs/admin/observability/alerting
	Warning *ObservableAlertDefinition

	// Critical alerts indicate that something is definitively wrong with Sourcegraph,
	// in a way that is very likely to be noticeable to users. We suggest using a
	// high-visibility notification channel, such as paging, for these alerts.
	//
	// Learn more about how alerting is used: https://sourcegraph.com/docs/admin/observability/alerting
	Critical *ObservableAlertDefinition

	// NoAlerts must be set by Observables that do not have any alerts. This ensures the
	// omission of alerts is intentional. If set to true, an Interpretation must be
	// provided in place of NextSteps.
	//
	// Consider adding at least a Warning or Critical alert to each Observable to make it
	// easy to identify when the target of this metric is misbehaving.
	NoAlert bool

	// NextSteps is Markdown describing possible next steps in the event that the alert is
	// firing. It does not have to indicate a definite solution, just the next steps that
	// Sourcegraph administrators (both within Sourcegraph and at customers) can understand
	// and leverage when get a notification for this alert.
	//
	// NextSteps should include debugging instructions, links to background information,
	// and potential actions to take. Contacting support should NOT be mentioned as part
	// of a possible solution, as it is already communicated elsewhere.
	//
	// This field is not required if no alerts are attached to this Observable. If there
	// is no clear potential resolution "none" must be explicitly stated, though if a
	// Critical alert is defined providing "none" is not allowed.
	//
	// Use the Interpretation field for additional guidance on understanding this Observable
	// that isn't directly related to solving it.
	//
	// To make writing the Markdown more friendly in Go, string literals like this:
	//
	// 	Observable{
	// 		NextSteps: `
	// 			- Foobar 'some code'
	// 		`
	// 	}
	//
	// Becomes:
	//
	// 	- Foobar `some code`
	//
	// In other words:
	//
	// 1. The preceding newline is removed.
	// 2. The indentation in the string literal is removed (based on the last line).
	// 3. Single quotes become backticks.
	// 4. The last line (which is all indention) is removed.
	// 5. Non-list items are converted to a list.
	//
	// The processed contents are rendered in https://docs.sourcegraph.com/admin/observability/alerts
	NextSteps string

	// Interpretation is Markdown that can serve as a reference for interpreting this
	// observable. For example, Interpretation could provide guidance on what sort of
	// patterns to look for in the observable's graph and document why this observable is
	// useful.
	//
	// If no alerts are configured for an observable, this field is required. If the
	// Description is sufficient to capture what this Observable describes, "none" must be
	// explicitly stated.
	//
	// To make writing the Markdown more friendly in Go, string literal processing as
	// NextSteps is provided, though the output is not converted to a list.
	//
	// The processed contents are rendered in https://docs.sourcegraph.com/admin/observability/dashboards
	Interpretation string

	// Panel provides options for how to render the metric in the Grafana panel.
	// A recommended set of options and customizations are available from the `Panel()`
	// constructor.
	//
	// Additional customizations can be made via `ObservablePanel.With()` for cases where
	// the provided `ObservablePanel` is insufficient - see `ObservablePanelOption` for
	// more details.
	Panel ObservablePanel

	// MultiInstance allows a panel to opt-in to a generated multi-instance overview
	// dashboard, which is created for Sourcegraph Cloud's centralized observability.
	MultiInstance bool
}

func (o Observable) validate(variables []ContainerVariable) error {
	if strings.Contains(o.Name, " ") || strings.ToLower(o.Name) != o.Name {
		return errors.Errorf("Name must be in lower_snake_case; found \"%s\"", o.Name)
	}
	if len(o.Description) == 0 {
		return errors.New("Description must be set")
	}
	// Avoid paragraphs in the brief description
	const maxDescriptionLength = 120
	if len(o.Description) > maxDescriptionLength {
		return errors.Newf("Description must be under %d characters (current length: %d) - to provide more context, use Interpretation and NextSteps fields instead",
			maxDescriptionLength, len(o.Description))
	}
	if first, second := string([]rune(o.Description)[0]), string([]rune(o.Description)[1]); first != strings.ToLower(first) && second == strings.ToLower(second) {
		return errors.Errorf("Description must be lowercase except for acronyms; found \"%s\"", o.Description)
	}

	// If there is an alert, the given owner must be valid
	if !o.NoAlert {
		if err := o.Owner.validate(); err != nil {
			return err
		}
	}

	if !o.Panel.panelType.validate() {
		return errors.New(`Panel.panelType must be "graph" or "heatmap"`)
	}

	// Check if query is valid
	allowIntervalVariables := o.NoAlert // if no alert is configured, allow interval variables
	if err := promql.Validate(o.Query, newVariableApplierWith(variables, allowIntervalVariables)); err != nil {
		return errors.Wrapf(err, "Query is invalid")
	}

	// Validate alerting on this observable
	allAlertsEmpty := o.alertsCount() == 0
	if allAlertsEmpty || o.NoAlert {
		// Ensure lack of alerts is intentional
		if allAlertsEmpty && !o.NoAlert {
			return errors.Errorf("Warning or Critical must be set or explicitly disable alerts with NoAlert")
		} else if !allAlertsEmpty && o.NoAlert {
			return errors.Errorf("An alert is set, but NoAlert is also true")
		}
		// NextSteps if there are no alerts is redundant and likely an error
		if o.NextSteps != "" {
			return errors.Errorf(`NextSteps is not required if no alerts are configured - did you mean to provide an Interpretation instead?`)
		}
		// Interpretation must be provided and valid
		if o.Interpretation == "" {
			return errors.Errorf("Interpretation must be provided if no alerts are set")
		} else if o.Interpretation != "none" {
			if _, err := toMarkdown(o.Interpretation, false); err != nil {
				return errors.Errorf("Interpretation cannot be converted to Markdown: %w", err)
			}
		}
	} else {
		// Ensure alerts are valid
		for alertLevel, alert := range map[string]*ObservableAlertDefinition{
			"Warning":  o.Warning,
			"Critical": o.Critical,
		} {
			if err := alert.validate(); err != nil {
				return errors.Errorf("%s Alert: %w", alertLevel, err)
			}
		}

		// NextSteps must be provided and valid
		if o.NextSteps == "" {
			return errors.Errorf(`NextSteps must list steps or an explicit "none"`)
		}

		// If a critical alert is set, NextSteps must be provided. Empty case
		if !o.Critical.isEmpty() && o.NextSteps == "none" {
			return errors.Newf(`NextSteps must be provided if a critical alert is set`)
		}

		// Check if provided NextSteps is valid
		if o.NextSteps != "none" {
			if nextSteps, err := toMarkdown(o.NextSteps, true); err != nil {
				return errors.Errorf("NextSteps cannot be converted to Markdown: %w", err)
			} else if l := strings.ToLower(nextSteps); strings.Contains(l, "contact support") || strings.Contains(l, "contact us") {
				return errors.Errorf("NextSteps should not include mentions of contacting support")
			}
		}
	}

	return nil
}

func (o Observable) alertsCount() (count int) {
	if !o.Warning.isEmpty() {
		count++
	}
	if !o.Critical.isEmpty() {
		count++
	}
	return
}

type panelRenderOptions struct {
	groupIndex  int
	rowIndex    int
	panelIndex  int
	panelWidth  int
	panelHeight int

	offsetY int
}

type panelManipulationOptions struct {
	injectLabelMatchers []*labels.Matcher
	injectGroupings     []string
}

func (o Observable) renderPanel(c *Dashboard, manipulations panelManipulationOptions, opts *panelRenderOptions) (*sdk.Panel, error) {
	panelTitle := strings.ToTitle(string([]rune(o.Description)[0])) + string([]rune(o.Description)[1:])

	var panel *sdk.Panel
	switch o.Panel.panelType {
	case PanelTypeGraph:
		panel = sdk.NewGraph(panelTitle)
	case PanelTypeHeatmap:
		panel = sdk.NewHeatmap(panelTitle)
	}

	// Set attributes based on position, if available
	if opts != nil {
		// Generating a stable ID
		panel.ID = observablePanelID(opts.groupIndex, opts.rowIndex, opts.panelIndex)

		// Set positioning
		setPanelSize(panel, opts.panelWidth, opts.panelHeight)
		setPanelPos(panel, opts.panelIndex*opts.panelWidth, opts.offsetY)
	}

	// Add reference links
	panel.Links = []sdk.Link{{
		Title:       "Panel reference",
		URL:         pointers.Ptr(fmt.Sprintf("%s#%s", canonicalDashboardsDocsURL, observableDocAnchor(c, o))),
		TargetBlank: pointers.Ptr(true),
	}}
	if !o.NoAlert {
		panel.Links = append(panel.Links, sdk.Link{
			Title:       "Alerts reference",
			URL:         pointers.Ptr(fmt.Sprintf("%s#%s", canonicalAlertDocsURL, observableDocAnchor(c, o))),
			TargetBlank: pointers.Ptr(true),
		})
	}

	// Build the graph panel
	o.Panel.build(o, panel)

	// Apply injected label matchers
	for _, target := range *panel.GetTargets() {
		var err error
		target.Expr, err = promql.InjectMatchers(target.Expr, manipulations.injectLabelMatchers, newVariableApplier(c.Variables))
		if err != nil {
			return nil, errors.Wrap(err, target.Query)
		}

		if len(manipulations.injectGroupings) > 0 {
			target.Expr, err = promql.InjectGroupings(target.Expr, manipulations.injectGroupings, newVariableApplier(c.Variables))
			if err != nil {
				return nil, errors.Wrap(err, target.Query)
			}

			for _, g := range manipulations.injectGroupings {
				target.LegendFormat = fmt.Sprintf("%s - {{%s}}", target.LegendFormat, g)
			}
		}

		panel.SetTarget(&target)
	}

	return panel, nil
}

// Alert provides a builder for defining alerting on an Observable.
func Alert() *ObservableAlertDefinition {
	return &ObservableAlertDefinition{}
}

// ObservableAlertDefinition defines when an alert would be considered firing.
type ObservableAlertDefinition struct {
	greaterThan bool
	lessThan    bool
	duration    time.Duration
	// Wrap the query in `max()` or `min()` so that if there are multiple series (e.g. per-container)
	// they get "flattened" into a single metric. The `aggregator` variable sets the required operator.
	//
	// We only support per-service alerts, not per-container/replica, and not doing so can cause issues.
	// See https://github.com/sourcegraph/sourcegraph/issues/11571#issuecomment-654571953,
	// https://github.com/sourcegraph/sourcegraph/issues/17599, and related pull requests.
	aggregator Aggregator
	// Comparator sets how a metric should be compared against a threshold.
	comparator string
	// Threshold sets the value to be compared against.
	threshold float64
	// alternative customQuery to use for an alert instead of the observables customQuery.
	customQuery string
	// alternative description to use for an alert instead of the observables description.
	description string
}

// GreaterOrEqual indicates the alert should fire when greater or equal the given value.
func (a *ObservableAlertDefinition) GreaterOrEqual(f float64) *ObservableAlertDefinition {
	a.greaterThan = true
	a.aggregator = AggregatorMax
	a.comparator = ">="
	a.threshold = f
	return a
}

// LessOrEqual indicates the alert should fire when less than or equal to the given value.
func (a *ObservableAlertDefinition) LessOrEqual(f float64) *ObservableAlertDefinition {
	a.lessThan = true
	a.aggregator = AggregatorMin
	a.comparator = "<="
	a.threshold = f
	return a
}

// Greater indicates the alert should fire when strictly greater to this value.
func (a *ObservableAlertDefinition) Greater(f float64) *ObservableAlertDefinition {
	a.greaterThan = true
	a.aggregator = AggregatorMax
	a.comparator = ">"
	a.threshold = f
	return a
}

// Less indicates the alert should fire when strictly less than this value.
func (a *ObservableAlertDefinition) Less(f float64) *ObservableAlertDefinition {
	a.lessThan = true
	a.aggregator = AggregatorMin
	a.comparator = "<"
	a.threshold = f
	return a
}

// For indicates how long the given thresholds must be exceeded for this alert to be
// considered firing. Defaults to 0s (immediately alerts when threshold is exceeded).
func (a *ObservableAlertDefinition) For(d time.Duration) *ObservableAlertDefinition {
	a.duration = d
	return a
}

// CustomQuery sets a different query to be used for this alert instead of the query used
// in the Grafana panel. Note that thresholds, etc will still be generated for the panel, so
// ensure the panel query still makes sense in the context of an alert with a custom
// query.
func (a *ObservableAlertDefinition) CustomQuery(query string) *ObservableAlertDefinition {
	a.customQuery = query
	return a
}

// CustomDescription sets a different description to be used for this alert instead of the description
// used for the Grafana panel.
func (a *ObservableAlertDefinition) CustomDescription(desc string) *ObservableAlertDefinition {
	a.description = desc
	return a
}

type Aggregator string

const (
	AggregatorSum = "sum"
	AggregatorMax = "max"
	AggregatorMin = "min"
)

// AggregateBy configures the aggregator to use for this alert. Make sure to only call
// this after setting one of GreaterOrEqual, LessOrEqual, etc.
//
// By default, Less* thresholds are configured with AggregatorMin, and
// Greater* thresholds are configured with AggregatorMax.
func (a *ObservableAlertDefinition) AggregateBy(aggregator Aggregator) *ObservableAlertDefinition {
	a.aggregator = aggregator
	return a
}

func (a *ObservableAlertDefinition) isEmpty() bool {
	return a == nil || (*a == ObservableAlertDefinition{}) || (!a.greaterThan && !a.lessThan)
}

func (a *ObservableAlertDefinition) validate() error {
	if a.isEmpty() {
		return nil
	}
	if a.greaterThan && a.lessThan {
		return errors.New("only one bound (greater or less) can be set")
	}

	if a.customQuery != "" {
		// Check if custom query is a valid alert query. Also note that custom queries
		// should not use variables, so we don't provide them here.
		if _, err := promql.InjectAsAlert(a.customQuery, nil, nil); err != nil {
			return errors.Wrapf(err, "CustomQuery is invalid")
		}
	}
	return nil
}

func (a *ObservableAlertDefinition) generateAlertQuery(o Observable, injectLabelMatchers []*labels.Matcher, vars promql.VariableApplier) (string, error) {
	// The alertQuery must contribute a query that returns true when it should be firing.
	var alertQuery string
	if a.customQuery != "" {
		alertQuery = fmt.Sprintf("%s((%s) %s %v)", a.aggregator, a.customQuery, a.comparator, a.threshold)
	} else {
		alertQuery = fmt.Sprintf("%s((%s) %s %v)", a.aggregator, o.Query, a.comparator, a.threshold)
	}

	// If the data must exist, we alert if the query returns no value as well
	if o.DataMustExist {
		alertQuery = fmt.Sprintf("(%s) OR (absent(%s) == 1)", alertQuery, o.Query)
	}

	// Inject label matchers
	return promql.InjectAsAlert(alertQuery, injectLabelMatchers, vars)
}
