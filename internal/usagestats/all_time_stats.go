package usagestats

import (
	"context"
	"sync/atomic"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

const (
	keyPrefix         = "user_activity:"
	fSearchOccurred   = "searchoccurred"
	fFindRefsOccurred = "findrefsoccurred"
)

var (
	store            = redispool.Store
	searchOccurred   int32
	findRefsOccurred int32
)

// logSiteSearchOccurred records that a search has occurred on the Sourcegraph instance.
func logSiteSearchOccurred() error {
	if !atomic.CompareAndSwapInt32(&searchOccurred, 0, 1) {
		return nil
	}
	key := keyPrefix + fSearchOccurred
	return store.Set(key, "true")
}

// logSiteFindRefsOccurred records that a search has occurred on the Sourcegraph instance.
func logSiteFindRefsOccurred() error {
	if !atomic.CompareAndSwapInt32(&findRefsOccurred, 0, 1) {
		return nil
	}
	key := keyPrefix + fFindRefsOccurred
	return store.Set(key, "true")
}

// HasSearchOccurred indicates whether a search has ever occurred on this instance.
func HasSearchOccurred(ctx context.Context) (bool, error) {
	s, err := store.WithContext(ctx).Get(keyPrefix + fSearchOccurred).Bool()
	if err != nil && err != redis.ErrNil {
		return s, err
	}
	return s, nil
}

// HasFindRefsOccurred indicates whether a find-refs has ever occurred on this instance.
func HasFindRefsOccurred(ctx context.Context) (bool, error) {
	r, err := store.WithContext(ctx).Get(keyPrefix + fFindRefsOccurred).Bool()
	if err != nil && err != redis.ErrNil {
		return r, err
	}
	return r, nil
}
