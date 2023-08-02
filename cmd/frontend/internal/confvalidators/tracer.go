package confvalidators

import (
	"fmt"
	"html/template"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

// Contribute validation for tracing package
func validateTracerConfig(c conftypes.SiteConfigQuerier) conf.Problems {
	tracing := c.SiteConfig().ObservabilityTracing
	if tracing == nil || tracing.UrlTemplate == "" {
		return nil
	}
	if _, err := template.New("").Parse(tracing.UrlTemplate); err != nil {
		return conf.NewSiteProblems(fmt.Sprintf("observability.tracing.traceURL is not a valid template: %s", err.Error()))
	}
	return nil
}
