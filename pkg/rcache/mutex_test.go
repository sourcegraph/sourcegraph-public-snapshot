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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_868(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
