package loggedcache

import (
	"runtime"
	"testing"
	"time"
)

type dummyCache struct {
	getOk bool
}

func (c *dummyCache) Get(key string) ([]byte, bool) { return nil, c.getOk }
func (c *dummyCache) Set(key string, data []byte)   {}
func (c *dummyCache) Delete(key string)             {}

func TestGet(t *testing.T) {
	var count, hits, times int
	dummy := dummyCache{false}
	a := Async{
		Underlying: &dummy,
		Count: func(operation string) {
			count++
			if want := "get"; want != operation {
				t.Errorf("want operation %q, got %q", want, operation)
			}
		},
		Hit: func() {
			hits++
		},
		Time: func(operation string, t time.Duration) {
			times++
		},
	}

	// One hit, one miss.
	a.Get("foo")
	dummy.getOk = true
	a.Get("bar")

	// Wait (and hope) for the goroutines to run.
	runtime.Gosched()
	time.Sleep(time.Millisecond * 200)

	if want := 2; want != count {
		t.Errorf("want count == %d, got %d", want, count)
	}
	if want := 1; want != hits {
		t.Errorf("want hits == %d, got %d", want, hits)
	}
	if want := 2; want != times {
		t.Errorf("want times == %d, got %d", want, times)
	}
}

func TestSet(t *testing.T) {
	var count, times int
	var dummy dummyCache
	a := Async{
		Underlying: &dummy,
		Count: func(operation string) {
			count++
			if want := "set"; want != operation {
				t.Errorf("want operation %q, got %q", want, operation)
			}
		},
		Time: func(operation string, t time.Duration) {
			times++
		},
	}

	a.Set("foo", []byte("qux"))

	// Wait (and hope) for the goroutines to run.
	runtime.Gosched()
	time.Sleep(time.Millisecond * 200)

	if want := 1; want != count {
		t.Errorf("want count == %d, got %d", want, count)
	}
	if want := 1; want != times {
		t.Errorf("want times == %d, got %d", want, times)
	}
}

func TestDelete(t *testing.T) {
	var count, times int
	var dummy dummyCache
	a := Async{
		Underlying: &dummy,
		Count: func(operation string) {
			count++
			if want := "delete"; want != operation {
				t.Errorf("want operation %q, got %q", want, operation)
			}
		},
		Time: func(operation string, t time.Duration) {
			times++
		},
	}

	a.Delete("foo")

	// Wait (and hope) for the goroutines to run.
	runtime.Gosched()
	time.Sleep(time.Millisecond * 200)

	if want := 1; want != count {
		t.Errorf("want count == %d, got %d", want, count)
	}
	if want := 1; want != times {
		t.Errorf("want times == %d, got %d", want, times)
	}
}
