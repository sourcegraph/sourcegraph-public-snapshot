package repos

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type externalServiceLister interface {
	List(context.Context, database.ExternalServicesListOptions) ([]*types.ExternalService, error)
}

// RateLimitSyncer syncs rate limits based on external service configuration
type RateLimitSyncer struct {
	registry      *ratelimit.Registry
	serviceLister externalServiceLister
	// How many services to fetch in each DB call
	pageSize int
	// Rate limit to apply when making DB requests, optional.
	limiter *ratelimit.InstrumentedLimiter
}

type RateLimitSyncerOpts struct {
	// The number of external services to fetch while paginating. Optional, will
	// default to 500.
	PageSize int
	// We need to rate limit our rate limit syncing (!). This is because when
	// encryption is enabled on an instance, fetching external services is not free
	// as it might require a decryption step. On Cloud this incurs an API call to
	// Cloud KMS.
	//
	// If a limiter is supplied we ensure that PageSize is never larger than the
	// limiters burst size. The limiter is optional.
	Limiter *ratelimit.InstrumentedLimiter
}

// NewRateLimitSyncer returns a new syncer
func NewRateLimitSyncer(registry *ratelimit.Registry, serviceLister externalServiceLister, opts RateLimitSyncerOpts) *RateLimitSyncer {
	pageSize := opts.PageSize
	if pageSize == 0 {
		pageSize = 500
	}
	if opts.Limiter != nil && pageSize > opts.Limiter.Burst() {
		pageSize = opts.Limiter.Burst()
	}
	r := &RateLimitSyncer{
		registry:      registry,
		serviceLister: serviceLister,
		pageSize:      pageSize,
		limiter:       opts.Limiter,
	}
	return r
}

// SyncRateLimiters syncs rate limiters for external services, the sync will
// happen for all external services if no IDs are given.
func (r *RateLimitSyncer) SyncRateLimiters(ctx context.Context, ids ...int64) error {
	return r.SyncLimitersSince(ctx, time.Time{}, ids...)
}

// SyncLimitersSince is the same as SyncRateLimiters but will only sync rate limiters
// for external service that have been update after `updateAfter`.
func (r *RateLimitSyncer) SyncLimitersSince(ctx context.Context, updateAfter time.Time, ids ...int64) error {
	cursor := database.LimitOffset{
		Limit: r.pageSize,
	}
	for {
		if r.limiter != nil {
			if err := r.limiter.WaitN(ctx, cursor.Limit); err != nil {
				return errors.Wrap(err, "waiting for rate limiter")
			}
		}
		services, err := r.serviceLister.List(ctx,
			database.ExternalServicesListOptions{
				IDs:          ids,
				LimitOffset:  &cursor,
				UpdatedAfter: updateAfter,
			},
		)
		if err != nil {
			return errors.Wrap(err, "listing external services")
		}

		if len(services) == 0 {
			break
		}
		cursor.Offset += len(services)

		if err := r.SyncServices(ctx, services); err != nil {
			return errors.Wrap(err, "syncing services")
		}

		if len(services) < r.pageSize {
			break
		}
	}
	return nil
}

// SyncServices syncs a know slice of services without fetching them from the
// database.
func (r *RateLimitSyncer) SyncServices(ctx context.Context, services []*types.ExternalService) error {
	for _, svc := range services {
		limit, err := extsvc.ExtractEncryptableRateLimit(ctx, svc.Config, svc.Kind)
		if err != nil {
			if errors.HasType(err, extsvc.ErrRateLimitUnsupported{}) {
				continue
			}
			return errors.Wrap(err, "getting rate limit configuration")
		}

		l := r.registry.Get(svc.URN())
		l.SetLimit(limit)
	}
	return nil
}
