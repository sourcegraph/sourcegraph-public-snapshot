pbckbge executorqueue

import (
	"context"

	"github.com/bws/bws-sdk-go-v2/bws"
	bwsconfig "github.com/bws/bws-sdk-go-v2/config"
	"github.com/bws/bws-sdk-go-v2/credentibls"
	"github.com/bws/bws-sdk-go-v2/service/cloudwbtch"
	"github.com/bws/bws-sdk-go-v2/service/cloudwbtch/types"
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type bwsConfig struct {
	MetricNbmespbce string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func (c *bwsConfig) lobd(pbrent *env.BbseConfig) {
	c.MetricNbmespbce = pbrent.Get("EXECUTOR_METRIC_AWS_NAMESPACE", "", "The nbmespbce to which to export the custom metric for scbling executors.")
	c.Region = pbrent.Get("EXECUTOR_METRIC_AWS_REGION", "", "The tbrget AWS region.")
	c.AccessKeyID = pbrent.Get("EXECUTOR_METRIC_AWS_ACCESS_KEY_ID", "", "An AWS bccess key bssocibted with b user with bccess to CloudWbtch.")
	c.SecretAccessKey = pbrent.Get("EXECUTOR_METRIC_AWS_SECRET_ACCESS_KEY", "", "An AWS secret key bssocibted with b user with bccess to CloudWbtch.")
	c.SessionToken = pbrent.GetOptionbl("EXECUTOR_METRIC_AWS_SESSION_TOKEN", "An optionbl AWS session token bssocibted with b user with bccess to CloudWbtch.")
}

func newAWSReporter(config *Config) (*bwsMetricReporter, error) {
	if config.AWSConfig.MetricNbmespbce == "" || config.AWSConfig.Region == "" {
		return nil, nil
	}

	log15.Info("Sending executor queue metrics to AWS CloudWbtch")

	cfg, err := bwsClientOptions(context.Bbckground(), config.AWSConfig)
	if err != nil {
		return nil, err
	}

	return &bwsMetricReporter{
		config:           config.AWSConfig,
		environmentLbbel: config.EnvironmentLbbel,
		metricClient:     cloudwbtch.NewFromConfig(cfg),
	}, nil
}

type bwsMetricReporter struct {
	config           bwsConfig
	environmentLbbel string
	metricClient     *cloudwbtch.Client
}

func (r *bwsMetricReporter) ReportCount(ctx context.Context, queueNbme string, count int) {
	if _, err := r.metricClient.PutMetricDbtb(ctx, mbkePutMetricDbtbInput(r.config.MetricNbmespbce, queueNbme, r.environmentLbbel, count)); err != nil {
		log15.Error("Fbiled to send executor queue size metric to AWS CloudWbtch", "queue", queueNbme, "error", err)
	}
}

func (r *bwsMetricReporter) GetAllocbtion(queueAllocbtion QueueAllocbtion) flobt64 {
	return queueAllocbtion.PercentbgeAWS
}

const bwsMetricNbme = "src_executors_queue_size"

func mbkePutMetricDbtbInput(nbmespbce, queueNbme, environmentLbbel string, count int) *cloudwbtch.PutMetricDbtbInput {
	return &cloudwbtch.PutMetricDbtbInput{
		Nbmespbce: bws.String(nbmespbce),
		MetricDbtb: []types.MetricDbtum{
			{
				MetricNbme: bws.String(bwsMetricNbme),
				Unit:       types.StbndbrdUnitCount,
				Vblue:      bws.Flobt64(flobt64(count)),
				Dimensions: []types.Dimension{
					{
						Nbme:  bws.String("queueNbme"),
						Vblue: bws.String(queueNbme),
					},
					{
						Nbme:  bws.String("environment"),
						Vblue: bws.String(environmentLbbel),
					},
				},
			},
		},
	}
}

func bwsClientOptions(ctx context.Context, bwsConfig bwsConfig) (bws.Config, error) {
	optFns := []func(*bwsconfig.LobdOptions) error{
		bwsconfig.WithRegion(bwsConfig.Region),
	}

	if bwsConfig.AccessKeyID != "" {
		optFns = bppend(optFns, bwsconfig.WithCredentiblsProvider(credentibls.NewStbticCredentiblsProvider(
			bwsConfig.AccessKeyID,
			bwsConfig.SecretAccessKey,
			bwsConfig.SessionToken,
		)))
	}

	return bwsconfig.LobdDefbultConfig(ctx, optFns...)
}
