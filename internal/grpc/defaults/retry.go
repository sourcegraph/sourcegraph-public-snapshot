package defaults

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	internalRetryDelayBase, _   = time.ParseDuration(env.Get("SRC_GRPC_RETRY_DELAY_BASE", "50ms", "Base retry delay duration for internal GRPC requests"))
	internalRetryMaxAttempts, _ = strconv.Atoi(env.Get("SRC_GRPC_RETRY_MAX_ATTEMPTS", "20", "Max retry attempts for internal GRPC requests"))
	internalRetryMaxDuration, _ = time.ParseDuration(env.Get("SRC_GRPC_RETRY_MAX_DURATION", "3s", "Max retry duration for internal GRPC requests"))
)

// RetryPolicy is the default retry policy for internal GRPC requests.
//
// The retry policy will trigger on Unavailable and ResourceExhausted status errors, and will retry up to 20 times using an
// exponential backoff policy with a maximum duration of 3s in between retries.
//
// Only Unary (1:1) and ServerStreaming (1:N) requests are retried. All other types of requests will immediately
// return an Unimplemented status error. It's up to the caller to manually retry these requests.
//
// These defaults can be overridden with the following environment variables:
// - SRC_GRPC_RETRY_DELAY_BASE: Base retry delay duration for internal GRPC requests
// - SRC_GRPC_RETRY_MAX_ATTEMPTS: Max retry attempts for internal GRPC requests
// - SRC_GRPC_RETRY_MAX_DURATION: Max retry duration for internal GRPC requests
var RetryPolicy = []grpc.CallOption{
	retry.WithCodes(codes.Unavailable, codes.ResourceExhausted),

	// Together with the default options, the maximum delay will behave like this:
	// Retry# Delay
	// 1	0.05s
	// 2	0.1s
	// 3	0.2s
	// 4	0.4s
	// 5	0.8s
	// 6	1.6s
	// 7	3.0s
	// 8	3.0s
	// ...
	// 20	3.0s
	retry.WithMax(uint(internalRetryMaxAttempts)),
	retry.WithBackoff(fullJitter(internalRetryDelayBase, internalRetryMaxDuration)),
}

// fullJitter returns a retry.BackOff function that generates
// a random backoff duration in the range [base, min(base*2^attempt, max)).
//
// base and max must both be >= 0, and max must be greater than base. If any of these
// conditions are not met, this function panics.
//
// See the full jitter algorithm described here:
// http://www.awsarchitectureblog.com/2015/03/backoff.html
func fullJitter(base, max time.Duration) retry.BackoffFunc {
	if base < 0 {
		panic(fmt.Sprintf("base must be >= 0, got %v", base))
	}

	if base >= max {
		panic(fmt.Sprintf("max must be > base, got base=%v and max=%v", base, max))
	}

	return func(ctx context.Context, attempt uint) time.Duration {
		if attempt <= 1 {
			return base // save some CPU cycles
		}

		// Note: "attempt" is always > 0, so this is safe
		multiplier := (1 << attempt) >> 1 // powers of 2: 1, 2, 4, 8

		upperLimit := base * time.Duration(multiplier) // base * 2^attempt
		if !(base < upperLimit && upperLimit <= max) { // handle underflow, overflow, or something greater than our specified max
			upperLimit = max
		}

		jitter := time.Duration(rand.Int63n(int64(upperLimit - base)))
		return base + jitter
	}
}
