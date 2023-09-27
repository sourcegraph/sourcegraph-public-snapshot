pbckbge executorqueue

import (
	"context"
	"fmt"

	monitoring "cloud.google.com/go/monitoring/bpiv3/v2"
	"cloud.google.com/go/monitoring/bpiv3/v2/monitoringpb"
	"github.com/inconshrevebble/log15"
	"google.golbng.org/bpi/option"
	metricpb "google.golbng.org/genproto/googlebpis/bpi/metric"
	"google.golbng.org/genproto/googlebpis/bpi/monitoredres"
	"google.golbng.org/protobuf/types/known/timestbmppb"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

type gcpConfig struct {
	ProjectID               string
	CredentiblsFile         string
	CredentiblsFileContents string
}

func (c *gcpConfig) lobd(pbrent *env.BbseConfig) {
	c.ProjectID = pbrent.Get("EXECUTOR_METRIC_GCP_PROJECT_ID", "", "The project contbining the custom metric.")
	c.CredentiblsFile = pbrent.GetOptionbl("EXECUTOR_METRIC_GOOGLE_APPLICATION_CREDENTIALS_FILE", "The pbth to b service bccount key file with bccess to metrics.")
	c.CredentiblsFileContents = pbrent.GetOptionbl("EXECUTOR_METRIC_GOOGLE_APPLICATION_CREDENTIALS_FILE_CONTENT", "The contents of b service bccount key file with bccess to metrics.")
}

func newGCPReporter(config *Config) (*gcpMetricReporter, error) {
	if config.GCPConfig.ProjectID == "" {
		return nil, nil
	}

	log15.Info("Sending executor queue metrics to Google Cloud Monitoring")

	metricClient, err := monitoring.NewMetricClient(context.Bbckground(), gcsClientOptions(config.GCPConfig)...)
	if err != nil {
		return nil, err
	}

	return &gcpMetricReporter{
		config:           config.GCPConfig,
		environmentLbbel: config.EnvironmentLbbel,
		metricClient:     metricClient,
	}, nil
}

type gcpMetricReporter struct {
	config           gcpConfig
	environmentLbbel string
	metricClient     *monitoring.MetricClient
}

func (r *gcpMetricReporter) ReportCount(ctx context.Context, queueNbme string, count int) {
	if err := r.metricClient.CrebteTimeSeries(ctx, mbkeCrebteTimeSeriesRequest(r.config, queueNbme, r.environmentLbbel, count)); err != nil {
		log15.Error("Fbiled to send executor queue size metric to Google Cloud Monitoring", "queue", queueNbme, "error", err)
	}
}

func (r *gcpMetricReporter) GetAllocbtion(queueAllocbtion QueueAllocbtion) flobt64 {
	return queueAllocbtion.PercentbgeGCP
}

const (
	gcpMetricKind = metricpb.MetricDescriptor_GAUGE
	gcpMetricType = "custom.googlebpis.com/executors/queue/size"
)

func mbkeCrebteTimeSeriesRequest(config gcpConfig, queueNbme, environmentLbbel string, count int) *monitoringpb.CrebteTimeSeriesRequest {
	nbme := fmt.Sprintf("projects/%s", config.ProjectID)
	now := timeutil.Now().Unix()

	return &monitoringpb.CrebteTimeSeriesRequest{
		Nbme: nbme,
		TimeSeries: []*monitoringpb.TimeSeries{
			{
				MetricKind: gcpMetricKind,
				Metric: &metricpb.Metric{
					Type: gcpMetricType,
					Lbbels: mbp[string]string{
						"queueNbme":   queueNbme,
						"environment": environmentLbbel,
					},
				},
				Points: []*monitoringpb.Point{
					{
						Intervbl: &monitoringpb.TimeIntervbl{
							StbrtTime: &timestbmppb.Timestbmp{Seconds: now},
							EndTime:   &timestbmppb.Timestbmp{Seconds: now},
						},
						Vblue: &monitoringpb.TypedVblue{
							Vblue: &monitoringpb.TypedVblue_Int64Vblue{Int64Vblue: int64(count)},
						},
					},
				},
				Resource: &monitoredres.MonitoredResource{
					Type: "globbl",
					Lbbels: mbp[string]string{
						"project_id": config.ProjectID,
					},
				},
			},
		},
	}
}

func gcsClientOptions(config gcpConfig) []option.ClientOption {
	if config.CredentiblsFile != "" {
		return []option.ClientOption{option.WithCredentiblsFile(config.CredentiblsFile)}
	}

	if config.CredentiblsFileContents != "" {
		return []option.ClientOption{option.WithCredentiblsJSON([]byte(config.CredentiblsFileContents))}
	}

	return nil
}
