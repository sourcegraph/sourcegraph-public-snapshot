package metrics

import (
	"context"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/cockroachdb/errors"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/inconshreveable/log15"
	"google.golang.org/api/option"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func initGCPMetrics(config *GCPConfig, queueOptions map[string]handler.QueueOptions) error {
	if config.ProjectID == "" {
		return nil
	}

	metricClient, err := monitoring.NewMetricClient(context.Background(), gcsClientOptions(config)...)
	if err != nil {
		return err
	}

	for queueName, options := range queueOptions {
		initGCPMetric(config, metricClient, queueName, options.Store)
	}

	return nil
}

func initGCPMetric(config *GCPConfig, metricClient *monitoring.MetricClient, queueName string, store store.Store) {
	go func() {
		for {
			if err := sendGCPMetric(config, metricClient, queueName, store); err != nil {
				log15.Error("Failed to send executor queue size metric to GCP", "queue", queueName, "error", err)
			}

			time.Sleep(time.Second * 5)
		}
	}()
}

func sendGCPMetric(config *GCPConfig, metricClient *monitoring.MetricClient, queueName string, store store.Store) error {
	count, err := store.QueuedCount(context.Background(), true, nil)
	if err != nil {
		return errors.Wrap(err, "dbworkerstore.QueuedCount")
	}

	if err := metricClient.CreateTimeSeries(context.Background(), makeGCPMetricRequest(config, queueName, count)); err != nil {
		return errors.Wrap(err, "metricClient.CreateTimeSeries")
	}

	return nil
}

const (
	gcpMetricKind = metricpb.MetricDescriptor_GAUGE
	gcpMetricType = "custom.googleapis.com/executors/queue/size"
)

func makeGCPMetricRequest(config *GCPConfig, queueName string, count int) *monitoringpb.CreateTimeSeriesRequest {
	pbMetric := &metricpb.Metric{Type: gcpMetricType, Labels: map[string]string{"queueName": queueName, "environment": config.EnvironmentLabel}}
	now := &timestamp.Timestamp{Seconds: timeutil.Now().Unix()}
	pbInterval := &monitoringpb.TimeInterval{StartTime: now, EndTime: now}
	pbValue := &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_Int64Value{Int64Value: int64(count)}}
	pbTimeSeriesPoints := []*monitoringpb.Point{{Interval: pbInterval, Value: pbValue}}
	pbTimeSeries := &monitoringpb.TimeSeries{
		Metric:     pbMetric,
		MetricKind: gcpMetricKind,
		Points:     pbTimeSeriesPoints,
		Resource:   &monitoredres.MonitoredResource{Type: "global", Labels: map[string]string{"project_id": config.ProjectID}},
	}

	return &monitoringpb.CreateTimeSeriesRequest{
		Name:       fmt.Sprintf("projects/%s", config.ProjectID),
		TimeSeries: []*monitoringpb.TimeSeries{pbTimeSeries},
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
