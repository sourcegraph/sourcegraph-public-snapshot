package singlebinary

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/run"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

type svc struct{}

func (svc) Name() string { return "executor" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	conf := config.NewAppConfig()
	conf.Load()
	return conf, nil
}

func (svc) Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, cfg env.Config) error {
	conf := cfg.(*config.Config)

	runner := &util.RealCmdRunner{}
	// TODO(sqs) HACK(sqs): TODO(app): run executors for both queues
	otherConfig := *conf
	if conf.QueueName == "batches" {
		otherConfig.QueueName = "codeintel"
	} else {
		otherConfig.QueueName = "batches"
	}
	go func() {
		if err := run.StandaloneRun(ctx, runner, observationCtx.Logger, &otherConfig, false); err != nil {
			observationCtx.Logger.Warn("executor for other queue failed", log.Error(err))
		}
	}()

	return run.StandaloneRun(ctx, runner, observationCtx.Logger, conf, false)
}

var Service service.Service = svc{}
