package jsserver

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestNewPool_fail(t *testing.T) {
	p := NewPool([]byte("my!syntax(error"), 1)
	if _, err := p.Call(context.Background(), json.RawMessage(`""`)); err == nil {
		t.Fatal("want error")
	}
}

const js = `
process.stdout.write("\"ready\"\n");
process.stdin
	.on("data", function(data) {
		console.log(JSON.stringify("got " + JSON.parse(data)));
	});
`

func TestPool(t *testing.T) {
	p := NewPool([]byte(js), 2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := p.Call(ctx, json.RawMessage(`"hello"`))
	if err != nil {
		t.Fatal(err)
	}
	if want := `got hello`; string(resp) != want {
		t.Errorf("got %q, want %q", resp, want)
	}
}

func TestPool_parallel(t *testing.T) {
	p := NewPool([]byte(js), 2)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := p.Call(ctx, json.RawMessage(`"hello"`))
			if err != nil {
				t.Fatal(err)
			}
			if want := `got hello`; string(resp) != want {
				t.Errorf("got %q, want %q", resp, want)
			}
		}(i)
	}
	wg.Wait()
}

// TestPool_closeServer ensures that the pool restarts failed
// servers. Each failed server may cause one call to Call to fail, but
// it should not cause subsequent or cascading failures.
func TestPool_closeServer(t *testing.T) {
	p := NewPool([]byte(js), 2)

	var mu sync.Mutex
	numCloses := 0
	numErrs := 0

	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if i%3 == 0 {
				p := p.(*pool)

				// Close a server and make sure we recover.
				closed := false
				for s := range p.servers {
					if s != nil && !s.(*server).closed {
						closed = true
						if err := s.Close(); err != nil {
							t.Fatal(err)
						}
						if i%6 == 0 {
							s = nil
						} else {
							mu.Lock()
							numCloses++
							mu.Unlock()
						}
					}
					p.servers <- s
					if closed {
						break
					}
				}
				if !closed {
					t.Fatal("no server closed")
				}
			}

			resp, err := p.Call(ctx, json.RawMessage(`"hello"`))
			if err != nil {
				t.Logf("Call failed (1 failure per Close call is expected, so if this overall test passes, this error is expected): %s", err)
				mu.Lock()
				numErrs++
				mu.Unlock()
			} else if want := `got hello`; string(resp) != want {
				t.Errorf("got %q, want %q", resp, want)
			}
		}(i)
	}
	wg.Wait()

	if numCloses < numErrs {
		t.Errorf("got %d closes and %d errors, want closes < errors", numCloses, numErrs)
	}
}
