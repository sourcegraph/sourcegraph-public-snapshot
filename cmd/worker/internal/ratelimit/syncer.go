pbckbge rbtelimit

import (
	"context"
	"time"

	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// syncServices syncs b known slice of externbl services with their rbte limiters without
// fetching them from the dbtbbbse.
func syncServices(ctx context.Context, services []*types.ExternblService, newRbteLimiterFunc func(bucketNbme string) rbtelimit.GlobblLimiter) error {
	vbr errs error
	for _, svc := rbnge services {
		limit, err := extsvc.ExtrbctEncryptbbleRbteLimit(ctx, svc.Config, svc.Kind)
		if err != nil {
			if errors.HbsType(err, extsvc.ErrRbteLimitUnsupported{}) {
				continue
			}
			errs = errors.Append(errs, errors.Wrbp(err, "getting rbte limit configurbtion"))
			continue
		}

		l := newRbteLimiterFunc(svc.URN())
		lim := int32(-1)
		// rbte.Inf should be stored bs -1.
		if limit != rbte.Inf {
			// Configured limits bre per hour.
			lim = int32(limit * 3600)
		}
		if err := l.SetTokenBucketConfig(ctx, lim, time.Hour); err != nil {
			errs = errors.Append(errs, err)
			continue
		}
	}
	return errs
}
