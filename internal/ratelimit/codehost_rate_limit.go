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
	SetCodeHostAPIRateLimitConfig(ctx context.Context, codeHostURL string, apiQuota, apiReplenishmentInterval int32) error
	SetCodeHostGitRateLimitConfig(ctx context.Context, codeHostURL string, gitQuota, gitReplenishmentInterval int32) error
}

func NewCodeHostRateLimiter(rl redispool.RateLimiter) CodeHostRateLimiter {
	return &codeHostRateLimiter{
		rateLimiter: rl,
	}
}

func (c *codeHostRateLimiter) SetCodeHostAPIRateLimitConfig(ctx context.Context, codeHostURL string, apiQuota, apiReplenishmentInterval int32) error {
	return c.setCodeHostRateLimitConfig(ctx, fmt.Sprintf("%s:%s", codeHostURL, codeHostAPITokenBucketSuffix), apiQuota, apiReplenishmentInterval)
}

func (c *codeHostRateLimiter) SetCodeHostGitRateLimitConfig(ctx context.Context, codeHostURL string, gitQuota, gitReplenishmentInterval int32) error {
	return c.setCodeHostRateLimitConfig(ctx, fmt.Sprintf("%s:%s", codeHostURL, codeHostGitTokenBucketSuffix), gitQuota, gitReplenishmentInterval)
}

func (c *codeHostRateLimiter) setCodeHostRateLimitConfig(ctx context.Context, codeHostBucketName string, quota, replenishmentInterval int32) error {
	return c.rateLimiter.SetTokenBucketReplenishment(ctx, codeHostBucketName, quota, replenishmentInterval)
}
