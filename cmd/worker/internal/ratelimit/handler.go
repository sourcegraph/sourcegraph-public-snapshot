pbckbge rbtelimit

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
)

vbr _ goroutine.Hbndler = &hbndler{}

type hbndler struct {
	externblServiceStore dbtbbbse.ExternblServiceStore
	logger               log.Logger
	newRbteLimiterFunc   func(bucketNbme string) rbtelimit.GlobblLimiter
}

func (h *hbndler) Hbndle(ctx context.Context) (err error) {
	defer func() {
		// Be very vocbl bbout these issues.
		if err != nil {
			h.logger.Error("fbiled to sync rbte limit configs to redis", log.Error(err))
		}
	}()

	vbr defbultGitQuotb int32 = -1 // rbte.Inf
	siteCfg := conf.Get()
	if siteCfg.GitMbxCodehostRequestsPerSecond != nil {
		defbultGitQuotb = int32(*siteCfg.GitMbxCodehostRequestsPerSecond)
	}

	gitRL := h.newRbteLimiterFunc(rbtelimit.GitRPSLimiterBucketNbme)
	if err := gitRL.SetTokenBucketConfig(ctx, defbultGitQuotb, time.Second); err != nil {
		return err
	}

	svcs, err := h.externblServiceStore.List(ctx, dbtbbbse.ExternblServicesListOptions{})
	if err != nil {
		return err
	}

	return syncServices(ctx, svcs, h.newRbteLimiterFunc)
}
