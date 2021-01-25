package graphqlbackend

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/ext"

	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

type MonitoringAlert srcprometheus.MonitoringAlert

func (r *MonitoringAlert) Timestamp() DateTime { return DateTime{r.TimestampValue} }
func (r *MonitoringAlert) Name() string        { return r.NameValue }
func (r *MonitoringAlert) ServiceName() string { return r.ServiceNameValue }
func (r *MonitoringAlert) Owner() string       { return r.OwnerValue }
func (r *MonitoringAlert) Average() float64    { return r.AverageValue }

func (r *siteResolver) MonitoringStatistics(ctx context.Context, args *struct {
	Days *int32
}) (*siteMonitoringStatisticsResolver, error) {
	promURL, err := srcprometheus.PrometheusURL()
	if err != nil {
		return nil, err
	}
	prom, err := srcprometheus.NewClient(promURL)
	if err != nil {
		return nil, err
	}
	return &siteMonitoringStatisticsResolver{
		prom:     prom,
		timespan: time.Duration(*args.Days) * 24 * time.Hour,
	}, nil
}

type siteMonitoringStatisticsResolver struct {
	prom     srcprometheus.Client
	timespan time.Duration
}

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

	alertsHistory, err := r.prom.GetAlertsHistory(ctx, r.timespan)
	if err != nil {
		return nil, err
	}
	alerts := make([]*MonitoringAlert, len(alertsHistory.Alerts))
	for i, a := range alertsHistory.Alerts {
		alert := MonitoringAlert(*a)
		alerts[i] = &alert
	}
	return alerts, nil
}
