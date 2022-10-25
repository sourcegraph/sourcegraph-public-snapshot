package executorqueue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type awsConfig struct {
	MetricNamespace string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func (c *awsConfig) load(parent *env.BaseConfig) {
	c.MetricNamespace = parent.Get("EXECUTOR_METRIC_AWS_NAMESPACE", "", "The namespace to which to export the custom metric for scaling executors.")
	c.Region = parent.Get("EXECUTOR_METRIC_AWS_REGION", "", "The target AWS region.")
	c.AccessKeyID = parent.Get("EXECUTOR_METRIC_AWS_ACCESS_KEY_ID", "", "An AWS access key associated with a user with access to CloudWatch.")
	c.SecretAccessKey = parent.Get("EXECUTOR_METRIC_AWS_SECRET_ACCESS_KEY", "", "An AWS secret key associated with a user with access to CloudWatch.")
	c.SessionToken = parent.GetOptional("EXECUTOR_METRIC_AWS_SESSION_TOKEN", "An optional AWS session token associated with a user with access to CloudWatch.")
}

func newAWSReporter(config *Config) (*awsMetricReporter, error) {
	if config.AWSConfig.MetricNamespace == "" || config.AWSConfig.Region == "" {
		return nil, nil
	}

	log15.Info("Sending executor queue metrics to AWS CloudWatch")

	cfg, err := awsClientOptions(context.Background(), config.AWSConfig)
	if err != nil {
		return nil, err
	}

	return &awsMetricReporter{
		config:           config.AWSConfig,
		environmentLabel: config.EnvironmentLabel,
		metricClient:     cloudwatch.NewFromConfig(cfg),
	}, nil
}

type awsMetricReporter struct {
	config           awsConfig
	environmentLabel string
	metricClient     *cloudwatch.Client
}

func (r *awsMetricReporter) ReportCount(ctx context.Context, queueName string, count int) {
	if _, err := r.metricClient.PutMetricData(ctx, makePutMetricDataInput(r.config.MetricNamespace, queueName, r.environmentLabel, count)); err != nil {
		log15.Error("Failed to send executor queue size metric to AWS CloudWatch", "queue", queueName, "error", err)
	}
}

func (r *awsMetricReporter) GetAllocation(queueAllocation QueueAllocation) float64 {
	return queueAllocation.PercentageAWS
}

const awsMetricName = "src_executors_queue_size"

func makePutMetricDataInput(namespace, queueName, environmentLabel string, count int) *cloudwatch.PutMetricDataInput {
	return &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String(awsMetricName),
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

func awsClientOptions(ctx context.Context, awsConfig awsConfig) (aws.Config, error) {
	optFns := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(awsConfig.Region),
	}

	if awsConfig.AccessKeyID != "" {
		optFns = append(optFns, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			awsConfig.AccessKeyID,
			awsConfig.SecretAccessKey,
			awsConfig.SessionToken,
		)))
	}

	return awsconfig.LoadDefaultConfig(ctx, optFns...)
}
