package productsubscription

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type llmProxyAccessResolver struct {
	sub *productSubscription
}

func (r llmProxyAccessResolver) Enabled() bool { return r.sub.v.LLMProxyAccess.Enabled }

func (r llmProxyAccessResolver) RateLimit(ctx context.Context) (graphqlbackend.LLMProxyRateLimit, error) {
	if !r.Enabled() {
		return nil, nil
	}

	var rateLimit licensing.LLMProxyRateLimit

	// Get default access from active license. Call hydrate and access field directly to
	// avoid parsing license key which is done in (*productLicense).Info(), instead just
	// relying on what we know in DB.
	activeLicense, err := r.sub.computeActiveLicense(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not get active license")
	}
	var source graphqlbackend.LLMProxyRateLimitSource
	if activeLicense != nil {
		source = graphqlbackend.LLMProxyRateLimitSourcePlan
		rateLimit = licensing.NewLLMProxyRateLimit(licensing.PlanFromTags(activeLicense.LicenseTags))
	}

	// Apply overrides
	rateLimitOverrides := r.sub.v.LLMProxyAccess
	if rateLimitOverrides.RateLimit != nil {
		source = graphqlbackend.LLMProxyRateLimitSourceOverride
		rateLimit.Limit = *rateLimitOverrides.RateLimit
	}
	if rateLimitOverrides.RateIntervalSeconds != nil {
		source = graphqlbackend.LLMProxyRateLimitSourceOverride
		rateLimit.IntervalSeconds = *rateLimitOverrides.RateIntervalSeconds
	}

	return &llmProxyRateLimitResolver{v: rateLimit, source: source}, nil
}

func (r llmProxyAccessResolver) Usage(ctx context.Context) ([]graphqlbackend.LLMProxyUsageDatapoint, error) {
	if !r.Enabled() {
		return nil, nil
	}

	usage, err := NewLLMProxyService().UsageForSubscription(ctx, r.sub.UUID())
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.LLMProxyUsageDatapoint, 0, len(usage))
	for _, u := range usage {
		resolvers = append(resolvers, &llmProxyUsageDatapoint{
			date:  u.Date,
			count: u.Count,
		})
	}

	return resolvers, nil
}

type llmProxyRateLimitResolver struct {
	source graphqlbackend.LLMProxyRateLimitSource
	v      licensing.LLMProxyRateLimit
}

func (r *llmProxyRateLimitResolver) Source() graphqlbackend.LLMProxyRateLimitSource { return r.source }
func (r *llmProxyRateLimitResolver) Limit() int32                                   { return r.v.Limit }
func (r *llmProxyRateLimitResolver) IntervalSeconds() int32                         { return r.v.IntervalSeconds }

type llmProxyUsageDatapoint struct {
	date  time.Time
	count int
}

func (r *llmProxyUsageDatapoint) Date() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.date}
}

func (r *llmProxyUsageDatapoint) Count() int32 {
	return int32(r.count)
}
