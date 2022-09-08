package trace

import (
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
		return ""
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
