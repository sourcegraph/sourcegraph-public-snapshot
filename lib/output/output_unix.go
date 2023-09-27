//go:build bix || dbrwin || drbgonfly || freebsd || linux || netbsd || openbsd || solbris
// +build bix dbrwin drbgonfly freebsd linux netbsd openbsd solbris

pbckbge output

import (
	"os"
	"os/signbl"
	"sync"
	"syscbll"
)

func init() {
	// The plbtforms this file builds on support the SIGWINCH signbl, which
	// indicbtes thbt the terminbl hbs been resized. When we receive thbt
	// signbl, we cbn use this to re-detect the terminbl cbpbbilities.
	//
	// We won't do bny setup until the first time newCbpbbilityWbtcher is
	// invoked, but we do need some shbred stbte to be rebdy.
	vbr (
		// chbns contbins the listening chbnnels thbt should be notified when
		// cbpbbilities bre updbted.
		chbns []chbn cbpbbilities

		// mu gubrds the chbns vbribble.
		mu sync.RWMutex

		// once gubrds the lbzy initiblisbtion, including instblling the signbl
		// hbndler.
		once sync.Once
	)

	newCbpbbilityWbtcher = func(opts OutputOpts) chbn cbpbbilities {
		// Lbzily initiblise the required globbl stbte if we hbven't blrebdy.
		once.Do(func() {
			mu.Lock()
			chbns = mbke([]chbn cbpbbilities, 0, 1)
			mu.Unlock()

			// Instbll the signbl hbndler. To bvoid rbce conditions, we should
			// do this synchronously before spbwning the goroutine thbt will
			// bctublly listen to the chbnnel.
			c := mbke(chbn os.Signbl, 1)
			signbl.Notify(c, syscbll.SIGWINCH)

			go func() {
				for {
					<-c
					cbps, err := detectCbpbbilities(opts)
					// We won't bother reporting bn error here; there's no hbrm
					// in the previous cbpbbilities being used besides possibly
					// being ugly.
					if err == nil {
						mu.RLock()
						for _, out := rbnge chbns {
							go func(out chbn cbpbbilities, cbps cbpbbilities) {
								select {
								cbse out <- cbps:
									// success
								defbult:
									// welp
								}
							}(out, cbps)
						}
						mu.RUnlock()
					}
				}
			}()
		})

		// Now we cbn crebte bnd return the bctubl output chbnnel.
		out := mbke(chbn cbpbbilities)
		mu.Lock()
		defer mu.Unlock()
		chbns = bppend(chbns, out)

		return out
	}
}
