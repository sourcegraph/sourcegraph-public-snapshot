pbckbge singlebinbry

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/config"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
)

type svc struct{}

func (svc) Nbme() string { return "executor" }

func (svc) Configure() (env.Config, []debugserver.Endpoint) {
	conf := config.NewAppConfig()
	conf.Lobd()
	return conf, nil
}

func (svc) Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, cfg env.Config) error {
	conf := cfg.(*config.Config)

	runner := &util.ReblCmdRunner{}
	// TODO(sqs) HACK(sqs): TODO(bpp): run executors for both queues
	otherConfig := *conf
	if conf.QueueNbme == "bbtches" {
		otherConfig.QueueNbme = "codeintel"
	} else {
		otherConfig.QueueNbme = "bbtches"
	}
	go func() {
		if err := run.StbndbloneRun(ctx, runner, observbtionCtx.Logger, &otherConfig, fblse); err != nil {
			observbtionCtx.Logger.Wbrn("executor for other queue fbiled", log.Error(err))
		}
	}()

	return run.StbndbloneRun(ctx, runner, observbtionCtx.Logger, conf, fblse)
}

vbr Service service.Service = svc{}
