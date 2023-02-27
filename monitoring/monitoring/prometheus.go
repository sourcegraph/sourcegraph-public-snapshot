package monitoring

import (
	"fmt"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring/internal/promql"
)

const (
	alertRulesFileSuffix = "_alert_rules.yml"
)

var defaultRuleEvaluationInterval = model.Duration(30 * time.Second)

// prometheusAlertName creates an alertname that is unique given the combination of parameters
func prometheusAlertName(level, service, name string) string {
	return fmt.Sprintf("%s_%s_%s", level, service, name)
}

// PrometheusRule is a subset of a Prometheus recording or alert rule definition.
type PrometheusRule struct {
	// either Record or Alert
	Record string `yaml:",omitempty" json:"record,omitempty"` // https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/
	Alert  string `yaml:",omitempty" json:"alert,omitempty"`  // https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/

	Labels map[string]string `yaml:",omitempty" json:"labels,omitempty"`
	Expr   string            `json:"expr,omitempty"`

	// for Alert only
	For *model.Duration `yaml:",omitempty" json:"for,omitempty"`
}

func (r *PrometheusRule) validate() error {
	if r.Record != "" && r.Alert != "" {
		return errors.Errorf("promRule cannot be both a record (%q) and an alert (%q)", r.Record, r.Alert)
	}
	if r.Alert == "" && r.For != nil {
		return errors.Errorf("promRule can only have a 'for' (%q) if it is an alert", r.For.String())
	}
	return nil
}

// PrometheusRules represents a Prometheus recording rules file (which we use for defining our alerts)
// see:
//
// https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/
type PrometheusRules struct {
	Groups []PrometheusRuleGroup `json:"groups"`
}

type PrometheusRuleGroup struct {
	Name     string           `json:"name"`
	Rules    []PrometheusRule `json:"rules"`
	Interval *model.Duration  `json:"interval"`
}

func newPrometheusRuleGroup(name string) PrometheusRuleGroup {
	return PrometheusRuleGroup{Name: name, Interval: &defaultRuleEvaluationInterval}
}

func (g *PrometheusRuleGroup) validate() error {
	if g.Name == "" {
		return errors.New("PrometheusRuleGroup requires name")
	}
	if g.Interval == nil {
		return errors.New("PrometheusRuleGroup requires evaluation interval")
	}
	for _, r := range g.Rules {
		if err := r.validate(); err != nil {
			return errors.Errorf("PrometheusRuleGroup has invalid rule: %w", err)
		}
	}
	return nil
}

func (g *PrometheusRuleGroup) appendRow(alertQuery string, labels map[string]string, duration time.Duration) {
	labels["alert_type"] = "builtin" // indicate alert is generated
	var forDuration *model.Duration
	if duration > 0 {
		d := model.Duration(duration)
		forDuration = &d
	}

	alertName := prometheusAlertName(labels["level"], labels["service_name"], labels["name"])
	g.Rules = append(g.Rules,
		// Native prometheus alert, based on alertQuery which returns 0 if not firing or 1 if firing.
		PrometheusRule{
			Alert:  alertName,
			Labels: labels,
			Expr:   alertQuery,
			For:    forDuration,
		},
		// Record for generated alert, useful for indicating in Grafana dashboards if this alert
		// is defined at all. Prometheus's ALERTS metric does not track alerts with alertstate="inactive".
		//
		// Since ALERTS{alertname="value"} does not exist if the alert has never fired, we add set
		// the series to vector(0) instead.
		PrometheusRule{
			Record: "alert_count",
			Labels: labels,
			Expr:   fmt.Sprintf(`max(ALERTS{alertname=%q,alertstate="firing"} OR on() vector(0))`, alertName),
		})
}

func CustomPrometheusRules(injectLabelMatchers []*labels.Matcher) (*PrometheusRules, error) {
	// Hardcode the desired label matcher values as labels
	labelsMap := make(map[string]string)
	for _, matcher := range injectLabelMatchers {
		labelsMap[matcher.Name] = matcher.Value
	}

	var injectErrors error
	injectExpr := func(expr string) string {
		injected, err := promql.InjectMatchers(expr, injectLabelMatchers, nil)
		if err != nil {
			injectErrors = errors.Append(injectErrors, err)
		}
		return injected
	}

	rulesFile := &PrometheusRules{
		Groups: []PrometheusRuleGroup{{
			Name:     "cadvisor.rules",
			Interval: &defaultRuleEvaluationInterval,
			Rules: []PrometheusRule{{
				// The number of CPUs allocated to the container according to the configured Docker / Kubernetes limits.
				Record: "cadvisor_container_cpu_limit",
				Expr:   injectExpr("avg by (name)(container_spec_cpu_quota) / avg by (name)(container_spec_cpu_period)"),
				Labels: labelsMap,
			}, {
				// Percentage of CPU cores the container consumed on average over a 1m period.
				// For example, if a container has a 4 CPU limit and this metric reports 50%,
				// it means the container consumed 2 cores on average over that 1m period.
				Record: "cadvisor_container_cpu_usage_percentage_total",
				Expr:   injectExpr("(avg by (name)(rate(container_cpu_usage_seconds_total[1m])) / cadvisor_container_cpu_limit) * 100.0"),
				Labels: labelsMap,
			}, {
				// Percentage of memory usage the container is consuming.
				Record: "cadvisor_container_memory_usage_percentage_total",
				Expr:   injectExpr("max by (name)(container_memory_working_set_bytes / container_spec_memory_limit_bytes) * 100.0"),
				Labels: labelsMap,
			}},
		}},
	}

	return rulesFile, injectErrors
}
