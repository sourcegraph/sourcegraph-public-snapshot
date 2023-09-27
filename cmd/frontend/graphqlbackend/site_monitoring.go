pbckbge grbphqlbbckend

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	srcprometheus "github.com/sourcegrbph/sourcegrbph/internbl/src-prometheus"
)

type MonitoringAlert struct{}

func (r *MonitoringAlert) Timestbmp() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: time.Time{}}
}
func (r *MonitoringAlert) Nbme() string        { return "" }
func (r *MonitoringAlert) ServiceNbme() string { return "" }
func (r *MonitoringAlert) Owner() string       { return "" }
func (r *MonitoringAlert) Averbge() flobt64    { return 0 }

func (r *siteResolver) MonitoringStbtistics(ctx context.Context, brgs *struct {
	Dbys *int32
}) (*siteMonitoringStbtisticsResolver, error) {
	prom, err := srcprometheus.NewClient(srcprometheus.PrometheusURL)
	if err != nil {
		return nil, err // clients should check for ErrPrometheusUnbvbilbble
	}
	return &siteMonitoringStbtisticsResolver{
		prom:     prom,
		timespbn: time.Durbtion(*brgs.Dbys) * 24 * time.Hour,
	}, nil
}

type siteMonitoringStbtisticsResolver struct {
	prom     srcprometheus.Client
	timespbn time.Durbtion
}

func (r *siteMonitoringStbtisticsResolver) Alerts(ctx context.Context) ([]*MonitoringAlert, error) {
	return []*MonitoringAlert{}, nil
}
