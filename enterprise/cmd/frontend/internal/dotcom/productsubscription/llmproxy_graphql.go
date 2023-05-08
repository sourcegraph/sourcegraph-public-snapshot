package productsubscription

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type llmProxyAccessResolver struct{ sub *productSubscription }

func (l llmProxyAccessResolver) Enabled() bool { return l.sub.v.LLMProxyAccess.Enabled }

func (l llmProxyAccessResolver) RateLimit(ctx context.Context) (graphqlbackend.LLMProxyRateLimit, error) {
	if !l.sub.v.LLMProxyAccess.Enabled {
		return nil, nil
	}

	var rateLimit licensing.LLMProxyRateLimit

	// Get default access from active license. Call hydrate and access field directly to
	// avoid parsing license key which is done in (*productLicense).Info(), instead just
	// relying on what we know in DB.
	l.sub.hydrateActiveLicense(ctx)
	if l.sub.activeLicenseErr != nil {
		return nil, errors.Wrap(l.sub.activeLicenseErr, "could not get active license")
	}
	if l.sub.activeLicense != nil {
		rateLimit = licensing.NewLLMProxyRateLimit(licensing.PlanFromTags(l.sub.activeLicense.LicenseTags))
	}

	// Apply overrides
	rateLimitOverrides := l.sub.v.LLMProxyAccess
	if rateLimitOverrides.RateLimit != nil {
		rateLimit.Limit = *rateLimitOverrides.RateLimit
	}
	if rateLimitOverrides.RateIntervalSeconds != nil {
		rateLimit.IntervalSeconds = *rateLimitOverrides.RateIntervalSeconds
	}

	return &llmProxyRateLimitResolver{v: rateLimit}, nil
}

type llmProxyRateLimitResolver struct{ v licensing.LLMProxyRateLimit }

func (l *llmProxyRateLimitResolver) Limit() int32           { return l.v.Limit }
func (l *llmProxyRateLimitResolver) IntervalSeconds() int32 { return l.v.IntervalSeconds }
