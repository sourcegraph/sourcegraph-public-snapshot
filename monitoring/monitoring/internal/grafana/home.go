package grafana

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"text/template"

	"github.com/prometheus/prometheus/model/labels"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring/internal/promql"
)

// homeJson is the raw dashboards JSON for the home dashboard.
//
//go:embed home.json.tmpl
var homeJsonTmpl string

// Home is the definition for the home dashboard. It is provided as raw JSON because it
// is defined outside of the monitoring generator.
func Home(folder string, injectLabelMatchers []*labels.Matcher) ([]byte, error) {
	// Build template variables
	vars := map[string]string{
		"WarningAlertsExpr":       "sum by (service_name)(max by (level,service_name,name,description)(alert_count{name!=\"\",level=\"warning\"}))",
		"CriticalAlertsExpr":      "sum by (service_name)(max by (level,service_name,name,description)(alert_count{name!=\"\",level=\"critical\"}))",
		"AlertCountByServiceExpr": "count(sum(alert_count{name!=\"\"}) by (service_name))",
		"AlertCountByLevelExpr":   "count(sum(alert_count{name!=\"\"}) by (level,description))",
		"AlertLabelQuery":         "sum by (level,service_name,description,grafana_panel_id)(max by (level,service_name,name,description,grafana_panel_id)(alert_count{name!=\"\"}))",
	}
	for k, v := range vars {
		var err error
		vars[k], err = promql.InjectMatchers(v, injectLabelMatchers, nil)
		if err != nil {
			return nil, errors.Wrap(err, k)
		}
	}

	// Add static vars
	uid := "overview"
	if folder != "" {
		uid = fmt.Sprintf("%s-%s", folder, uid)
	}
	vars["UID"] = uid

	// Build and execute template
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"escape": func(val string) string {
			quoted := strconv.Quote(val)
			return quoted[1 : len(quoted)-1] // strip leading and trailing quotes
		},
	}).Parse(homeJsonTmpl)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, &vars); err != nil {
		return nil, err
	}

	// Validate JSON
	data := buf.Bytes()
	if err := json.Unmarshal(data, &map[string]interface{}{}); err != nil {
		return nil, errors.Wrap(err, "generated dashboard is invalid")
	}

	return data, nil
}
