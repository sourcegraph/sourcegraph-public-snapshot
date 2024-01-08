package opencodegraph

import (
	"context"
	"fmt"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	RegisterProvider(grafanaProvider{})
}

type grafanaProvider struct{}

func (grafanaProvider) Name() string { return "grafana" }

func (grafanaProvider) Capabilities(ctx context.Context, params schema.CapabilitiesParams) (*schema.CapabilitiesResult, error) {
	return &schema.CapabilitiesResult{
		Selector: []*schema.Selector{
			{Path: "**/*.go", ContentContains: "prometheus."},
		},
	}, nil
}

func (grafanaProvider) Annotations(ctx context.Context, params schema.AnnotationsParams) (*schema.AnnotationsResult, error) {
	var result schema.AnnotationsResult

	metrics, ranges := prometheusMetricDefs(params.Content)
	for i, metric := range metrics {
		id := fmt.Sprintf("%s:%d", metric, i)
		item := &schema.OpenCodeGraphItem{
			Id:    id,
			Title: "📟 Prometheus: " + metric,
			Url:   "https://sourcegraph.sourcegraph.com/-/debug/grafana/explore?orgId=1&left=%5B%22now-6h%22,%22now%22,%22Prometheus%22,%7B%22expr%22:%22" + metric + "%22,%22datasource%22:%22Prometheus%22,%22exemplar%22:true%7D%5D",
		}

		result.Items = append(result.Items, item)
		result.Annotations = append(result.Annotations, &schema.OpenCodeGraphAnnotation{
			Item:  schema.OpenCodeGraphItemRef{Id: id},
			Range: ranges[i],
		})
	}

	return &result, nil
}

var prometheusMetricDef = regexp.MustCompile(`(?ms)prometheus\.(?:Gauge|Counter|Histogram)Opts\{[^}]+\s+Name:\s*"([^"]+)"`)

func prometheusMetricDefs(content string) (gauges []string, ranges []schema.OpenCodeGraphRange) {
	return firstSubmatchNamesAndRanges(prometheusMetricDef, content)
}
