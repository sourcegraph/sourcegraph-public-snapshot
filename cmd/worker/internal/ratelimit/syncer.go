package ratelimit

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// syncServices syncs a known slice of external services with their rate limiters without
// fetching them from the database.
func syncServices(ctx context.Context, services []*types.ExternalService, newRateLimiterFunc func(bucketName string) ratelimit.GlobalLimiter) error {
	var errs error
	for _, svc := range services {
		limit, err := extsvc.ExtractEncryptableRateLimit(ctx, svc.Config, svc.Kind)
		if err != nil {
			if errors.HasType[extsvc.ErrRateLimitUnsupported](err) {
				continue
			}
			errs = errors.Append(errs, errors.Wrap(err, "getting rate limit configuration"))
			continue
		}

		l := newRateLimiterFunc(svc.URN())
		lim := int32(-1)
		// rate.Inf should be stored as -1.
		if limit != rate.Inf {
			// Configured limits are per hour.
			lim = int32(limit * 3600)
		}
		if err := l.SetTokenBucketConfig(ctx, lim, time.Hour); err != nil {
			errs = errors.Append(errs, err)
			continue
		}
	}
	return errs
}
