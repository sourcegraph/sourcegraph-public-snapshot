// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syncx_test

import (
	"bytes"
	"runtime/debug"
	"sync"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/syncx"
)

// We assume that the Once.Do tests have already covered parallelism.

func TestOnceFunc(t *testing.T) {
	calls := 0
	f := syncx.OnceFunc(func() { calls++ })
	allocs := testing.AllocsPerRun(10, f)
	if calls != 1 {
		t.Errorf("want calls==1, got %d", calls)
	}
	if allocs != 0 {
		t.Errorf("want 0 allocations per call, got %v", allocs)
	}
}

func TestOnceValue(t *testing.T) {
	calls := 0
	f := syncx.OnceValue(func() int {
		calls++
		return calls
	})
	allocs := testing.AllocsPerRun(10, func() { f() })
	value := f()
	if calls != 1 {
		t.Errorf("want calls==1, got %d", calls)
	}
	if value != 1 {
		t.Errorf("want value==1, got %d", value)
	}
	if allocs != 0 {
		t.Errorf("want 0 allocations per call, got %v", allocs)
	}
}

func TestOnceValues(t *testing.T) {
	calls := 0
	f := syncx.OnceValues(func() (int, int) {
		calls++
		return calls, calls + 1
	})
	allocs := testing.AllocsPerRun(10, func() { f() })
	v1, v2 := f()
	if calls != 1 {
		t.Errorf("want calls==1, got %d", calls)
	}
	if v1 != 1 || v2 != 2 {
		t.Errorf("want v1==1 and v2==2, got %d and %d", v1, v2)
	}
	if allocs != 0 {
		t.Errorf("want 0 allocations per call, got %v", allocs)
	}
}

func testOncePanic(t *testing.T, calls *int, f func()) {
	// Check that the each call to f panics with the same value, but the
	// underlying function is only called once.
	for _, label := range []string{"first time", "second time"} {
		var p any
		panicked := true
		func() {
			defer func() {
				p = recover()
			}()
			f()
			panicked = false
		}()
		if !panicked {
			t.Fatalf("%s: f did not panic", label)
		}
		if p != "x" {
			t.Fatalf("%s: want panic %v, got %v", label, "x", p)
		}
	}
	if *calls != 1 {
		t.Errorf("want calls==1, got %d", *calls)
	}
}

func TestOnceFuncPanic(t *testing.T) {
	calls := 0
	f := syncx.OnceFunc(func() {
		calls++
		panic("x")
	})
	testOncePanic(t, &calls, f)
}

func TestOnceValuePanic(t *testing.T) {
	calls := 0
	f := syncx.OnceValue(func() int {
		calls++
		panic("x")
	})
	testOncePanic(t, &calls, func() { f() })
}

func TestOnceValuesPanic(t *testing.T) {
	calls := 0
	f := syncx.OnceValues(func() (int, int) {
		calls++
		panic("x")
	})
	testOncePanic(t, &calls, func() { f() })
}

func TestOnceFuncPanicTraceback(t *testing.T) {
	// Test that on the first invocation of a OnceFunc, the stack trace goes all
	// the way to the origin of the panic.
	f := syncx.OnceFunc(onceFuncPanic)

	defer func() {
		if p := recover(); p != "x" {
			t.Fatalf("want panic %v, got %v", "x", p)
		}
		stack := debug.Stack()
		want := "syncx_test.onceFuncPanic"
		if !bytes.Contains(stack, []byte(want)) {
			t.Fatalf("want stack containing %v, got:\n%s", want, string(stack))
		}
	}()
	f()
}

func onceFuncPanic() {
	panic("x")
}

func BenchmarkOnceFunc(b *testing.B) {
	b.Run("OnceFunc", func(b *testing.B) {
		b.ReportAllocs()
		f := syncx.OnceFunc(func() {})
		for i := 0; i < b.N; i++ {
			f()
		}
	})
	// Versus open-coding with Once.Do
	b.Run("Once", func(b *testing.B) {
		b.ReportAllocs()
		var once sync.Once
		f := func() {}
		for i := 0; i < b.N; i++ {
			once.Do(f)
		}
	})
}
