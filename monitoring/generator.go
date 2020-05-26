//go:generate echo "Regenerating monitoring..."
//go:generate go build -o /tmp/monitoring-generator
//go:generate /tmp/monitoring-generator

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/grafana-tools/sdk"
	"gopkg.in/yaml.v2"
)

// Container describes a Docker container to be observed.
type Container struct {
	// Name of the Docker container, e.g. "syntect-server".
	Name string

	// Title of the Docker container, e.g. "Syntect Server".
	Title string

	// Description of the Docker container. It should describe what the container
	// is responsible for, so that the impact of issues in it is clear.
	Description string

	// Groups of observable information about the container.
	Groups []Group
}

func (c *Container) validate() error {
	if !isValidUID(c.Name) {
		return fmt.Errorf("Container.Name must be lowercase alphanumeric + dashes; found \"%s\"", c.Name)
	}
	if c.Title != strings.Title(c.Title) {
		return fmt.Errorf("Container.Title must be in Title Case; found \"%s\" want \"%s\"", c.Title, strings.Title(c.Title))
	}
	if c.Description != withPeriod(c.Description) || c.Description != upperFirst(c.Description) {
		return fmt.Errorf("Container.Description must be sentence starting with an uppercas eletter and ending with period; found \"%s\"", c.Description)
	}
	for _, g := range c.Groups {
		if err := g.validate(); err != nil {
			return fmt.Errorf("group %q: %v", g.Title, err)
		}
	}
	return nil
}

// Group describes a group of observable information about a container.
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

func (g Group) validate() error {
	if g.Title != upperFirst(g.Title) || g.Title == withPeriod(g.Title) {
		return fmt.Errorf("Group.Title must start with an uppercase letter and not end with a period; found \"%s\"", g.Title)
	}
	for i, r := range g.Rows {
		if err := r.validate(); err != nil {
			return fmt.Errorf("row %d: %v", i, err)
		}
	}
	return nil
}

// Row of observable metrics.
type Row []Observable

func (r Row) validate() error {
	if len(r) < 1 || len(r) > 4 {
		return fmt.Errorf("row must have 1 to 4 observables only, found %v", len(r))
	}
	for _, o := range r {
		if err := o.validate(); err != nil {
			return fmt.Errorf("observable %q: %v", o.Name, err)
		}
	}
	return nil
}

// Observable describes a metric about a container that can be observed. For example, memory usage.
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
	//
	// Good examples:
	//
	// 	"remaining GitHub API rate limit quota"
	// 	"number of search errors every 5m"
	//  "90th percentile search request duration over 5m"
	//
	// Bad examples:
	//
	// 	"GitHub rate limit"
	// 	"search errors[5m]"
	// 	"P90 search latency"
	//
	Description string

	// Query is the actual Prometheus query that should be observed.
	Query string

	// DataMayNotExist indicates if the query may not return data until some event occurs in the
	// future.
	//
	// For example, repo_updater_memory_usage should always have data present and an alert should
	// fire if for some reason that query is not returning any data, so this would be set to false.
	// In contrast, search_error_rate would depend on users actually performing searches and we
	// would not want an alert to fire if no data was present, so this would be set to true.
	DataMayNotExist bool

	// DataMayBeNaN indicates whether or not the query may return NaN regularly. Most often,
	// this should be false as NaN often indicates a mistaken divide by zero. However, for
	// some queries NaN values may be expected, in which case you should set this to true.
	//
	// When false, alerts will fire if the query returns NaN.
	DataMayBeNaN bool

	// Warning and Critical alert definitions. At least a Warning alert must be present.
	//
	// See README.md for why it is intentionally impossible to create a dashboard to monitor
	// something without at least a warning alert being defined.
	Warning, Critical Alert

	// PossibleSolutions is Markdown describing possible solutions in the event that the alert is
	// firing. If there is no clear potential resolution, "none" must be explicitly stated.
	//
	// Contacting support should not be mentioned as part of a possible solution, as it is
	// communicated elsewhere.
	//
	// To make writing the Markdown more friendly in Go, string literals like this:
	//
	// 	Observable{
	// 		PossibleSolutions: `
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
	//
	PossibleSolutions string

	// PanelOptions describes some options for how to render the metric in the Grafana panel.
	PanelOptions panelOptions
}

