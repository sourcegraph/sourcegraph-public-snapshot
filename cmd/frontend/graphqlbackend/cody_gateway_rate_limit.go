package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

type RateLimitStatus interface {
	Feature() string
	Limit() BigInt
	Usage() BigInt
	PercentUsed() int32
	Interval() string
	NextLimitReset() *gqlutil.DateTime
}

func (r *siteResolver) CodyGatewayRateLimitStatus(ctx context.Context) (*[]RateLimitStatus, error) {
	// 🚨 SECURITY: Only site admins may check rate limits.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	cgc, ok := codygateway.NewClientFromSiteConfig(httpcli.ExternalDoer)
	if !ok {
		// Not configured for chat/autocomplete.
		return nil, nil
	}

	limits, err := cgc.GetLimits(ctx)
	if err != nil {
		return nil, err
	}

	rateLimits := make([]RateLimitStatus, 0, len(limits))
	for _, limit := range limits {
		rateLimits = append(rateLimits, &CodyRateLimit{
			RateLimitStatus: limit,
		})
	}

	return &rateLimits, nil
}

type CodyRateLimit struct {
	RateLimitStatus codygateway.LimitStatus
}

func (c *CodyRateLimit) Feature() string {
	return c.RateLimitStatus.Feature.DisplayName()

}

func (c *CodyRateLimit) Limit() BigInt {
	return BigInt(c.RateLimitStatus.IntervalLimit)
}

func (c *CodyRateLimit) Usage() BigInt {
	return BigInt(c.RateLimitStatus.IntervalUsage)
}

func (c *CodyRateLimit) PercentUsed() int32 {
	return int32(c.RateLimitStatus.PercentUsed())
}

func (c *CodyRateLimit) NextLimitReset() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(c.RateLimitStatus.Expiry)
}

func (c *CodyRateLimit) Interval() string {
	return c.RateLimitStatus.TimeInterval
}
