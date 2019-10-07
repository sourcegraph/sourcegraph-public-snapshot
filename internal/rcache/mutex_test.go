package rcache

import (
	"testing"

	"context"
)

func TestTryAcquireMutex(t *testing.T) {
	SetupForTest(t)

	ctx, release, ok := TryAcquireMutex(context.Background(), "test")
	if !ok {
		t.Fatalf("expected to acquire mutex")
	}

	if _, _, ok = TryAcquireMutex(context.Background(), "test"); ok {
		t.Fatalf("expected to fail to acquire mutex")
	}

	release()
	if ctx.Err() == nil {
		t.Errorf("expected to cancel context")
	}

	// Test out if cancelling the parent context allows us to still release
	ctx, cancel := context.WithCancel(context.Background())
	_, release, ok = TryAcquireMutex(ctx, "test")
	if !ok {
		t.Fatalf("expected to acquire mutex")
	}
	cancel()
	release()

	_, release, ok = TryAcquireMutex(context.Background(), "test")
	if !ok {
		t.Fatalf("expected to acquire mutex")
	}
	release()
}
