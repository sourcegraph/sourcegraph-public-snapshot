package trace

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

// URL returns a trace URL for the given trace ID at the given external URL.
func URL(traceID string, querier conftypes.SiteConfigQuerier) string {
	if traceID == "" {
		return ""
	}
	c := querier.SiteConfig()
	tracing := c.ObservabilityTracing
	if tracing == nil || tracing.UrlTemplate == "" {
		return legacyTraceURL(traceID, c.ExternalURL)
	}

	tpl, err := template.New("traceURL").Parse(tracing.UrlTemplate)
	if err != nil {
		// We contribute a validator on tracer package init, so safe to no-op here
		return ""
	}

	var sb strings.Builder
	_ = tpl.Execute(&sb, map[string]string{
		"TraceID":     traceID,
		"ExternalURL": c.ExternalURL,
	})
	return sb.String()
}

// legacyTraceURL preserves the old trace URL generation behaviour if no template is
// provided.
func legacyTraceURL(traceID, externalURL string) string {
	if os.Getenv("ENABLE_GRAFANA_CLOUD_TRACE_URL") != "true" {
		// We proxy jaeger so we can construct URLs to traces.
		return strings.TrimSuffix(externalURL, "/") + "/-/debug/jaeger/trace/" + traceID
	}

	return "https://sourcegraph.grafana.net/explore?orgId=1&left=" + url.QueryEscape(fmt.Sprintf(
		`["now-1h","now","grafanacloud-sourcegraph-traces",{"query":"%s","queryType":"traceId"}]`,
		traceID,
	))
}