func (o Observable) validate() error {
	if strings.Contains(o.Name, " ") || strings.ToLower(o.Name) != o.Name {
		return fmt.Errorf("Observable.Name must be in lower_snake_case; found \"%s\"", o.Name)
	}
	if v := string([]rune(o.Description)[0]); v != strings.ToLower(v) {
		return fmt.Errorf("Observable.Description must be lowercase; found \"%s\"", o.Description)
	}
	if o.Warning.isEmpty() && o.Critical.isEmpty() {
		return fmt.Errorf("%s: a Warning or Critical alert MUST be defined", o.Name)
	}
	if err := o.Warning.validate(); err != nil && !o.Warning.isEmpty() {
		return fmt.Errorf("Warning: %v", err)
	}
	if err := o.Critical.validate(); err != nil && !o.Critical.isEmpty() {
		return fmt.Errorf("Critical: %v", err)
	}
	if l := strings.ToLower(o.PossibleSolutions); strings.Contains(l, "contact support") || strings.Contains(l, "contact us") {
		return fmt.Errorf("PossibleSolutions: should not include mentions of contacting support")
	}
	if o.PossibleSolutions == "" {
		return fmt.Errorf(`PossibleSolutions: must list solutions or "none"`)
	} else if o.PossibleSolutions != "none" {
		if _, err := goMarkdown(o.PossibleSolutions); err != nil {
			return fmt.Errorf("PossibleSolutions: %v", err)
		}
	}
	return nil
}

// Alert defines when an alert would be considered firing.
type Alert struct {
	// GreaterOrEqual, when non-zero, indicates the alert should fire when
	// greater or equal to this value.
	GreaterOrEqual float64

	// LessOrEqual, when non-zero, indicates the alert should fire when less
	// than or equal to this value.
	LessOrEqual float64
}

func (a Alert) isEmpty() bool {
	return a == Alert{} || (a.GreaterOrEqual == 0 && a.LessOrEqual == 0)
}

func (a Alert) validate() error {
	if a.isEmpty() {
		return errors.New("empty")
	}
	if a.GreaterOrEqual != 0 && a.LessOrEqual != 0 {
		return errors.New("only one of GreaterOrEqual,LessOrEqual may be specified")
	}
	return nil
}

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
	default:
		panic("never here")
	}
}

// From https://sourcegraph.com/github.com/grafana/grafana@b63b82976b3708b082326c0b7d42f38d4bc261fa/-/blob/packages/grafana-data/src/valueFormats/categories.ts#L23
const (
	// Number is the default unit type.
	Number UnitType = "short"

	// Milliseconds for representing time.
	Milliseconds UnitType = "dtdurationms"

	// Seconds for representing time.
	Seconds UnitType = "dtdurations"

	// Percentage in the range of 0-100.
	Percentage UnitType = "percent"

	// Bytes in IEC (1024) format, e.g. for representing storage sizes.
	Bytes UnitType = "bytes"

	// BitsPerSecond, e.g. for representing network and disk IO.
	BitsPerSecond UnitType = "bps"
)

type panelOptions struct {
	min, max     *float64
	minAuto      bool
	legendFormat string
	unitType     UnitType
}

// Min sets the minimum value of the Y axis on the panel. The default is zero.
func (p panelOptions) Min(min float64) panelOptions {
	p.min = &min
	return p
}

// Min sets the minimum value of the Y axis on the panel to auto, instead of
// the default zero.
//
// This is generally only useful if trying to show negative numbers.
func (p panelOptions) MinAuto() panelOptions {
	p.minAuto = true
	return p
}

// Max sets the maximum value of the Y axis on the panel. The default is auto.
func (p panelOptions) Max(min float64) panelOptions {
	p.min = &min
	return p
}

// LegendFormat sets the panel's legend format, which may use Go template strings to select
// labels from the Prometheus query.
func (p panelOptions) LegendFormat(format string) panelOptions {
	p.legendFormat = format
	return p
}

