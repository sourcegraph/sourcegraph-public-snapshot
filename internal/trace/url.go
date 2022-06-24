package trace

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

// URL returns a trace URL for the given trace ID at the given external URL.
func URL(traceID, externalURL, traceProvider string) string {
	if traceID == "" {
		return ""
	}
	if tracer.TracerType(traceProvider) == tracer.Datadog {
		return "https://app.datadoghq.com/apm/trace/" + traceID
	}

	if os.Getenv("ENABLE_GRAFANA_CLOUD_TRACE_URL") != "true" {
		// We proxy jaeger so we can construct URLs to traces.
		return strings.TrimSuffix(externalURL, "/") + "/-/debug/jaeger/trace/" + traceID
	}

	return "https://sourcegraph.grafana.net/explore?orgId=1&left=" + url.QueryEscape(fmt.Sprintf(
		`["now-1h","now","grafanacloud-sourcegraph-traces",{"query":"%s","queryType":"traceId"}]`,
		traceID,
	))
}
