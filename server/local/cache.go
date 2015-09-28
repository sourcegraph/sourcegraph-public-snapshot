package local

import (
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sqs/grpccache"
)

const (
	maxMutableVCSAge = 15 * time.Second

	// NoSetCacheKey is the context key that is set to indicate that
	// the cache trailer should NOT be set. The cache trailer can only
	// be set once, so the first service method that is called that
	// sets the trailer effectively sets the cache duration. This
	// method typically maps to the gRPC endpoint, but sometimes
	// another service method is called before (e.g., in order to get
	// helper data). In this case, the caller should set this context
	// item to signal that no cache trailer should be set by the
	// auxiliary service method.
	NoSetCacheKey = "no-set-cache"
)

func cacheFor(ctx context.Context, maxAge time.Duration) {
	if ctx.Value(NoSetCacheKey) != nil {
		return
	}

	grpccache.SetCacheControl(ctx, grpccache.CacheControl{MaxAge: maxAge})
}

func noCache(ctx context.Context) {
	grpccache.SetCacheControl(ctx, grpccache.CacheControl{})
}

func veryShortCache(ctx context.Context) {
	cacheFor(ctx, 7*time.Second)
}

func shortCache(ctx context.Context) {
	cacheFor(ctx, 60*time.Second)
}

func mediumCache(ctx context.Context) {
	cacheFor(ctx, 300*time.Second)
}

func cacheOnCommitID(ctx context.Context, commitID string) {
	if len(commitID) == 40 {
		cacheForever(ctx)
	} else {
		cacheFor(ctx, maxMutableVCSAge)
	}
}

func cacheForever(ctx context.Context) {
	cacheFor(ctx, 10*365*24*time.Hour)
}
