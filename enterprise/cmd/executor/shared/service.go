package shared

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "executor" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	var config config.Config
	config.Load()
	return &config, nil
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, cfg env.Config) error {
	config := cfg.(*config.Config)

	// TODO(sqs) HACK(sqs): run executors for both queues
	if deploy.IsDeployTypeSingleProgram(deploy.Type()) {
		otherConfig := *config
		if config.QueueName == "batches" {
			otherConfig.QueueName = "codeintel"
		} else {
			otherConfig.QueueName = "batches"
		}
		go func() {
			if err := Main(ctx, observationCtx, ready, &otherConfig); err != nil {
				observationCtx.Logger.Fatal("executor for other queue failed", log.Error(err))
			}
		}()
	}

	return Main(ctx, observationCtx, ready, config)
}

var Service service.Service = svc{}
