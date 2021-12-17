package executorqueue

import (
	"context"
	"fmt"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/inconshreveable/log15"
	"google.golang.org/api/option"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

type gcpConfig struct {
	env.BaseConfig

	ProjectID               string
	CredentialsFile         string
	CredentialsFileContents string
}

func (c *gcpConfig) load(parent *env.BaseConfig) {
	c.ProjectID = parent.Get("EXECUTOR_METRIC_GCP_PROJECT_ID", "", "The project containing the custom metric.")
	c.CredentialsFile = parent.GetOptional("EXECUTOR_METRIC_GOOGLE_APPLICATION_CREDENTIALS_FILE", "The path to a service account key file with access to metrics.")
	c.CredentialsFileContents = parent.GetOptional("EXECUTOR_METRIC_GOOGLE_APPLICATION_CREDENTIALS_FILE_CONTENT", "The contents of a service account key file with access to metrics.")
}

func newGCPReporter(config *Config) (*gcpMetricReporter, error) {
	if config.GCPConfig.ProjectID == "" {
		return nil, nil
	}

	log15.Info("Sending executor queue metrics to Google Cloud Monitoring")

	metricClient, err := monitoring.NewMetricClient(context.Background(), gcsClientOptions(config.GCPConfig)...)
	if err != nil {
		return nil, err
	}

	return &gcpMetricReporter{
		config:           config.GCPConfig,
		environmentLabel: config.EnvironmentLabel,
		metricClient:     metricClient,
	}, nil
}

type gcpMetricReporter struct {
	config           gcpConfig
	environmentLabel string
	metricClient     *monitoring.MetricClient
}

func (r *gcpMetricReporter) ReportCount(ctx context.Context, queueName string, count int) {
	if err := r.metricClient.CreateTimeSeries(ctx, makeCreateTimeSeriesRequest(r.config, queueName, r.environmentLabel, count)); err != nil {
		log15.Error("Failed to send executor queue size metric to Google Cloud Monitoring", "queue", queueName, "error", err)
	}
}

func (r *gcpMetricReporter) GetAllocation(queueAllocation QueueAllocation) float64 {
	return queueAllocation.PercentageGCP
}

const (
	gcpMetricKind = metricpb.MetricDescriptor_GAUGE
	gcpMetricType = "custom.googleapis.com/executors/queue/size"
)

func makeCreateTimeSeriesRequest(config gcpConfig, queueName, environmentLabel string, count int) *monitoringpb.CreateTimeSeriesRequest {
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
							StartTime: &timestamppb.Timestamp{Seconds: now},
							EndTime:   &timestamppb.Timestamp{Seconds: now},
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

func gcsClientOptions(config gcpConfig) []option.ClientOption {
	if config.CredentialsFile != "" {
		return []option.ClientOption{option.WithCredentialsFile(config.CredentialsFile)}
	}

	if config.CredentialsFileContents != "" {
		return []option.ClientOption{option.WithCredentialsJSON([]byte(config.CredentialsFileContents))}
	}

	return nil
}