// Unit sets the panel's Y axis unit type.
func (p panelOptions) Unit(t UnitType) panelOptions {
	p.unitType = t
	return p
}

func (p panelOptions) withDefaults() panelOptions {
	if p.min == nil && !p.minAuto {
		defaultMin := 0.0
		p.min = &defaultMin
	}
	if p.legendFormat == "" {
		p.legendFormat = "value"
	}
	if p.unitType == "" {
		p.unitType = Number
	}
	return p
}

func PanelOptions() panelOptions { return panelOptions{} }

// dashboard generates the Grafana dashboard for this container.
func (c *Container) dashboard() *sdk.Board {
	board := sdk.NewBoard(c.Title)
	board.Version = uint(rand.Uint32())
	board.UID = c.Name
	board.ID = 0
	board.Timezone = "utc"
	board.Timepicker.RefreshIntervals = []string{"5s", "10s", "30s", "1m", "5m", "15m", "30m", "1h", "2h", "1d"}
	board.Time.From = "now-6h"
	board.Time.To = "now"
	board.SharedCrosshair = true
	board.AddTags("builtin")
	board.Templating.List = []sdk.TemplateVar{
		{
			Label:      "Filter alert level",
			Name:       "alert_level",
			AllValue:   ".*",
			Current:    sdk.Current{Text: "all", Value: "$__all"},
			IncludeAll: true,
			Options: []sdk.Option{
				{Text: "all", Value: "$__all", Selected: true},
				{Text: "critical", Value: "critical"},
				{Text: "warning", Value: "warning"},
			},
			Query: "critical,warning",
			Type:  "custom",
		},
	}

	description := sdk.NewText("")
	description.Title = "" // Removes vertical space the title would otherwise take up
	setPanelSize(description, 24, 3)
	description.TextPanel.Mode = "html"
	description.TextPanel.Content = fmt.Sprintf(`
	<div style="text-align: left;">
	  <img src="https://storage.googleapis.com/sourcegraph-assets/sourcegraph-logo-light.png" style="height:30px; margin:0.5rem"></img>
	  <div style="margin-left: 1rem; margin-top: 0.5rem; font-size: 20px;"><span style="color: #8e8e8e">%s:</span> %s <a style="font-size: 15px" target="_blank" href="https://docs.sourcegraph.com/dev/architecture">(⧉ architecture diagram)</a></span>
	</div>
	`, c.Name, c.Description)
	board.Panels = append(board.Panels, description)

	alertsDefined := sdk.NewTable("Alerts defined")
	setPanelSize(alertsDefined, 9, 5)
	setPanelPos(alertsDefined, 0, 3)
	alertsDefined.TablePanel.Sort = &sdk.Sort{Desc: true, Col: 4}
	alertsDefined.TablePanel.Styles = []sdk.ColumnStyle{
		{
			Pattern: "Time",
			Type:    "hidden",
		},
		{
			Pattern: "level",
			Type:    "hidden",
		},
		{
			Pattern: "_01_level",
			Alias:   stringPtr("level"),
		},
		{
			Pattern:     "Value",
			Alias:       stringPtr("firing?"),
			ColorMode:   stringPtr("row"),
			Colors:      &[]string{"rgba(50, 172, 45, 0.97)", "rgba(237, 129, 40, 0.89)", "rgba(245, 54, 54, 0.9)"},
			Thresholds:  &[]string{"0.99999", "1"},
			Type:        "string",
			MappingType: 1,
			ValueMaps: []sdk.ColumnStyleValueMap{
				{Text: "false", Value: "0"},
				{Text: "true", Value: "1"},
			},
		},
	}
	alertsDefined.AddTarget(&sdk.Target{
		Expr:    fmt.Sprintf(`label_replace(sum(max by (level,service_name,name,description)(alert_count{service_name="%s",name!="",level=~"$alert_level"})) by (level,description), "_01_level", "$1", "level", "(.*)")`, c.Name),
		Format:  "table",
		Instant: true,
	})
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
			Show:     true,
		},
		{
			Format:  "short",
			LogBase: 1,
			Show:    true,
		},
	}
	alertsFiring.AddTarget(&sdk.Target{
		Expr:         fmt.Sprintf(`sum by (service_name,level,name)(max by (level,service_name,name,description)(alert_count{service_name="%s",name!="",level=~"$alert_level"}) >= 1)`, c.Name),
		LegendFormat: "{{level}}: {{name}}",
	})
	board.Panels = append(board.Panels, alertsFiring)

	baseY := 8
	offsetY := baseY
	for _, group := range c.Groups {
		// Non-general groups are shown as collapsible panels.
		var rowPanel *sdk.Panel
		if group.Title != "General" {
			rowPanel = &sdk.Panel{RowPanel: &sdk.RowPanel{}}
			rowPanel.OfType = sdk.RowType
			rowPanel.Type = "row"
			rowPanel.Title = group.Title
			offsetY++
			setPanelPos(rowPanel, 0, offsetY)
			rowPanel.Collapsed = group.Hidden
			rowPanel.Panels = []sdk.Panel{} // cannot be null
			board.Panels = append(board.Panels, rowPanel)
		}

		// Generate a panel for displaying each observable in each row.
		for _, row := range group.Rows {
			panelWidth := 24 / len(row)
			offsetY++
			for i, o := range row {
				panelTitle := strings.ToTitle(string([]rune(o.Description)[0])) + string([]rune(o.Description)[1:])
				panel := sdk.NewGraph(panelTitle)
				setPanelSize(panel, panelWidth, 5)
				setPanelPos(panel, i*panelWidth, offsetY)
				panel.GraphPanel.Legend.Show = true
				panel.GraphPanel.Fill = 1
				panel.GraphPanel.Lines = true
				panel.GraphPanel.Linewidth = 1
				panel.GraphPanel.NullPointMode = "connected"
				panel.GraphPanel.Pointradius = 2
				panel.GraphPanel.AliasColors = map[string]string{}
				panel.GraphPanel.Xaxis = sdk.Axis{
					Show: true,
				}

				opt := o.PanelOptions.withDefaults()
				leftAxis := sdk.Axis{
					Decimals: 0,
					Format:   string(opt.unitType),
					LogBase:  1,
					Show:     true,
				}

				if o.Warning.GreaterOrEqual != 0 {
					// Warning threshold
					panel.GraphPanel.Thresholds = append(panel.GraphPanel.Thresholds, sdk.Threshold{
						Value:     float32(o.Warning.GreaterOrEqual),
						Op:        "gt",
						ColorMode: "custom",
						Fill:      true,
						Line:      false,
						FillColor: "rgba(255, 73, 53, 0.8)",
					})
				}
				if o.Critical.GreaterOrEqual != 0 {
					// Critical threshold
					panel.GraphPanel.Thresholds = append(panel.GraphPanel.Thresholds, sdk.Threshold{
						Value:     float32(o.Critical.GreaterOrEqual),
						Op:        "gt",
						ColorMode: "custom",
						Fill:      true,
						Line:      false,
						FillColor: "rgba(255, 17, 36, 0.8)",
					})
				}
				if o.Warning.LessOrEqual != 0 {
					// Warning threshold
					panel.GraphPanel.Thresholds = append(panel.GraphPanel.Thresholds, sdk.Threshold{
						Value:     float32(o.Warning.LessOrEqual),
						Op:        "lt",
						ColorMode: "custom",
						Fill:      true,
						Line:      false,
						FillColor: "rgba(255, 73, 53, 0.8)",
					})
				}
				if o.Critical.LessOrEqual != 0 {
					// Critical threshold
					panel.GraphPanel.Thresholds = append(panel.GraphPanel.Thresholds, sdk.Threshold{
						Value:     float32(o.Critical.LessOrEqual),
						Op:        "lt",
						ColorMode: "custom",
						Fill:      true,
						Line:      false,
						FillColor: "rgba(255, 17, 36, 0.8)",
					})
				}

				if opt.min != nil {
					leftAxis.Min = sdk.NewFloatString(*opt.min)
				}
				if opt.max != nil {
					leftAxis.Max = sdk.NewFloatString(*opt.max)
				}
				panel.GraphPanel.Yaxes = []sdk.Axis{
					leftAxis,
					{
						Format:  "short",
						LogBase: 1,
						Show:    true,
					},
				}
				panel.AddTarget(&sdk.Target{
					Expr:         o.Query,
					LegendFormat: opt.legendFormat,
				})
				if rowPanel != nil && group.Hidden {
					rowPanel.RowPanel.Panels = append(rowPanel.RowPanel.Panels, *panel)
				} else {
					board.Panels = append(board.Panels, panel)
				}
			}
		}
	}
	return board
}

