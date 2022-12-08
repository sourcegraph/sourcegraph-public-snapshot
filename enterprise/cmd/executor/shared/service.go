package shared

import (
	"context"
	"os"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "executor" }

func (svc) Configure() env.Config {
	var config config.Config
	config.Load()
	return &config
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, cfg env.Config) error {
	config := cfg.(*config.Config)

	// TODO(sqs) HACK(sqs): run executors for both queues
	if os.Getenv("DEPLOY_TYPE") == "single-program" {
		otherConfig := *config
		if config.QueueName == "batches" {
			otherConfig.QueueName = "codeintel"
		} else {
			otherConfig.QueueName = "batches"
		}
		go func() {
			if err := Main(ctx, observationCtx, &otherConfig); err != nil {
				observationCtx.Logger.Fatal("executor for other queue failed", log.Error(err))
			}
		}()
	}

	return Main(ctx, observationCtx, config)
}

var Service service.Service = svc{}
