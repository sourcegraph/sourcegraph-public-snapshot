package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

const (
	metricName = "src_executors_queue_size"
)

func InitReportCounter(environmentLabel string) (*awsMetricReporter, error) {
	if awsConfig.MetricNamespace == "" {
		return nil, nil
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to load aws default config")
	}

	metricClient := cloudwatch.NewFromConfig(cfg)

	return &awsMetricReporter{
		config:           awsConfig,
		environmentLabel: environmentLabel,
		metricClient:     metricClient,
	}, nil
}

type awsMetricReporter struct {
	config           *AWSConfig
	environmentLabel string
	metricClient     *cloudwatch.Client
}

func (r *awsMetricReporter) ReportCount(ctx context.Context, queueName string, store store.Store, count int) {
	if err := sendAWSMetric(r.config, r.metricClient, queueName, r.environmentLabel, store, count); err != nil {
		log15.Error("Failed to send executor queue size metric to GCP", "queue", queueName, "error", err)
	}
}

func sendAWSMetric(c *AWSConfig, metricClient *cloudwatch.Client, queueName, environmentLabel string, store store.Store, count int) error {
	if _, err := metricClient.PutMetricData(context.Background(), makeAWSMetricRequest(c, queueName, environmentLabel, count)); err != nil {
		return errors.Wrap(err, "metricClient.PutMetricData")
	}

	return nil
}

func makeAWSMetricRequest(c *AWSConfig, queueName, environmentLabel string, count int) *cloudwatch.PutMetricDataInput {
	input := &cloudwatch.PutMetricDataInput{
		Namespace: strptr(c.MetricNamespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: strptr(metricName),
				Unit:       types.StandardUnitCount,
				Value:      f64ptr(float64(count)),
				Dimensions: []types.Dimension{
					{
						Name:  strptr("environment"),
						Value: strptr(environmentLabel),
					},
					{
						Name:  strptr("queueName"),
						Value: strptr(queueName),
					},
				},
			},
		},
	}
	return input
}

func strptr(s string) *string { return &s }

func f64ptr(f float64) *float64 { return &f }