// alertDescription generates an alert description for the specified coontainer's alert.
func (c *Container) alertDescription(o Observable, alert Alert) string {
	if alert.isEmpty() {
		panic("never here")
	}
	if alert.GreaterOrEqual != 0 {
		// e.g. "zoekt-indexserver: 20+ indexed search request errors every 5m by code"
		return fmt.Sprintf("%s: %v%s+ %s", c.Name, alert.GreaterOrEqual, o.PanelOptions.unitType.short(), o.Description)
	} else if alert.LessOrEqual != 0 {
		// e.g. "zoekt-indexserver: less than 20 indexed search requests every 5m by code"
		return fmt.Sprintf("%s: less than %v%s %s", c.Name, alert.LessOrEqual, o.PanelOptions.unitType.short(), o.Description)
	}
	panic("never here")
}

// promAlertsFile generates the Prometheus rules file which defines our
// high-level alerting metrics for the container. For more information about
// how these work, see:
//
// https://docs.sourcegraph.com/admin/observability/metrics_guide#high-level-alerting-metrics
//
func (c *Container) promAlertsFile() *promRulesFile {
	f := &promRulesFile{}
	group := promGroup{Name: c.Name}
	for _, g := range c.Groups {
		for _, r := range g.Rows {
			for _, o := range r {
				for level, alert := range map[string]Alert{
					"warning":  o.Warning,
					"critical": o.Critical,
				} {
					if alert.isEmpty() {
						continue
					}
					labels := map[string]string{}
					labels["service_name"] = c.Name
					labels["level"] = level
					labels["name"] = o.Name
					labels["description"] = c.alertDescription(o, alert)

					// The alertQuery must contribute a query that returns a value < 1 when it is not
					// firing, or a value of >= 1 when it is firing.
					var alertQuery string
					if alert.GreaterOrEqual != 0 {
						// By dividing the query value and the greaterOrEqual value, we produce a
						// value of 1 when the query reaches the greaterOrEqual value and < 1
						// otherwise. Examples:
						//
						// 	query_value=50 / greaterOrEqual=50 == 1.0
						// 	query_value=25 / greaterOrEqual=50 == 0.5
						// 	query_value=0 / greaterOrEqual=50 == 0.0
						//
						alertQuery = fmt.Sprintf("(%s) / %v", o.Query, alert.GreaterOrEqual)

						// Replace no-data with zero values, so the alert does not fire, if desired.
						if o.DataMayNotExist {
							alertQuery = fmt.Sprintf("(%s) OR on() vector(0)", alertQuery)
						}

						// Replace NaN values with zero (not firing) or one (firing) if they are present.
						fireOnNan := "1"
						if o.DataMayBeNaN {
							fireOnNan = "0"
						}
						alertQuery = fmt.Sprintf("((%s) >= 0) OR on() vector(%v)", alertQuery, fireOnNan)
					} else if alert.LessOrEqual != 0 {
						//
						// 	lessOrEqual=50 / query_value=100 == 0.5
						// 	lessOrEqual=50 / query_value=50 == 1.0
						// 	lessOrEqual=50 / query_value=25 == 2.0
						// 	lessOrEqual=50 / query_value=0 (0.0000001) == 500000000
						// 	lessOrEqual=50 / query_value=-50 (0.0000001) == 500000000
						//
						alertQuery = fmt.Sprintf("%v / clamp_min(%s, 0.0000001)", alert.LessOrEqual, o.Query)

						// Replace no-data with zero values, so the alert does not fire, if desired.
						if o.DataMayNotExist {
							alertQuery = fmt.Sprintf("(%s) OR on() vector(0)", alertQuery)
						}

						// Replace NaN values with zero (not firing) or one (firing) if they are present.
						fireOnNan := "1"
						if o.DataMayBeNaN {
							fireOnNan = "0"
						}
						alertQuery = fmt.Sprintf("((%s) >= 0) OR on() vector(%v)", alertQuery, fireOnNan)
					}

					// This wrapper clamp/floor/default vector should be present on ALL alert_count rule
					// definitions because:
					//
					// 1. Clamping and flooring ensures that a single alert definition can only ever
					//    contribute a single 0 OR 1 value, and as such cannot artificially inflate
					//    alert_count or cause it to become a non-whole number.
					//
					// 3. "OR on() vector(1)" ensures that the alert is always firing if the inner
					//    alertQuery does not return values for any reason (e.g. the query is for a
					//    metric that does not exist.)
					//
					expr := "clamp_max(clamp_min(floor(\n" + alertQuery + "\n), 0), 1) OR on() vector(1)"
					group.Rules = append(group.Rules, promRule{
						Record: "alert_count",
						Labels: labels,
						Expr:   expr,
					})
				}
			}
		}
	}
	f.Groups = append(f.Groups, group)
	return f
}

