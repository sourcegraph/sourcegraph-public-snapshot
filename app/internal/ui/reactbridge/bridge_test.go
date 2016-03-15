package reactbridge

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestBridge_simple(t *testing.T) {
	b := New(`function main(a, callback) { callback("Hello, " + a + "!"); }`, 1, nil)

	ctx := context.Background()
	resp, err := b.CallMain(ctx, "world")
	if err != nil {
		t.Fatal(err)
	}

	if want := "Hello, world!"; resp != want {
		t.Errorf("got %q, want %q", resp, want)
	}
}

func TestBridge_reuse(t *testing.T) {
	b := New(`function main(a, callback) { callback(a); }`, 1, nil)

	ctx := context.Background()
	for i := 0; i < 50; i++ {
		want := fmt.Sprintf("hello %d", i)
		resp, err := b.CallMain(ctx, want)
		if err != nil {
			t.Fatal(err)
		}
		if resp != want {
			t.Errorf("got %q, want %q", resp, want)
		}
	}
}

func TestBridge_concurrent(t *testing.T) {
	b := New(`function main(a, callback) { callback(a); }`, 1, nil)

	var wg sync.WaitGroup
	ctx := context.Background()
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			want := fmt.Sprintf("hello %d", i)
			resp, err := b.CallMain(ctx, want)
			if err != nil {
				t.Fatal(err)
			}
			if resp != want {
				t.Errorf("got %q, want %q", resp, want)
			}
		}()
	}
	wg.Wait()
}

func TestBridge_async(t *testing.T) {
	b := New(`
function main(a, callback) {
	setTimeout(function() {
		callback("Hello, " + a + "!");
	}, 100);
}`, 1, nil)

	ctx := context.Background()
	resp, err := b.CallMain(ctx, "world")
	if err != nil {
		t.Fatal(err)
	}

	if want := "Hello, world!"; resp != want {
		t.Errorf("got %q, want %q", resp, want)
	}
}

func TestBridge_timeout(t *testing.T) {
	b := New(`
function main(a, callback) {
	setTimeout(function() {
		callback("Hello, " + a + "!");
	}, 200);
}`, 1, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	resp, err := b.CallMain(ctx, "world")
	if want := ""; resp != want {
		t.Errorf("got resp %q, want %q", resp, want)
	}
	if want := context.DeadlineExceeded; err != want {
		t.Errorf("got error %q, want %q", err, want)
	}

	// Wait just to see what happens when the callback is called (to
	// make sure it doesn't crash).
	time.Sleep(200 * time.Millisecond)
}

func TestBridge_throw(t *testing.T) {
	b := New(`function main(a, callback) { throw new Error("error"); }`, 1, nil)

	ctx := context.Background()
	resp, err := b.CallMain(ctx, "")
	if resp != "" {
		t.Errorf("got resp %q, want empty", resp)
	}
	if want := "Error: error"; err == nil || !strings.HasPrefix(err.Error(), want) {
		t.Errorf("got error %q, want it to have prefix %q", err, want)
	}
}

func TestBridge_timeoutThrow(t *testing.T) {
	b := New(`
function f(i, callback) {
	if (i > 50) {
		setTimeout(function() {
			callback("a");
		}, 0);
	} else {
		f(i+1, callback);
	}
}

function main(a, callback) {
	f(0, callback);
}`, 1, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()
	resp, err := b.CallMain(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if want := "a"; resp != want {
		t.Errorf("got resp %q, want %q", resp, want)
	}
}

func TestBridge_compileError(t *testing.T) {
	errCh := make(chan error)
	New(`function main(a, ,) { { }`, 1, errCh)
	err := <-errCh
	if want := "SyntaxError: expected identifier (line 1)"; err == nil || err.Error() != want {
		t.Errorf("got error %q, want %q", err, want)
	}
}
