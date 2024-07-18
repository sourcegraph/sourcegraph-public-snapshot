package productsubscription

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayactor"
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayevents"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type codyGatewayAccessResolver struct {
	sub *productSubscription
}

// NewCodyGatewayAccessResolver returns a new CodyGatewayAccessResolver for the
// given product subscription. ONLY FOR TESTING, DO NOT USE - see package
// 'dotcomproductsubscriptionstest', this should be removed when that package
// is removed.
func NewCodyGatewayAccessResolver(ctx context.Context, logger log.Logger, db database.DB, subID string) (*codyGatewayAccessResolver, error) {
	sub, err := productSubscriptionByDBID(ctx, logger, db, subID, "access")
	if err != nil {
		return nil, err
	}
	return &codyGatewayAccessResolver{sub: sub}, nil
}

func (r codyGatewayAccessResolver) Enabled() bool { return r.sub.v.CodyGatewayAccess.Enabled }

func (r codyGatewayAccessResolver) ChatCompletionsRateLimit(ctx context.Context) (graphqlbackend.CodyGatewayRateLimit, error) {
	if !r.Enabled() {
		return nil, nil
	}

	var rateLimit licensing.CodyGatewayRateLimit

	// Get default access from active license. Call hydrate and access field directly to
	// avoid parsing license key which is done in (*productLicense).Info(), instead just
	// relying on what we know in DB.
	activeLicense, err := r.sub.computeActiveLicense(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not get active license")
	}
	var source graphqlbackend.CodyGatewayRateLimitSource
	if activeLicense != nil {
		source = graphqlbackend.CodyGatewayRateLimitSourcePlan
		rateLimit = licensing.NewCodyGatewayChatRateLimit(licensing.PlanFromTags(activeLicense.LicenseTags), activeLicense.LicenseUserCount)
	}

	// Apply overrides
	rateLimitOverrides := r.sub.v.CodyGatewayAccess
	if rateLimitOverrides.ChatRateLimit.RateLimit != nil {
		source = graphqlbackend.CodyGatewayRateLimitSourceOverride
		rateLimit.Limit = *rateLimitOverrides.ChatRateLimit.RateLimit
	}
	if rateLimitOverrides.ChatRateLimit.RateIntervalSeconds != nil {
		source = graphqlbackend.CodyGatewayRateLimitSourceOverride
		rateLimit.IntervalSeconds = *rateLimitOverrides.ChatRateLimit.RateIntervalSeconds
	}

	return &codyGatewayRateLimitResolver{
		feature:     types.CompletionsFeatureChat,
		actorID:     r.sub.UUID(),
		actorSource: codygatewayactor.ActorSourceEnterpriseSubscription,
		v:           rateLimit,
		source:      source,
	}, nil
}

func (r codyGatewayAccessResolver) CodeCompletionsRateLimit(ctx context.Context) (graphqlbackend.CodyGatewayRateLimit, error) {
	if !r.Enabled() {
		return nil, nil
	}

	var rateLimit licensing.CodyGatewayRateLimit

	// Get default access from active license. Call hydrate and access field directly to
	// avoid parsing license key which is done in (*productLicense).Info(), instead just
	// relying on what we know in DB.
	activeLicense, err := r.sub.computeActiveLicense(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not get active license")
	}
	var source graphqlbackend.CodyGatewayRateLimitSource
	if activeLicense != nil {
		source = graphqlbackend.CodyGatewayRateLimitSourcePlan
		rateLimit = licensing.NewCodyGatewayCodeRateLimit(licensing.PlanFromTags(activeLicense.LicenseTags), activeLicense.LicenseUserCount)
	}

	// Apply overrides
	rateLimitOverrides := r.sub.v.CodyGatewayAccess
	if rateLimitOverrides.CodeRateLimit.RateLimit != nil {
		source = graphqlbackend.CodyGatewayRateLimitSourceOverride
		rateLimit.Limit = *rateLimitOverrides.CodeRateLimit.RateLimit
	}
	if rateLimitOverrides.CodeRateLimit.RateIntervalSeconds != nil {
		source = graphqlbackend.CodyGatewayRateLimitSourceOverride
		rateLimit.IntervalSeconds = *rateLimitOverrides.CodeRateLimit.RateIntervalSeconds
	}

	return &codyGatewayRateLimitResolver{
		feature:     types.CompletionsFeatureCode,
		actorID:     r.sub.UUID(),
		actorSource: codygatewayactor.ActorSourceEnterpriseSubscription,
		v:           rateLimit,
		source:      source,
	}, nil
}

