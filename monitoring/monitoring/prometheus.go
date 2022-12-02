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

// prometheusAlertName creates an alertname that is unique given the combination of parameters
func prometheusAlertName(level, service, name string) string {
	return fmt.Sprintf("%s_%s_%s", level, service, name)
}

// promRule is a subset of a Prometheus recording or alert rule definition.
type promRule struct {
	// either Record or Alert
	Record string `yaml:",omitempty"` // https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/
	Alert  string `yaml:",omitempty"` // https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/

	Labels map[string]string `yaml:",omitempty"`
	Expr   string

	// for Alert only
	For *model.Duration `yaml:",omitempty"`
}

func (r *promRule) validate() error {
	if r.Record != "" && r.Alert != "" {
		return errors.Errorf("promRule cannot be both a record (%q) and an alert (%q)", r.Record, r.Alert)
	}
	if r.Alert == "" && r.For != nil {
		return errors.Errorf("promRule can only have a 'for' (%q) if it is an alert", r.For.String())
	}
	return nil
}

// promRulesFile represents a Prometheus recording rules file (which we use for defining our alerts)
// see:
//
// https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/
type promRulesFile struct {
	Groups []promGroup
}

type promGroup struct {
	Name  string
	Rules []promRule
}

func (g *promGroup) validate() error {
	if g.Name == "" {
		return errors.New("promGroup requires name")
	}
	for _, r := range g.Rules {
		if err := r.validate(); err != nil {
			return errors.Errorf("promGroup has invalid rule: %w", err)
		}
	}
	return nil
}

func (g *promGroup) appendRow(alertQuery string, labels map[string]string, duration time.Duration) {
	labels["alert_type"] = "builtin" // indicate alert is generated
	var forDuration *model.Duration
	if duration > 0 {
		d := model.Duration(duration)
		forDuration = &d
	}

	alertName := prometheusAlertName(labels["level"], labels["service_name"], labels["name"])
	g.Rules = append(g.Rules,
		// Native prometheus alert, based on alertQuery which returns 0 if not firing or 1 if firing.
		promRule{
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
		promRule{
			Record: "alert_count",
			Labels: labels,
			Expr:   fmt.Sprintf(`max(ALERTS{alertname=%q,alertstate="firing"} OR on() vector(0))`, alertName),
		})
}

func customPrometheusRules(injectLabelMatchers []*labels.Matcher) (*promRulesFile, error) {
	// Hardcode the desired label matcher values as labels
	labels := make(map[string]string)
	for _, matcher := range injectLabelMatchers {
		labels[matcher.Name] = matcher.Value
	}

	var injectErrors error
	injectExpr := func(expr string) string {
		injected, err := promql.InjectMatchers(expr, injectLabelMatchers, nil)
		if err != nil {
			injectErrors = errors.Append(injectErrors, err)
		}
		return injected
	}

	rulesFile := &promRulesFile{
		Groups: []promGroup{{
			Name: "cadvisor.rules",
			Rules: []promRule{{
				// The number of CPUs allocated to the container according to the configured Docker / Kubernetes limits.
				Record: "cadvisor_container_cpu_limit",
				Expr:   injectExpr("avg by (name)(container_spec_cpu_quota) / avg by (name)(container_spec_cpu_period)"),
				Labels: labels,
			}, {
				// Percentage of CPU cores the container consumed on average over a 1m period.
				// For example, if a container has a 4 CPU limit and this metric reports 50%,
				// it means the container consumed 2 cores on average over that 1m period.
				Record: "cadvisor_container_cpu_usage_percentage_total",
				Expr:   injectExpr("(avg by (name)(rate(container_cpu_usage_seconds_total[1m])) / cadvisor_container_cpu_limit) * 100.0"),
				Labels: labels,
			}, {
				// Percentage of memory usage the container is consuming.
				Record: "cadvisor_container_memory_usage_percentage_total",
				Expr:   injectExpr("max by (name)(container_memory_working_set_bytes / container_spec_memory_limit_bytes) * 100.0"),
				Labels: labels,
			}},
		}},
	}

	return rulesFile, injectErrors
}
