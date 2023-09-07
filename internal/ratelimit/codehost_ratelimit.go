package ratelimit

import (
	"context"
	"fmt"
	"time"

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
	SetCodeHostAPIRateLimitConfig(ctx context.Context, codeHostURL string, apiQuota int32, apiReplenishmentInterval time.Duration) error
	SetCodeHostGitRateLimitConfig(ctx context.Context, codeHostURL string, gitQuota int32, gitReplenishmentInterval time.Duration) error
}

func NewCodeHostRateLimiter(rl redispool.RateLimiter) CodeHostRateLimiter {
	return &codeHostRateLimiter{
		rateLimiter: rl,
	}
}

func (c *codeHostRateLimiter) SetCodeHostAPIRateLimitConfig(ctx context.Context, codeHostURL string, apiQuota int32, apiReplenishmentInterval time.Duration) error {
	return c.setCodeHostRateLimitConfig(ctx, fmt.Sprintf("%s:%s", codeHostURL, codeHostAPITokenBucketSuffix), apiQuota, apiReplenishmentInterval)
}

func (c *codeHostRateLimiter) SetCodeHostGitRateLimitConfig(ctx context.Context, codeHostURL string, gitQuota int32, gitReplenishmentInterval time.Duration) error {
	return c.setCodeHostRateLimitConfig(ctx, fmt.Sprintf("%s:%s", codeHostURL, codeHostGitTokenBucketSuffix), gitQuota, gitReplenishmentInterval)
}

func (c *codeHostRateLimiter) setCodeHostRateLimitConfig(ctx context.Context, codeHostBucketName string, quota int32, replenishmentInterval time.Duration) error {
	return c.rateLimiter.SetTokenBucketConfig(ctx, codeHostBucketName, quota, replenishmentInterval)
}
