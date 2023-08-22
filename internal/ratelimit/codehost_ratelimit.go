package ratelimit

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

var (
	codeHostAPITokenBucketSuffix = "api_tokens"
	codeHostGitTokenBucketSuffix = "git_tokens"
)

type codeHostRateLimiter struct {
	rateLimiter redispool.RateLimiter
}

type CodeHostRateLimiter interface {
	SetCodeHostAPIRateLimitConfig(ctx context.Context, codeHostURL string, apiQuota, apiReplenishmentIntervalSeconds int32) error
	SetCodeHostGitRateLimitConfig(ctx context.Context, codeHostURL string, gitQuota, gitReplenishmentIntervalSeconds int32) error
}

func NewCodeHostRateLimiter(rl redispool.RateLimiter) CodeHostRateLimiter {
	return &codeHostRateLimiter{
		rateLimiter: rl,
	}
}

func (c *codeHostRateLimiter) SetCodeHostAPIRateLimitConfig(ctx context.Context, codeHostURL string, apiQuota, apiReplenishmentIntervalSeconds int32) error {
	return c.setCodeHostRateLimitConfig(ctx, fmt.Sprintf("%s:%s", codeHostURL, codeHostAPITokenBucketSuffix), apiQuota, apiReplenishmentIntervalSeconds)
}

func (c *codeHostRateLimiter) SetCodeHostGitRateLimitConfig(ctx context.Context, codeHostURL string, gitQuota, gitReplenishmentIntervalSeconds int32) error {
	return c.setCodeHostRateLimitConfig(ctx, fmt.Sprintf("%s:%s", codeHostURL, codeHostGitTokenBucketSuffix), gitQuota, gitReplenishmentIntervalSeconds)
}

func (c *codeHostRateLimiter) setCodeHostRateLimitConfig(ctx context.Context, codeHostBucketName string, quota, replenishmentIntervalSeconds int32) error {
	return c.rateLimiter.SetTokenBucketReplenishment(ctx, codeHostBucketName, quota, replenishmentIntervalSeconds)
}