func (r codyGatewayAccessResolver) EmbeddingsRateLimit(ctx context.Context) (graphqlbackend.CodyGatewayRateLimit, error) {
	if !r.Enabled() {
		return nil, nil
	}

	var rateLimit licensing.CodyGatewayRateLimit

	// Get default access from active license. Call hydrate and access field directly to
	// avoid parsing license key which is done in (*productLicense).Info(), instead just
	// relying on what we know in DB.
	activeLicense, err := r.sub.computeActiveLicense(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not get active license")
	}
	var source graphqlbackend.CodyGatewayRateLimitSource
	if activeLicense != nil {
		source = graphqlbackend.CodyGatewayRateLimitSourcePlan
		rateLimit = licensing.NewCodyGatewayEmbeddingsRateLimit(licensing.PlanFromTags(activeLicense.LicenseTags), activeLicense.LicenseUserCount)
	}

	// Apply overrides
	rateLimitOverrides := r.sub.v.CodyGatewayAccess
	if rateLimitOverrides.EmbeddingsRateLimit.RateLimit != nil {
		source = graphqlbackend.CodyGatewayRateLimitSourceOverride
		rateLimit.Limit = *rateLimitOverrides.EmbeddingsRateLimit.RateLimit
	}
	if rateLimitOverrides.EmbeddingsRateLimit.RateIntervalSeconds != nil {
		source = graphqlbackend.CodyGatewayRateLimitSourceOverride
		rateLimit.IntervalSeconds = *rateLimitOverrides.EmbeddingsRateLimit.RateIntervalSeconds
	}

	return &codyGatewayRateLimitResolver{
		actorID:     r.sub.UUID(),
		actorSource: codygatewayactor.ActorSourceEnterpriseSubscription,
		v:           rateLimit,
		source:      source,
	}, nil
}

type codyGatewayRateLimitResolver struct {
	actorID     string
	actorSource codygatewayactor.ActorSource
	feature     types.CompletionsFeature
	source      graphqlbackend.CodyGatewayRateLimitSource
	v           licensing.CodyGatewayRateLimit
}

func (r *codyGatewayRateLimitResolver) Source() graphqlbackend.CodyGatewayRateLimitSource {
	return r.source
}

func (r *codyGatewayRateLimitResolver) AllowedModels() []string { return r.v.AllowedModels }

func (r *codyGatewayRateLimitResolver) Limit() graphqlbackend.BigInt {
	return graphqlbackend.BigInt(r.v.Limit)
}

func (r *codyGatewayRateLimitResolver) IntervalSeconds() int32 { return r.v.IntervalSeconds }

func (r codyGatewayRateLimitResolver) Usage(ctx context.Context) ([]graphqlbackend.CodyGatewayUsageDatapoint, error) {
	var (
		usage []codygatewayevents.SubscriptionUsage
		err   error
	)
	if r.feature != "" {
		usage, err = NewCodyGatewayService().CompletionsUsageForActor(ctx, r.feature, r.actorSource, r.actorID)
		if err != nil {
			return nil, err
		}
	} else {
		usage, err = NewCodyGatewayService().EmbeddingsUsageForActor(ctx, r.actorSource, r.actorID)
		if err != nil {
			return nil, err
		}
	}

	resolvers := make([]graphqlbackend.CodyGatewayUsageDatapoint, 0, len(usage))
	for _, u := range usage {
		resolvers = append(resolvers, &codyGatewayUsageDatapoint{
			date:  u.Date,
			model: u.Model,
			count: u.Count,
		})
	}

	return resolvers, nil
}

type codyGatewayUsageDatapoint struct {
	date  time.Time
	model string
	count int64
}

func (r *codyGatewayUsageDatapoint) Date() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.date}
}

func (r *codyGatewayUsageDatapoint) Model() string {
	return r.model
}

func (r *codyGatewayUsageDatapoint) Count() graphqlbackend.BigInt {
	return graphqlbackend.BigInt(r.count)
}
