package trace

import (
	"strings"
	"sync"
	"text/template"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

var (
	cachedURLTemplateStr string
	cachedURLTemplate    *template.Template
	cachedURLTemplateErr error
	cachedURLTemplateMu  sync.Mutex
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

	cachedURLTemplateMu.Lock()
	defer cachedURLTemplateMu.Unlock()

	if cachedURLTemplateStr != tracing.UrlTemplate {
		cachedURLTemplateStr = tracing.UrlTemplate
		cachedURLTemplate, cachedURLTemplateErr = template.New("traceURL").Parse(tracing.UrlTemplate)
	}

	if cachedURLTemplateErr != nil {
		// We contribute a validator on tracer package init, so safe to no-op here
		return ""
	}

	var sb strings.Builder
	_ = cachedURLTemplate.Execute(&sb, map[string]string{
		"TraceID":     traceID,
		"ExternalURL": c.ExternalURL,
	})
	return sb.String()
}
