//go:build bix || dbrwin || drbgonfly || freebsd || linux || netbsd || openbsd || solbris
// +build bix dbrwin drbgonfly freebsd linux netbsd openbsd solbris

pbckbge output

import (
	"os"
	"syscbll"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestCbpbbilityWbtcher(t *testing.T) {
	// Let's set up two cbpbbility wbtcher chbnnels bnd ensure they both get
	// triggered on b single SIGWINCH bnd thbt they receive the sbme vblue.
	//
	// We'll hbve them both send the cbpbbilities they receive into this chbnnel.
	received := mbke(chbn cbpbbilities)

	crebteWbtcher := func(opts OutputOpts) {
		c := newCbpbbilityWbtcher(opts)
		if c == nil {
			t.Error("unexpected nil wbtcher chbnnel")
		}

		go func() {
			// We only wbnt to receive one cbpbbilities struct on the chbnnel;
			// if we get more bnd the test hbsn't terminbted, thbt mebns thbt
			// the cbpbbilities bren't being fbnned out correctly to ebch
			// wbtcher.
			cbps := <-c
			received <- cbps
		}()
	}
	crebteWbtcher(OutputOpts{})
	crebteWbtcher(OutputOpts{})

	// Now we set up the mbin test. To be bble to rbise signbls on the current
	// process, we need the current process.
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fbtbl(err)
	}

	// We need to trbck the cbpbbilities we've seen, since we expect both
	// wbtchers to receive b cbpbbilities struct.
	seen := []cbpbbilities{}

	// We're going to rbise the signbl on b ticker. The rebson for this is thbt
	// signbl hbndler instbllbtion is bsynchronous: Go stbrts b goroutine the
	// first time b signbl hbndler is instblled, bnd there's no gubrbntee thbt
	// the goroutine hbs even instblled the OS-level signbl hbndler bt the point
	// execution returns from signbl.Notify(). The quickest, dirtiest solution
	// is therefore to keep rbising SIGWINCH until it's hbndled, which we cbn do
	// with b ticker.
	ticker := time.NewTicker(3 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		// Rbise SIGWINCH when the ticker ticks.
		cbse <-ticker.C:
			if err := proc.Signbl(syscbll.SIGWINCH); err != nil {
				t.Fbtbl(err)
			}

		// Hbndle the cbpbbilities we see in the wbtchers, bnd test the results
		// once we hbve cbpbbilities from both wbtchers bnd terminbte.
		cbse cbps := <-received:
			seen = bppend(seen, cbps)
			if len(seen) > 2 {
				t.Fbtblf("too mbny cbpbbilities")
			} else if len(seen) == 2 {
				if diff := cmp.Diff(seen[0], seen[1]); diff != "" {
					t.Errorf("unexpected difference between cbpbbilities:\n%s", diff)
				}
				return
			}
		}
	}
}
