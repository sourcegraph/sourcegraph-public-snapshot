package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

	metricsconfig "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/metrics/config"
)

const metricName = "src_executors_queue_size"

func NewReporter(environmentLabel string) (*awsMetricReporter, error) {
	if awsConfig.MetricNamespace == "" {
		return nil, nil
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to load aws default config")
	}

	log15.Info("Sending executor queue metrics to AWS CloudWatch")

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

func (r *awsMetricReporter) ReportCount(ctx context.Context, queueName string, count int) {
	if _, err := r.metricClient.PutMetricData(ctx, makePutMetricDataInput(r.config, queueName, r.environmentLabel, count)); err != nil {
		log15.Error("Failed to send executor queue size metric to AWS CloudWatch", "queue", queueName, "error", err)
	}
}

func (r *awsMetricReporter) GetAllocation(queueAllocation metricsconfig.QueueAllocation) float64 {
	return queueAllocation.PercentageAWS
}

func makePutMetricDataInput(config *AWSConfig, queueName, environmentLabel string, count int) *cloudwatch.PutMetricDataInput {
	return &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(config.MetricNamespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String(metricName),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(float64(count)),
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("queueName"),
						Value: aws.String(queueName),
					},
					{
						Name:  aws.String("environment"),
						Value: aws.String(environmentLabel),
					},
				},
			},
		},
	}
}
