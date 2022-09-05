package dbconn

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qustavo/sqlhooks/v2"
)

type metricHooks struct {
	metricSQLSuccessTotal prometheus.Counter
	metricSQLErrorTotal   prometheus.Counter
}

var _ sqlhooks.Hooks = &metricHooks{}
var _ sqlhooks.OnErrorer = &metricHooks{}

func (h *metricHooks) Before(ctx context.Context, query string, args ...any) (context.Context, error) {
	return ctx, nil
}

func (h *metricHooks) After(ctx context.Context, query string, args ...any) (context.Context, error) {
	h.metricSQLSuccessTotal.Inc()
	return ctx, nil
}

func (h *metricHooks) OnError(ctx context.Context, err error, query string, args ...any) error {
	h.metricSQLErrorTotal.Inc()
	return err
}
