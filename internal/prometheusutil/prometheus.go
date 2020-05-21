package prometheusutil

import (
	"context"

	"github.com/pkg/errors"
	prometheusAPI "github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var PrometheusURL = env.Get("PROMETHEUS_URL", "", "prometheus server URL")

// PrometheusQuerier provides a shim around prometheus.API
type PrometheusQuerier interface {
	// QueryRange performs a query for the given range.
	QueryRange(ctx context.Context, query string, r prometheus.Range) (model.Value, prometheus.Warnings, error)
}

// ErrPrometheusUnavailable is raised specifically when prometheusURL is unset or when
// prometheus API access times out, both of which indicate that the server API has likely
// been configured to explicitly disallow access to prometheus, or that prometheus is not
// deployed at all. The website checks for this error in `fetchMonitoringStats`, for example.
var ErrPrometheusUnavailable = errors.New("prometheus API is unavailable")

func NewPrometheusQuerier() (PrometheusQuerier, error) {
	if PrometheusURL == "" {
		return nil, ErrPrometheusUnavailable
	}
	c, err := prometheusAPI.NewClient(prometheusAPI.Config{
		Address:      PrometheusURL,
		RoundTripper: &roundTripper{},
	})
	if err != nil {
		return nil, errors.Wrap(err, "prometheus configuration malformed")
	}
	return prometheus.NewAPI(c), nil
}
