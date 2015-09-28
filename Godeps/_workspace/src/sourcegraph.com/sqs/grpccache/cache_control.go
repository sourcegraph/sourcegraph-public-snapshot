package grpccache

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// CacheControl is passed by the CachedXyzServer wrapper to the
// underlying server's method implementation to allow control over the
// duration and nature of caching on a per-request basis.
type CacheControl struct {
	// MaxAge is maximum duration (since the original retrieval) that
	// an item is considered fresh.
	MaxAge time.Duration
}

func (cc *CacheControl) cacheable() bool {
	return cc.MaxAge > 0
}

// IsZero returns true if cc refers to an empty CacheControl struct.
func (cc *CacheControl) IsZero() bool {
	return *cc == CacheControl{}
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
		*existingCC = cc
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
	return grpc.SetTrailer(ctx, metadata.MD{"cache-control:max-age": cc.MaxAge.String()})
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
	if maxAgeStr, present := md["cache-control:max-age"]; present {
		maxAge, err := time.ParseDuration(maxAgeStr)
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
