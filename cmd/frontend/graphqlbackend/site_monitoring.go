package graphqlbackend

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
)

type MonitoringAlert struct{}

func (r *MonitoringAlert) Timestamp() gqlutil.DateTime {
	return gqlutil.DateTime{Time: time.Time{}}
}
func (r *MonitoringAlert) Name() string        { return "" }
func (r *MonitoringAlert) ServiceName() string { return "" }
func (r *MonitoringAlert) Owner() string       { return "" }
func (r *MonitoringAlert) Average() float64    { return 0 }

func (r *siteResolver) MonitoringStatistics(ctx context.Context, args *struct {
	Days *int32
}) (*siteMonitoringStatisticsResolver, error) {
	prom, err := srcprometheus.NewClient(srcprometheus.PrometheusURL)
	if err != nil {
		return nil, err // clients should check for ErrPrometheusUnavailable
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
	return []*MonitoringAlert{}, nil
}
