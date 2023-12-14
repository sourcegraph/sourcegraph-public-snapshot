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
		// Not configured.
		return nil, nil
	}

	limits, err := cgc.GetLimits(ctx)
	if err != nil {
		return nil, err
	}

	rateLimits := make([]RateLimitStatus, 0, len(limits))
	for _, limit := range limits {
		rateLimits = append(rateLimits, &codyRateLimit{
			rl: limit,
		})
	}

	return &rateLimits, nil
}

type codyRateLimit struct {
	rl codygateway.LimitStatus
}

func (c *codyRateLimit) Feature() string {
	return c.rl.Feature.DisplayName()

}

func (c *codyRateLimit) Limit() BigInt {
	return BigInt(c.rl.IntervalLimit)
}

func (c *codyRateLimit) Usage() BigInt {
	return BigInt(c.rl.IntervalUsage)
}

func (c *codyRateLimit) PercentUsed() int32 {
	return int32(c.rl.PercentUsed())
}

func (c *codyRateLimit) NextLimitReset() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(c.rl.Expiry)
}

func (c *codyRateLimit) Interval() string {
	return c.rl.TimeInterval
}
