package productsubscription

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type llmProxyAccessResolver struct {
	sub *productSubscription
}

func (r llmProxyAccessResolver) Enabled() bool { return r.sub.v.LLMProxyAccess.Enabled }

func (r llmProxyAccessResolver) ChatCompletionsRateLimit(ctx context.Context) (graphqlbackend.LLMProxyRateLimit, error) {
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
		rateLimit = licensing.NewLLMProxyChatRateLimit(licensing.PlanFromTags(activeLicense.LicenseTags), activeLicense.LicenseUserCount, activeLicense.LicenseTags)
	}

	// Apply overrides
	rateLimitOverrides := r.sub.v.LLMProxyAccess
	if rateLimitOverrides.ChatRateLimit.RateLimit != nil {
		source = graphqlbackend.LLMProxyRateLimitSourceOverride
		rateLimit.Limit = *rateLimitOverrides.ChatRateLimit.RateLimit
	}
	if rateLimitOverrides.ChatRateLimit.RateIntervalSeconds != nil {
		source = graphqlbackend.LLMProxyRateLimitSourceOverride
		rateLimit.IntervalSeconds = *rateLimitOverrides.ChatRateLimit.RateIntervalSeconds
	}
	if rateLimitOverrides.ChatRateLimit.AllowedModels != nil {
		source = graphqlbackend.LLMProxyRateLimitSourceOverride
		rateLimit.AllowedModels = rateLimitOverrides.ChatRateLimit.AllowedModels
	}

	return &llmProxyRateLimitResolver{
		feature: types.CompletionsFeatureChat,
		subUUID: r.sub.UUID(),
		v:       rateLimit,
		source:  source,
	}, nil
}

func (r llmProxyAccessResolver) CodeCompletionsRateLimit(ctx context.Context) (graphqlbackend.LLMProxyRateLimit, error) {
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
		rateLimit = licensing.NewLLMProxyCodeRateLimit(licensing.PlanFromTags(activeLicense.LicenseTags), activeLicense.LicenseUserCount, activeLicense.LicenseTags)
	}

	// Apply overrides
	rateLimitOverrides := r.sub.v.LLMProxyAccess
	if rateLimitOverrides.CodeRateLimit.RateLimit != nil {
		source = graphqlbackend.LLMProxyRateLimitSourceOverride
		rateLimit.Limit = *rateLimitOverrides.CodeRateLimit.RateLimit
	}
	if rateLimitOverrides.CodeRateLimit.RateIntervalSeconds != nil {
		source = graphqlbackend.LLMProxyRateLimitSourceOverride
		rateLimit.IntervalSeconds = *rateLimitOverrides.CodeRateLimit.RateIntervalSeconds
	}
	if rateLimitOverrides.CodeRateLimit.AllowedModels != nil {
		source = graphqlbackend.LLMProxyRateLimitSourceOverride
		rateLimit.AllowedModels = rateLimitOverrides.CodeRateLimit.AllowedModels
	}

	return &llmProxyRateLimitResolver{
		feature: types.CompletionsFeatureCode,
		subUUID: r.sub.UUID(),
		v:       rateLimit,
		source:  source,
	}, nil
}

type llmProxyRateLimitResolver struct {
	subUUID string
	feature types.CompletionsFeature
	source  graphqlbackend.LLMProxyRateLimitSource
	v       licensing.LLMProxyRateLimit
}

func (r *llmProxyRateLimitResolver) Source() graphqlbackend.LLMProxyRateLimitSource { return r.source }

func (r *llmProxyRateLimitResolver) AllowedModels() []string { return r.v.AllowedModels }

func (r *llmProxyRateLimitResolver) Limit() int32 { return r.v.Limit }

func (r *llmProxyRateLimitResolver) IntervalSeconds() int32 { return r.v.IntervalSeconds }

func (r llmProxyRateLimitResolver) Usage(ctx context.Context) ([]graphqlbackend.LLMProxyUsageDatapoint, error) {
	usage, err := NewLLMProxyService().UsageForSubscription(ctx, r.feature, r.subUUID)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.LLMProxyUsageDatapoint, 0, len(usage))
	for _, u := range usage {
		resolvers = append(resolvers, &llmProxyUsageDatapoint{
			date:  u.Date,
			model: u.Model,
			count: u.Count,
		})
	}

	return resolvers, nil
}

type llmProxyUsageDatapoint struct {
	date  time.Time
	model string
	count int
}

func (r *llmProxyUsageDatapoint) Date() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.date}
}

func (r *llmProxyUsageDatapoint) Model() string {
	return r.model
}

func (r *llmProxyUsageDatapoint) Count() int32 {
	return int32(r.count)
}
