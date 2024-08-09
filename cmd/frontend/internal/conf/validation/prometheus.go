package validation

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	contributeWarning(newPrometheusValidator(srcprometheus.NewClient(srcprometheus.PrometheusURL)))
}

// newPrometheusValidator renders problems with the Prometheus deployment and relevant site configuration
// as reported by `prom-wrapper` inside the `sourcegraph/prometheus` container if Prometheus is enabled.
//
// It also accepts the error from creating `srcprometheus.Client` as an parameter, to validate
// Prometheus configuration.
func newPrometheusValidator(prom srcprometheus.Client, promErr error) conf.Validator {
	return func(c conftypes.SiteConfigQuerier) conf.Problems {
		// surface new prometheus client error if it was unexpected
		prometheusUnavailable := errors.Is(promErr, srcprometheus.ErrPrometheusUnavailable)
		if promErr != nil && !prometheusUnavailable {
			return conf.NewSiteProblems(fmt.Sprintf("Prometheus (`PROMETHEUS_URL`) might be misconfigured: %v", promErr))
		}

		// no need to validate prometheus config if no `observability.*` settings are configured
		observabilityNotConfigured := len(c.SiteConfig().ObservabilityAlerts) == 0 && len(c.SiteConfig().ObservabilitySilenceAlerts) == 0
		if observabilityNotConfigured {
			// no observability configuration, no checks to make
			return nil
		} else if prometheusUnavailable {
			// no prometheus, but observability is configured
			return conf.NewSiteProblems("`observability.alerts` or `observability.silenceAlerts` are configured, but Prometheus is not available")
		}

		// use a short timeout to avoid having this block problems from loading
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		// get reported problems
		status, err := prom.GetConfigStatus(ctx)
		if err != nil {
			return conf.NewSiteProblems(fmt.Sprintf("`observability`: failed to fetch alerting configuration status: %v", err))
		}
		return status.Problems
	}
}
