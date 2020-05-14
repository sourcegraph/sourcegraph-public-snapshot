package graphqlbackend

import (
	"context"
	"fmt"
	"time"

	prometheusAPI "github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
)

// MonitoringAlert implements the GraphQL type MonitoringAlert.
type MonitoringAlert struct {
	TimestampValue   DateTime
	NameValue        string
	ServiceNameValue string
	ValueValue       int32
}

func (r *MonitoringAlert) Timestamp() DateTime { return r.TimestampValue }
func (r *MonitoringAlert) Name() string        { return r.NameValue }
func (r *MonitoringAlert) ServiceName() string { return r.ServiceNameValue }
func (r *MonitoringAlert) Value() int32        { return r.ValueValue }

func (r *siteResolver) MonitoringStatistics(ctx context.Context, args *struct {
	Days *int32
}) (*siteMonitoringStatisticsResolver, error) {
	c, err := prometheusAPI.NewClient(prometheusAPI.Config{})
	if err != nil {
		return nil, err
	}
	return &siteMonitoringStatisticsResolver{
		ctx:      ctx,
		prom:     prometheus.NewAPI(c),
		timespan: time.Duration(*args.Days) * 24 * time.Hour,
	}, nil
}

type siteMonitoringStatisticsResolver struct {
	ctx      context.Context
	prom     prometheus.API
	timespan time.Duration
}

func (r *siteMonitoringStatisticsResolver) Alerts() ([]*MonitoringAlert, error) {
	results, warn, err := r.prom.Query(r.ctx, "sum by (service_name,name)(alert_count{name!=\"\"})", time.Now().Add(r.timespan))
	if err != nil {
		fmt.Printf("ROBERT %+v\n", err)
		return nil, err
	}
	if len(warn) > 0 {
		fmt.Printf("ROBERT %+v\n", warn)
		return nil, err
	}
	fmt.Printf("ROBERT %+v\n", results)
	return nil, err
}