// isValidUID checks if the given string is a valid UID for entry into a Grafana dashboard. This is
// primarily used in the URL, e.g. /-/debug/grafana/d/syntect-server/<UID> and allows us to have
// static URLs we can document like:
//
// 	Go to https://sourcegraph.example.com/-/debug/grafana/d/syntect-server/syntect-server
//
// Instead of having to describe all the steps to navigate there because the UID is random.
func isValidUID(s string) bool {
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

// upperFirst returns s with an uppercase first rune.
func upperFirst(s string) string {
	return strings.ToUpper(string([]rune(s)[0])) + string([]rune(s)[1:])
}

// withPeriod returns s ending with a period.
func withPeriod(s string) string {
	if !strings.HasSuffix(s, ".") {
		return s + "."
	}
	return s
}

func generateDocs(containers []*Container) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, `# Alert solutions

This document contains possible solutions for when you find alerts are firing in Sourcegraph's monitoring.
If your alert isn't mentioned here, or if the solution doesn't help, [contact us](mailto:support@sourcegraph.com)
for assistance.

<!-- DO NOT EDIT: generated via: go generate ./monitoring -->

`)
	for _, c := range containers {
		for _, g := range c.Groups {
			for _, r := range g.Rows {
				for _, o := range r {
					if o.PossibleSolutions == "none" {
						continue
					}

					fmt.Fprintf(&b, "# %s: %s\n\n", c.Name, o.Name)

					fmt.Fprintf(&b, "**Descriptions:**\n")
					for _, alert := range []Alert{
						o.Warning,
						o.Critical,
					} {
						if alert.isEmpty() {
							continue
						}
						fmt.Fprintf(&b, "\n- _%s_\n\n", c.alertDescription(o, alert))
					}

					fmt.Fprintf(&b, "**Possible solutions:**\n\n")
					possibleSolutions, _ := goMarkdown(o.PossibleSolutions)
					fmt.Fprintf(&b, "%s\n\n", possibleSolutions)
				}
			}
		}
	}
	return b.Bytes()
}

