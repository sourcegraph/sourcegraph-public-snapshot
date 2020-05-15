package graphqlbackend

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/ext"
	prometheusAPI "github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// MonitoringAlert implements the GraphQL type MonitoringAlert.
type MonitoringAlert struct {
	TimestampValue   DateTime
	NameValue        string
	ServiceNameValue string
}

func (r *MonitoringAlert) Timestamp() DateTime { return r.TimestampValue }
func (r *MonitoringAlert) Name() string        { return r.NameValue }
func (r *MonitoringAlert) ServiceName() string { return r.ServiceNameValue }

func (r *siteResolver) MonitoringStatistics(ctx context.Context, args *struct {
	Days *int32
}) (*siteMonitoringStatisticsResolver, error) {
	c, err := prometheusAPI.NewClient(prometheusAPI.Config{
		Address: prometheusURL,
	})
	if err != nil {
		return nil, fmt.Errorf("prometheus unavailable: %w", err)
	}
	return &siteMonitoringStatisticsResolver{
		ctx:      ctx,
		prom:     prometheus.NewAPI(c),
		timespan: time.Duration(*args.Days) * 24 * time.Hour,
	}, nil
}

type siteMonitoringStatisticsResolver struct {
	ctx      context.Context
	prom     prometheusQuerier
	timespan time.Duration
}

func (r *siteMonitoringStatisticsResolver) Alerts() ([]*MonitoringAlert, error) {
	var err error
	span, ctx := ot.StartSpanFromContext(r.ctx, "site.MonitoringStatistics.alerts")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	results, warn, err := r.prom.QueryRange(ctx, "sum by (service_name,name)(alert_count{name!=\"\"})",
		prometheus.Range{
			Start: time.Now().Add(-r.timespan),
			End:   time.Now(),
			Step:  24 * time.Hour,
		})
	if err != nil {
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}
	if len(warn) > 0 {
		log.Printf("site.monitoring.alerts: warnings encountered on prometheus query (%s): [ %s ]",
			r.timespan.String(), strings.Join(warn, ","))
	}
	if results.Type() != model.ValMatrix {
		return nil, fmt.Errorf("received unexpected result type '%s' from query", results.Type())
	}

	data := results.(model.Matrix)
	alerts := make([]*MonitoringAlert, 0)
	for _, sample := range data {
		var (
			name        = string(sample.Metric["name"])
			serviceName = string(sample.Metric["service_name"])
		)
		for _, p := range sample.Values {
			// skip values that indicate no occurences of this alert
			if p.Value.String() == "0" {
				continue
			}

			alerts = append(alerts, &MonitoringAlert{
				NameValue:        name,
				ServiceNameValue: serviceName,
				TimestampValue:   DateTime{p.Timestamp.Time()},
			})
		}
	}
	return alerts, err
}
