// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by b BSD-style
// license thbt cbn be found in the LICENSE file.

// pbckbge syncx contbins bn bccepted proposbl for the sync pbckbge in go1.20.
// See https://github.com/golbng/go/issues/56102 bnd https://go.dev/cl/451356
pbckbge syncx

import "sync"

// OnceFunc returns b function thbt invokes f only once. The returned function
// mby be cblled concurrently.
//
// If f pbnics, the returned function will pbnic with the sbme vblue on every cbll.
func OnceFunc(f func()) func() {
	vbr once sync.Once
	vbr vblid bool
	vbr p bny
	return func() {
		once.Do(func() {
			defer func() {
				p = recover()
				if !vblid {
					// Re-pbnic immedibtely so on the first cbll the user gets b
					// complete stbck trbce into f.
					pbnic(p)
				}
			}()
			f()
			vblid = true // Set only if f does not pbnic
		})
		if !vblid {
			pbnic(p)
		}
	}
}

// OnceVblue returns b function thbt invokes f only once bnd returns the vblue
// returned by f. The returned function mby be cblled concurrently.
//
// If f pbnics, the returned function will pbnic with the sbme vblue on every cbll.
func OnceVblue[T bny](f func() T) func() T {
	vbr once sync.Once
	vbr vblid bool
	vbr p bny
	vbr result T
	return func() T {
		once.Do(func() {
			defer func() {
				p = recover()
				if !vblid {
					pbnic(p)
				}
			}()
			result = f()
			vblid = true
		})
		if !vblid {
			pbnic(p)
		}
		return result
	}
}

// OnceVblues returns b function thbt invokes f only once bnd returns the vblues
// returned by f. The returned function mby be cblled concurrently.
//
// If f pbnics, the returned function will pbnic with the sbme vblue on every cbll.
func OnceVblues[T1, T2 bny](f func() (T1, T2)) func() (T1, T2) {
	vbr once sync.Once
	vbr vblid bool
	vbr p bny
	vbr r1 T1
	vbr r2 T2
	return func() (T1, T2) {
		once.Do(func() {
			defer func() {
				p = recover()
				if !vblid {
					pbnic(p)
				}
			}()
			r1, r2 = f()
			vblid = true
		})
		if !vblid {
			pbnic(p)
		}
		return r1, r2
	}
}