func goMarkdown(m string) (string, error) {
	m = strings.TrimPrefix(m, "\n")

	// Replace single quotes with backticks.
	// Replace escaped single quotes with single quotes.
	m = strings.Replace(m, `\'`, `$ESCAPED_SINGLE_QUOTE`, -1)
	m = strings.Replace(m, `'`, "`", -1)
	m = strings.Replace(m, `$ESCAPED_SINGLE_QUOTE`, "'", -1)

	// Unindent based on the indention of the last line.
	lines := strings.Split(m, "\n")
	baseIndention := lines[len(lines)-1]
	if strings.TrimSpace(baseIndention) == "" {
		if strings.Contains(baseIndention, " ") {
			return "", errors.New("Go string literal indention must be tabs")
		}
		indentionLevel := strings.Count(baseIndention, "\t")
		removeIndention := strings.Repeat("\t", indentionLevel+1)
		for i, l := range lines[:len(lines)-1] {
			newLine := strings.TrimPrefix(l, removeIndention)
			if l == newLine {
				return "", fmt.Errorf("inconsistent indention (line %d %q expected to start with %q)", i, l, removeIndention)
			}
			lines[i] = newLine
		}
		m = strings.Join(lines[:len(lines)-1], "\n")
	}
	return m, nil
}

var isDev, _ = strconv.ParseBool(os.Getenv("DEV"))

