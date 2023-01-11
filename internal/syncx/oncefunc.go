// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package syncx contains an accepted proposal for the sync package in go1.20.
// See https://github.com/golang/go/issues/56102 and https://go.dev/cl/451356
package syncx

import "sync"

// OnceFunc returns a function that invokes f only once. The returned function
// may be called concurrently.
//
// If f panics, the returned function will panic with the same value on every call.
func OnceFunc(f func()) func() {
	var once sync.Once
	var valid bool
	var p any
	return func() {
		once.Do(func() {
			defer func() {
				p = recover()
				if !valid {
					// Re-panic immediately so on the first call the user gets a
					// complete stack trace into f.
					panic(p)
				}
			}()
			f()
			valid = true // Set only if f does not panic
		})
		if !valid {
			panic(p)
		}
	}
}

// OnceValue returns a function that invokes f only once and returns the value
// returned by f. The returned function may be called concurrently.
//
// If f panics, the returned function will panic with the same value on every call.
func OnceValue[T any](f func() T) func() T {
	var once sync.Once
	var valid bool
	var p any
	var result T
	return func() T {
		once.Do(func() {
			defer func() {
				p = recover()
				if !valid {
					panic(p)
				}
			}()
			result = f()
			valid = true
		})
		if !valid {
			panic(p)
		}
		return result
	}
}

// OnceValues returns a function that invokes f only once and returns the values
// returned by f. The returned function may be called concurrently.
//
// If f panics, the returned function will panic with the same value on every call.
func OnceValues[T1, T2 any](f func() (T1, T2)) func() (T1, T2) {
	var once sync.Once
	var valid bool
	var p any
	var r1 T1
	var r2 T2
	return func() (T1, T2) {
		once.Do(func() {
			defer func() {
				p = recover()
				if !valid {
					panic(p)
				}
			}()
			r1, r2 = f()
			valid = true
		})
		if !valid {
			panic(p)
		}
		return r1, r2
	}
}
