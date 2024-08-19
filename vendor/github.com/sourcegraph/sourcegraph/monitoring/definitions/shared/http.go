package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

type http struct{}

var HTTP http

func (http) NewHandlersGroup(name string) monitoring.Group {
	return monitoring.Group{
		Title:  "HTTP handlers",
		Hidden: true,
		Rows: []monitoring.Row{
			{
				{
					Name:           "healthy_request_rate",
					Description:    "requests per second, by route, when status code is 200",
					Query:          fmt.Sprintf("sum by (route) (rate(src_http_request_duration_seconds_count{app=\"%s\",code=~\"2..\"}[5m]))", name),
					NoAlert:        true,
					Panel:          monitoring.Panel().LegendFormat("{{route}}").Unit(monitoring.Number),
					Owner:          monitoring.ObservableOwnerSource,
					Interpretation: "The number of healthy HTTP requests per second to internal HTTP api",
				},
				{
					Name:           "unhealthy_request_rate",
					Description:    "requests per second, by route, when status code is not 200",
					Query:          fmt.Sprintf("sum by (route) (rate(src_http_request_duration_seconds_count{app=\"%s\",code!~\"2..\"}[5m]))", name),
					NoAlert:        true,
					Panel:          monitoring.Panel().LegendFormat("{{route}}").Unit(monitoring.Number),
					Owner:          monitoring.ObservableOwnerSource,
					Interpretation: "The number of unhealthy HTTP requests per second to internal HTTP api",
				},
				{
					Name:           "request_rate_by_code",
					Description:    "requests per second, by status code",
					Query:          fmt.Sprintf("sum by (code) (rate(src_http_request_duration_seconds_count{app=\"%s\"}[5m]))", name),
					NoAlert:        true,
					Panel:          monitoring.Panel().LegendFormat("{{code}}").Unit(monitoring.Number),
					Owner:          monitoring.ObservableOwnerSource,
					Interpretation: "The number of HTTP requests per second by code",
				},
			},
			{
				{
					Name:           "95th_percentile_healthy_requests",
					Description:    "95th percentile duration by route, when status code is 200",
					Query:          fmt.Sprintf("histogram_quantile(0.95, sum(rate(src_http_request_duration_seconds_bucket{app=\"%s\",code=~\"2..\"}[5m])) by (le, route))", name),
					NoAlert:        true,
					Panel:          monitoring.Panel().LegendFormat("{{route}}").Unit(monitoring.Seconds),
					Owner:          monitoring.ObservableOwnerSource,
					Interpretation: "The 95th percentile duration by route when the status code is 200 ",
				},
				{
					Name:           "95th_percentile_unhealthy_requests",
					Description:    "95th percentile duration by route, when status code is not 200",
					Query:          fmt.Sprintf("histogram_quantile(0.95, sum(rate(src_http_request_duration_seconds_bucket{app=\"%s\",code!~\"2..\"}[5m])) by (le, route))", name),
					NoAlert:        true,
					Panel:          monitoring.Panel().LegendFormat("{{route}}").Unit(monitoring.Seconds),
					Owner:          monitoring.ObservableOwnerSource,
					Interpretation: "The 95th percentile duration by route when the status code is not 200 ",
				},
			},
		},
	}
}
