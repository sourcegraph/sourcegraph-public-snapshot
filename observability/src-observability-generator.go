//go:generate echo "Regenerating observability..."
//go:generate go build -o /tmp/src-observability-generator
//go:generate /tmp/src-observability-generator

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	// Name is the name of the Docker container, e.g. "syntect-server".
	Name string

	// Title is the title of the Docker container, e.g. "Syntect Server".
	Title string

	// Description is the description of the Docker container. It should describe what the
	// container is responsible for, so that the impact of issues in it is clear.
	Description string

	// Rows
	Rows []Row
}

type Row struct {
	Title       string
	Observables []Observable
}

func (r Row) validate() error {
	for _, o := range r.Observables {
		if err := o.validate(); err != nil {
			return err
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

	// Warning and Critical alert definitions. At least a Warning alert must be present.
	//
	// See README.md for why it is intentionally impossible to create a dashboard to monitor
	// something without at least a warning alert being defined.
	Warning, Critical Alert

	// PanelOptions describes some options for how to render the metric in the Grafana panel.
	PanelOptions panelOptions
}

func (o Observable) validate() error {
	if strings.Contains(o.Name, " ") || strings.ToLower(o.Name) != o.Name {
		return fmt.Errorf("Observable names should be in lower_snake_case: %q", o.Name)
	}
	if o.Warning.isEmpty() && o.Critical.isEmpty() {
		return fmt.Errorf("%s: a Warning or Critical alert MUST be defined", o.Name)
	}
	return nil
}

// Alert defines when an alert would be considered firing.
type Alert struct {
	// GreaterOrEqual indicates the value at which the alert should fire.
	GreaterOrEqual float64
}

func (a Alert) isEmpty() bool {
	return a == Alert{} || a.GreaterOrEqual <= 0
}

// UnitType for controlling the unit type display on graphs.
type UnitType string

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
	if !isValidUID(c.Name) {
		panic(fmt.Sprintf("expected Name to be alphanumeric + dashes: found \"%s\"", c.Name))
	}

	board := sdk.NewBoard(c.Title)

	// Note: being able to test edits quickly to dashboards is useful, but without setting this expansion and
	// unexpansion of rows counts as an "edit" and the site admin would be warned about an unsaved change that
	// they cannot actually save when navigating away, so we set this when not in dev environments.
	board.Editable = isDev

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
	alertsDefined.TablePanel.Sort = &sdk.Sort{Desc: true}
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
		Expr:    fmt.Sprintf(`label_replace(sum(alert_count{service_name="%s",name!="",level=~"$alert_level"}) by (level,description), "_01_level", "$1", "level", "(.*)")`, c.Name),
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
		Expr:         fmt.Sprintf(`sum by (service_name,level,name)(alert_count{service_name="%s",name!="",level=~"$alert_level"} >= 1)`, c.Name),
		LegendFormat: "{{level}}: {{name}}",
	})
	board.Panels = append(board.Panels, alertsFiring)

	baseY := 8
	for rowIndex, row := range c.Rows {
		var rowPanel *sdk.Panel
		if row.Title != "General" {
			rowPanel = &sdk.Panel{RowPanel: &sdk.RowPanel{}}
			rowPanel.OfType = sdk.RowType
			rowPanel.Type = "row"
			rowPanel.Title = row.Title
			setPanelPos(rowPanel, 0, baseY+rowIndex)
			rowPanel.Collapsed = true
			board.Panels = append(board.Panels, rowPanel)
		}

		panelWidth := 24 / len(row.Observables)
		for i, o := range row.Observables {
			panel := sdk.NewGraph(strings.ToTitle(string([]rune(o.Description)[0])) + string([]rune(o.Description)[1:]))
			setPanelSize(panel, panelWidth, 5)
			setPanelPos(panel, i*panelWidth, baseY+rowIndex)
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

			if o.Warning.GreaterOrEqual > 0 {
				// Warning threshold
				panel.GraphPanel.Thresholds = append(panel.GraphPanel.Thresholds, sdk.Threshold{
					Value:     float32(o.Warning.GreaterOrEqual),
					Op:        "gt",
					ColorMode: "custom",
					Fill:      true,
					Line:      true,
					FillColor: "rgba(255, 152, 48, 0.5)",
					LineColor: "rgba(31, 96, 196, 0.6)",
				})
			}
			if o.Critical.GreaterOrEqual > 0 {
				// Critical threshold
				panel.GraphPanel.Thresholds = append(panel.GraphPanel.Thresholds, sdk.Threshold{
					Value:     float32(o.Critical.GreaterOrEqual),
					Op:        "gt",
					ColorMode: "custom",
					Fill:      true,
					Line:      true,
					FillColor: "rgba(242, 73, 92, 0.5)",
					LineColor: "rgba(31, 96, 196, 0.6)",
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
			if rowPanel != nil {
				rowPanel.RowPanel.Panels = append(rowPanel.RowPanel.Panels, *panel)
			} else {
				board.Panels = append(board.Panels, panel)
			}
		}
	}
	return board
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
	for _, row := range c.Rows {
		for _, o := range row.Observables {
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

				var alertQuery string
				if alert.GreaterOrEqual > 0 {
					labels["description"] = fmt.Sprintf("%s: %v+ %s", c.Name, alert.GreaterOrEqual, o.Description)
					alertQuery = fmt.Sprintf("(%s) / %v", o.Query, alert.GreaterOrEqual)
					if o.DataMayNotExist {
						alertQuery = fmt.Sprintf("(%s) OR on() vector(0)", alertQuery)
					}
				}
				group.Rules = append(group.Rules, promRule{
					Record: "alert_count",
					Labels: labels,
					Expr:   "clamp_max(clamp_min(floor(\n" + alertQuery + "\n), 0), 1) OR on() vector(1)",
				})
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
	for _, r := range s {
		if !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-') {
			return false
		}
	}
	return true
}

var isDev, _ = strconv.ParseBool(os.Getenv("DEV"))

func main() {
	grafanaDir, ok := os.LookupEnv("DASHBOARD_DIR")
	if !ok {
		grafanaDir = "../docker-images/grafana/config/provisioning/dashboards/sourcegraph/"
	}
	prometheusDir, ok := os.LookupEnv("PROMETHEUS_DIR")
	if !ok {
		prometheusDir = "../docker-images/prometheus/config/"
	}

	reloadValue, ok := os.LookupEnv("RELOAD")
	if !ok && isDev {
		reloadValue = "true"
	}
	reload, _ := strconv.ParseBool(reloadValue)

	for _, container := range []*Container{
		SyntectServer(),
	} {
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

func uintPtr(x uint) *uint {
	return &x
}

func stringPtr(s string) *string {
	return &s
}
