package ratelimit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetCodeHostAPIRateLimitConfig(t *testing.T) {
	mockRL := NewMockRateLimiter()
	codeHostRL := NewCodeHostRateLimiter(mockRL)
	mockRL.SetTokenBucketConfigFunc.SetDefaultHook(func(ctx context.Context, bucketKey string, quota int32, repRate time.Duration) error {
		assert.Equal(t, fmt.Sprintf("%s:%s", "github.com", codeHostAPITokenBucketSuffix), bucketKey)
		assert.Equal(t, quota, int32(10))
		assert.Equal(t, repRate, time.Duration(20)*time.Second)

		return nil
	})

	err := codeHostRL.SetCodeHostAPIRateLimitConfig(context.Background(), "github.com", 10, time.Duration(20)*time.Second)
	assert.Nil(t, err)
}

func TestSetCodeHostGitRateLimitConfig(t *testing.T) {
	mockRL := NewMockRateLimiter()
	codeHostRL := NewCodeHostRateLimiter(mockRL)
	mockRL.SetTokenBucketConfigFunc.SetDefaultHook(func(ctx context.Context, bucketKey string, quota int32, repRate time.Duration) error {
		assert.Equal(t, fmt.Sprintf("%s:%s", "github.com", codeHostGitTokenBucketSuffix), bucketKey)
		assert.Equal(t, quota, int32(10))
		assert.Equal(t, repRate, time.Duration(20)*time.Second)

		return nil
	})

	err := codeHostRL.SetCodeHostGitRateLimitConfig(context.Background(), "github.com", 10, time.Duration(20)*time.Second)
	assert.Nil(t, err)
}
