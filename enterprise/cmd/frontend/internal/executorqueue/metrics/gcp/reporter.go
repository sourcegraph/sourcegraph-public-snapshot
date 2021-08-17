package gcp

import (
	"context"
	"fmt"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/inconshreveable/log15"
	"google.golang.org/api/option"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	metricsconfig "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/metrics/config"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

const (
	gcpMetricKind = metricpb.MetricDescriptor_GAUGE
	gcpMetricType = "custom.googleapis.com/executors/queue/size"
)

func NewReporter(environmentLabel string) (*gcpMetricReporter, error) {
	if gcpConfig.ProjectID == "" {
		return nil, nil
	}

	log15.Info("Sending executor queue metrics to Google Cloud Monitoring")

	metricClient, err := monitoring.NewMetricClient(context.Background(), gcsClientOptions(gcpConfig)...)
	if err != nil {
		return nil, err
	}

	return &gcpMetricReporter{
		config:           gcpConfig,
		environmentLabel: environmentLabel,
		metricClient:     metricClient,
	}, nil
}

type gcpMetricReporter struct {
	config           *GCPConfig
	environmentLabel string
	metricClient     *monitoring.MetricClient
}

func (r *gcpMetricReporter) ReportCount(ctx context.Context, queueName string, count int) {
	if err := r.metricClient.CreateTimeSeries(ctx, makeCreateTimeSeriesRequest(r.config, queueName, r.environmentLabel, count)); err != nil {
		log15.Error("Failed to send executor queue size metric to Google Cloud Monitoring", "queue", queueName, "error", err)
	}
}

func (r *gcpMetricReporter) GetAllocation(queueAllocation metricsconfig.QueueAllocation) float64 {
	return queueAllocation.PercentageGCP
}

func makeCreateTimeSeriesRequest(config *GCPConfig, queueName, environmentLabel string, count int) *monitoringpb.CreateTimeSeriesRequest {
	name := fmt.Sprintf("projects/%s", config.ProjectID)
	now := timeutil.Now().Unix()

	return &monitoringpb.CreateTimeSeriesRequest{
		Name: name,
		TimeSeries: []*monitoringpb.TimeSeries{
			{
				MetricKind: gcpMetricKind,
				Metric: &metricpb.Metric{
					Type: gcpMetricType,
					Labels: map[string]string{
						"queueName":   queueName,
						"environment": environmentLabel,
					},
				},
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							StartTime: &timestamp.Timestamp{Seconds: now},
							EndTime:   &timestamp.Timestamp{Seconds: now},
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{Int64Value: int64(count)},
						},
					},
				},
				Resource: &monitoredres.MonitoredResource{
					Type: "global",
					Labels: map[string]string{
						"project_id": config.ProjectID,
					},
				},
			},
		},
	}
}

func gcsClientOptions(config *GCPConfig) []option.ClientOption {
	if config.CredentialsFile != "" {
		return []option.ClientOption{option.WithCredentialsFile(config.CredentialsFile)}
	}

	if config.CredentialsFileContents != "" {
		return []option.ClientOption{option.WithCredentialsJSON([]byte(config.CredentialsFileContents))}
	}

	return nil
}
