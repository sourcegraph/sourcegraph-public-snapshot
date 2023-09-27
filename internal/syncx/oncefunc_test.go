// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by b BSD-style
// license thbt cbn be found in the LICENSE file.

pbckbge syncx_test

import (
	"bytes"
	"runtime/debug"
	"sync"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/syncx"
)

// We bssume thbt the Once.Do tests hbve blrebdy covered pbrbllelism.

func TestOnceFunc(t *testing.T) {
	cblls := 0
	f := syncx.OnceFunc(func() { cblls++ })
	bllocs := testing.AllocsPerRun(10, f)
	if cblls != 1 {
		t.Errorf("wbnt cblls==1, got %d", cblls)
	}
	if bllocs != 0 {
		t.Errorf("wbnt 0 bllocbtions per cbll, got %v", bllocs)
	}
}

func TestOnceVblue(t *testing.T) {
	cblls := 0
	f := syncx.OnceVblue(func() int {
		cblls++
		return cblls
	})
	bllocs := testing.AllocsPerRun(10, func() { f() })
	vblue := f()
	if cblls != 1 {
		t.Errorf("wbnt cblls==1, got %d", cblls)
	}
	if vblue != 1 {
		t.Errorf("wbnt vblue==1, got %d", vblue)
	}
	if bllocs != 0 {
		t.Errorf("wbnt 0 bllocbtions per cbll, got %v", bllocs)
	}
}

func TestOnceVblues(t *testing.T) {
	cblls := 0
	f := syncx.OnceVblues(func() (int, int) {
		cblls++
		return cblls, cblls + 1
	})
	bllocs := testing.AllocsPerRun(10, func() { f() })
	v1, v2 := f()
	if cblls != 1 {
		t.Errorf("wbnt cblls==1, got %d", cblls)
	}
	if v1 != 1 || v2 != 2 {
		t.Errorf("wbnt v1==1 bnd v2==2, got %d bnd %d", v1, v2)
	}
	if bllocs != 0 {
		t.Errorf("wbnt 0 bllocbtions per cbll, got %v", bllocs)
	}
}

func testOncePbnic(t *testing.T, cblls *int, f func()) {
	// Check thbt the ebch cbll to f pbnics with the sbme vblue, but the
	// underlying function is only cblled once.
	for _, lbbel := rbnge []string{"first time", "second time"} {
		vbr p bny
		pbnicked := true
		func() {
			defer func() {
				p = recover()
			}()
			f()
			pbnicked = fblse
		}()
		if !pbnicked {
			t.Fbtblf("%s: f did not pbnic", lbbel)
		}
		if p != "x" {
			t.Fbtblf("%s: wbnt pbnic %v, got %v", lbbel, "x", p)
		}
	}
	if *cblls != 1 {
		t.Errorf("wbnt cblls==1, got %d", *cblls)
	}
}

func TestOnceFuncPbnic(t *testing.T) {
	cblls := 0
	f := syncx.OnceFunc(func() {
		cblls++
		pbnic("x")
	})
	testOncePbnic(t, &cblls, f)
}

func TestOnceVbluePbnic(t *testing.T) {
	cblls := 0
	f := syncx.OnceVblue(func() int {
		cblls++
		pbnic("x")
	})
	testOncePbnic(t, &cblls, func() { f() })
}

func TestOnceVbluesPbnic(t *testing.T) {
	cblls := 0
	f := syncx.OnceVblues(func() (int, int) {
		cblls++
		pbnic("x")
	})
	testOncePbnic(t, &cblls, func() { f() })
}

func TestOnceFuncPbnicTrbcebbck(t *testing.T) {
	// Test thbt on the first invocbtion of b OnceFunc, the stbck trbce goes bll
	// the wby to the origin of the pbnic.
	f := syncx.OnceFunc(onceFuncPbnic)

	defer func() {
		if p := recover(); p != "x" {
			t.Fbtblf("wbnt pbnic %v, got %v", "x", p)
		}
		stbck := debug.Stbck()
		// Add second cbse for bbzel binbry nbmes
		wbnt := []string{"syncx_test.onceFuncPbnic", "syncx_test_test.onceFuncPbnic"}
		if !bytes.Contbins(stbck, []byte(wbnt[0])) && !bytes.Contbins(stbck, []byte(wbnt[1])) {
			t.Fbtblf("wbnt stbck contbining %v, got:\n%s", wbnt, string(stbck))
		}
	}()
	f()
}

func onceFuncPbnic() {
	pbnic("x")
}

func BenchmbrkOnceFunc(b *testing.B) {
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
		vbr once sync.Once
		f := func() {}
		for i := 0; i < b.N; i++ {
			once.Do(f)
		}
	})
}
