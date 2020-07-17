package leader

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

func TestDoWhileLeader(t *testing.T) {
	rcache.SetupForTest(t)

	key := "test-leader"
	ctx, cancel := context.WithCancel(context.Background())
	// In case we don't make it to cancel lower down
	t.Cleanup(cancel)

	var count int64

	fn := func(ctx context.Context, release func()) {
		select {
		case <-ctx.Done():
			return
		default:
		}
		atomic.AddInt64(&count, 1)
		defer release()
		<-ctx.Done()
	}

	options := Options{
		AcquireInterval: 50 * time.Millisecond,
		MutexOptions: rcache.MutexOptions{
			Tries:      1,
			RetryDelay: 10 * time.Millisecond,
		},
	}

	go Do(ctx, key, fn, options)
	go Do(ctx, key, fn, options)

	time.Sleep(500 * time.Millisecond)
	cancel()

	if count != 1 {
		t.Fatalf("Count > 1: %d", count)
	}
}
