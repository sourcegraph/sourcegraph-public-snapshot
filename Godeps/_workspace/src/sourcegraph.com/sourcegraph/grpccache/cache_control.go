package grpccache

import (
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// CacheControl is passed by the CachedXyzServer wrapper to the
// underlying server's method implementation to allow control over the
// duration and nature of caching on a per-request basis.
type CacheControl struct {
	sync.RWMutex

	// MaxAge is maximum duration (since the original retrieval) that
	// an item is considered fresh.
	MaxAge time.Duration
}

func (cc *CacheControl) cacheable() bool {
	cc.RLock()
	v := cc.MaxAge > 0
	cc.RUnlock()
	return v
}

// IsZero returns true if cc refers to an empty CacheControl struct.
func (cc *CacheControl) IsZero() bool {
	cc.RLock()
	v := *cc == CacheControl{}
	cc.RUnlock()
	return v
}

// SetCacheControl is called by gRPC server method implementations to
// tell the client how to cache the result.
//
// The last CacheControl set on ctx in the course of handling a
// request is written a gRPC header and/or trailer to communicate the
// cache control info to the client. It may be called multiple times;
// only the last value is used.
//
// If ctx was not previously wrapped with Internal_WithCacheControl,
// then nothing will happen and the cache control info will not be
// returned. Ensure that the CachedXyzServer wrapper methods are being
// used.
func SetCacheControl(ctx context.Context, cc CacheControl) {
	existingCC := cacheControlFromContext(ctx)
	if existingCC != nil {
		existingCC.Lock()
		existingCC.MaxAge = cc.MaxAge
		existingCC.Unlock()
	}
}

// Internal_WithCacheControl is an internal func called by the
// code-genned CachedXyzServer wrapper methods. It should not be
// called by user code.
func Internal_WithCacheControl(ctx context.Context) (context.Context, *CacheControl) {
	cc := &CacheControl{}
	return context.WithValue(ctx, cacheControlKey, cc), cc
}

// Internal_SetCacheControlTrailer is an internal func called by the
// code-genned CachedXyzServer wrapper methods. It should not be
// called by user code.
func Internal_SetCacheControlTrailer(ctx context.Context, cc CacheControl) error {
	return grpc.SetTrailer(ctx, metadata.MD{"cache-control:max-age": []string{cc.MaxAge.String()}})
}

// TODO(sqs): warn if nil?
func cacheControlFromContext(ctx context.Context) *CacheControl {
	cc, _ := ctx.Value(cacheControlKey).(*CacheControl)
	return cc
}

// cacheControlFromContext is called on the client to retrieve the
// server's CacheControl response metadata.
func cacheControlFromMetadata(md metadata.MD) (*CacheControl, error) {
	var cc *CacheControl
	if maxAgeStrs, present := md["cache-control:max-age"]; present && len(maxAgeStrs) > 0 {
		// Take the last value assuming that it is the most recently set one.
		maxAge, err := time.ParseDuration(maxAgeStrs[len(maxAgeStrs)-1])
		if err != nil {
			return nil, err
		}
		if cc == nil {
			cc = new(CacheControl)
		}
		cc.MaxAge = maxAge
	}
	return cc, nil
}
