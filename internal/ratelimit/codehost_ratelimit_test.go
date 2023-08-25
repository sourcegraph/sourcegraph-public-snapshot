package ratelimit

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetCodeHostAPIRateLimitConfig(t *testing.T) {
	mockRL := NewMockRateLimiter()
	codeHostRL := NewCodeHostRateLimiter(mockRL)
	mockRL.SetTokenBucketReplenishmentFunc.SetDefaultHook(func(ctx context.Context, bucketKey string, quota, repRate int32) error {
		assert.Equal(t, fmt.Sprintf("%s:%s", "github.com", codeHostAPITokenBucketSuffix), bucketKey)
		assert.Equal(t, quota, int32(10))
		assert.Equal(t, repRate, int32(20))

		return nil
	})

	err := codeHostRL.SetCodeHostAPIRateLimitConfig(context.Background(), "github.com", 10, 20)
	assert.Nil(t, err)
}

func TestSetCodeHostGitRateLimitConfig(t *testing.T) {
	mockRL := NewMockRateLimiter()
	codeHostRL := NewCodeHostRateLimiter(mockRL)
	mockRL.SetTokenBucketReplenishmentFunc.SetDefaultHook(func(ctx context.Context, bucketKey string, quota, repRate int32) error {
		assert.Equal(t, fmt.Sprintf("%s:%s", "github.com", codeHostGitTokenBucketSuffix), bucketKey)
		assert.Equal(t, quota, int32(10))
		assert.Equal(t, repRate, int32(20))

		return nil
	})

	err := codeHostRL.SetCodeHostGitRateLimitConfig(context.Background(), "github.com", 10, 20)
	assert.Nil(t, err)
}