func main() {
	grafanaDir, ok := os.LookupEnv("GRAFANA_DIR")
	if !ok {
		grafanaDir = "../docker-images/grafana/config/provisioning/dashboards/sourcegraph/"
	}
	prometheusDir, ok := os.LookupEnv("PROMETHEUS_DIR")
	if !ok {
		prometheusDir = "../docker-images/prometheus/config/"
	}
	docSolutionsFile, ok := os.LookupEnv("DOC_SOLUTIONS_FILE")
	if !ok {
		docSolutionsFile = "../doc/admin/observability/alert_solutions.md"
	}

	reloadValue, ok := os.LookupEnv("RELOAD")
	if !ok && isDev {
		reloadValue = "true"
	}
	reload, _ := strconv.ParseBool(reloadValue)

	containers := []*Container{
		Frontend(),
		GitServer(),
		GitHubProxy(),
		PreciseCodeIntelBundleManager(),
		PreciseCodeIntelWorker(),
		PreciseCodeIntelIndexer(),
		QueryRunner(),
		Replacer(),
		RepoUpdater(),
		Searcher(),
		Symbols(),
		SyntectServer(),
		ZoektIndexServer(),
		ZoektWebServer(),
	}
	for _, container := range containers {
		if err := container.validate(); err != nil {
			log.Fatal(err)
		}
		if grafanaDir != "" {
			board := container.dashboard()
			data, err := json.MarshalIndent(board, "", "  ")
			if err != nil {
				log.Fatal(err)
			}
			err = ioutil.WriteFile(filepath.Join(grafanaDir, container.Name+".json"), data, 0666)
			if err != nil {
				log.Fatal(err)
			}
			if reload {
				ctx := context.Background()
				client := sdk.NewClient("http://127.0.0.1:3370", "admin:admin", sdk.DefaultHTTPClient)
				_, err := client.SetDashboard(ctx, *board, sdk.SetDashboardParams{Overwrite: true})
				if err != nil {
					log.Fatal("updating dashboard:", err)
				}
			}
		}

		if prometheusDir != "" {
			promAlertsFile := container.promAlertsFile()
			data, err := yaml.Marshal(promAlertsFile)
			if err != nil {
				log.Fatal(err)
			}
			fileName := strings.Replace(container.Name, "-", "_", -1) + "_alert_rules.yml"
			err = ioutil.WriteFile(filepath.Join(prometheusDir, fileName), data, 0666)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	if prometheusDir != "" && reload {
		resp, err := http.Post("http://127.0.0.1:9090/-/reload", "", nil)
		if err != nil {
			log.Fatal("reloading Prometheus rules, got error:", err)
		}
		if resp.StatusCode != 200 {
			log.Fatal("reloading Prometheus rules, got status code:", resp.StatusCode)
		}
	}
	if reload && grafanaDir != "" && prometheusDir != "" {
		fmt.Println("Reloaded Prometheus rules & Grafana dashboards")
	}

	if docSolutionsFile != "" {
		solutions := generateDocs(containers)
		err := ioutil.WriteFile(docSolutionsFile, solutions, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// promRulesFile represents a Prometheus recording rules file (which we use for defining our alerts)
// see:
//
// https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/
//
type promRulesFile struct {
	Groups []promGroup
}

type promGroup struct {
	Name  string
	Rules []promRule
}

type promRule struct {
	Record string
	Labels map[string]string
	Expr   string
}

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

func stringPtr(s string) *string {
	return &s
}
