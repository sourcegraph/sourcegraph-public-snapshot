package graphqlbackend

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/ext"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sourcegraph/sourcegraph/internal/prometheusutil"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// MonitoringAlert implements the GraphQL type MonitoringAlert.
type MonitoringAlert struct {
	TimestampValue   DateTime
	NameValue        string
	ServiceNameValue string
	OccurrencesValue int32
}

func (r *MonitoringAlert) Timestamp() DateTime { return r.TimestampValue }
func (r *MonitoringAlert) Name() string        { return r.NameValue }
func (r *MonitoringAlert) ServiceName() string { return r.ServiceNameValue }
func (r *MonitoringAlert) Occurrences() int32  { return r.OccurrencesValue }

func (r *siteResolver) MonitoringStatistics(ctx context.Context, args *struct {
	Days *int32
}) (*siteMonitoringStatisticsResolver, error) {
	prom, err := prometheusutil.NewPrometheusQuerier()
	if err != nil {
		return nil, err
	}
	return &siteMonitoringStatisticsResolver{
		prom:     prom,
		timespan: time.Duration(*args.Days) * 24 * time.Hour,
	}, nil
}

type siteMonitoringStatisticsResolver struct {
	prom     prometheusutil.PrometheusQuerier
	timespan time.Duration
}

const alertsResolution = 12 * time.Hour

func (r *siteMonitoringStatisticsResolver) Alerts(ctx context.Context) ([]*MonitoringAlert, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	span, ctx := ot.StartSpanFromContext(ctx, "site.MonitoringStatistics.alerts")

	var err error
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		cancel()
		span.Finish()
	}()

	results, warn, err := r.prom.QueryRange(ctx, `sum by (service_name,name)(alert_count{name!=""})`,
		prometheus.Range{
			Start: time.Now().Add(-r.timespan),
			End:   time.Now(),
			Step:  alertsResolution,
		})
	if errors.Is(err, context.Canceled) {
		return nil, prometheusutil.ErrPrometheusUnavailable
	}
	if err != nil {
		return nil, errors.Wrap(err, "prometheus query failed")
	}
	if len(warn) > 0 {
		log15.Warn("site.monitoring.alerts: warnings encountered on prometheus query",
			"timespan", r.timespan.String(),
			"warnings", warn)
	}
	if results.Type() != model.ValMatrix {
		return nil, fmt.Errorf("received unexpected result type %q from prometheus", results.Type())
	}

	data := results.(model.Matrix)
	alerts := make([]*MonitoringAlert, 0)
	for _, sample := range data {
		var (
			name        = string(sample.Metric["name"])
			serviceName = string(sample.Metric["service_name"])
			prevVal     *model.SampleValue
		)
		for _, p := range sample.Values {
			// Check for nil so that we don't ignore the first occurrence of an alert - even if the
			// alert is never >0, we want to be aware that it is at least configured correctly and
			// being tracked. Otherwise, if the value in this window is the same as in the previous
			// window just discard it.
			if prevVal != nil && p.Value == *prevVal {
				continue
			}
			// copy value for comparison later
			v := p.Value
			prevVal = &v
			// record alert in results
			alerts = append(alerts, &MonitoringAlert{
				NameValue:        name,
				ServiceNameValue: serviceName,
				TimestampValue:   DateTime{p.Timestamp.Time().UTC().Truncate(time.Hour)},
				OccurrencesValue: int32(p.Value),
			})
		}
	}
	return alerts, err
}
