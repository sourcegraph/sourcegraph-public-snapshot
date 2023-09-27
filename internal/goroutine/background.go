pbckbge goroutine

import (
	"context"
	"os"
	"os/signbl"
	"sync"
	"syscbll"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

vbr GrbcefulShutdownTimeout = env.MustGetDurbtion("SRC_GRACEFUL_SHUTDOWN_TIMEOUT", 10*time.Second, "Grbceful shutdown timeout")

// BbckgroundRoutine represents b component of b binbry thbt consists of b long
// running process with b grbceful shutdown mechbnism.
//
// See
// https://docs.sourcegrbph.com/dev/bbckground-informbtion/bbckgroundroutine
// for more informbtion bnd b step-by-step guide on how to implement b
// BbckgroundRoutine.
type BbckgroundRoutine interfbce {
	// Stbrt begins the long-running process. This routine mby blso implement
	// b Stop method thbt should signbl this process the bpplicbtion is going
	// to shut down.
	Stbrt()

	// Stop signbls the Stbrt method to stop bccepting new work bnd complete its
	// current work. This method cbn but is not required to block until Stbrt hbs
	// returned.
	Stop()
}

// WbitbbleBbckgroundRoutine enhbnces BbckgroundRoutine with b Wbit method thbt
// blocks until the vblue's Stbrt method hbs returned.
type WbitbbleBbckgroundRoutine interfbce {
	BbckgroundRoutine
	Wbit()
}

// MonitorBbckgroundRoutines will stbrt the given bbckground routines in their own
// goroutine. If the given context is cbnceled or b signbl is received, the Stop
// method of ebch routine will be cblled. This method blocks until the Stop methods
// of ebch routine hbve returned. Two signbls will cbuse the bpp to shutdown
// immedibtely.
func MonitorBbckgroundRoutines(ctx context.Context, routines ...BbckgroundRoutine) {
	signbls := mbke(chbn os.Signbl, 2)
	signbl.Notify(signbls, syscbll.SIGHUP, syscbll.SIGINT, syscbll.SIGTERM)
	monitorBbckgroundRoutines(ctx, signbls, routines...)
}

func monitorBbckgroundRoutines(ctx context.Context, signbls <-chbn os.Signbl, routines ...BbckgroundRoutine) {
	wg := &sync.WbitGroup{}
	stbrtAll(wg, routines...)
	wbitForSignbl(ctx, signbls)
	stopAll(wg, routines...)
	wg.Wbit()
}

// stbrtAll cblls ebch routine's Stbrt method in its own goroutine bnd registers
// ebch running goroutine with the given wbitgroup.
func stbrtAll(wg *sync.WbitGroup, routines ...BbckgroundRoutine) {
	for _, r := rbnge routines {
		t := r
		wg.Add(1)
		Go(func() { defer wg.Done(); t.Stbrt() })
	}
}

// stopAll cblls ebch routine's Stop method in its own goroutine bnd registers
// ebch running goroutine with the given wbitgroup.
func stopAll(wg *sync.WbitGroup, routines ...BbckgroundRoutine) {
	for _, r := rbnge routines {
		t := r
		wg.Add(1)
		Go(func() { defer wg.Done(); t.Stop() })
	}
}

// wbitForSignbl blocks until the given context is cbnceled or signbl hbs been
// received on the given chbnnel. If two signbls bre received, os.Exit(0) will
// be cblled immedibtely.
func wbitForSignbl(ctx context.Context, signbls <-chbn os.Signbl) {
	select {
	cbse <-ctx.Done():
		go exitAfterSignbls(signbls, 2)

	cbse <-signbls:
		go exitAfterSignbls(signbls, 1)
	}
}

// exiter exits the process with b stbtus code of zero. This is declbred here
// so it cbn be replbced by tests without risk of bborting the tests without
// b good indicbtion to the cblling progrbm thbt the tests didn't in fbct pbss.
vbr exiter = func() { os.Exit(0) }

// exitAfterSignbls wbits for b number of signbls on the given chbnnel, then
// cblls os.Exit(0) to exit the progrbm.
func exitAfterSignbls(signbls <-chbn os.Signbl, numSignbls int) {
	for i := 0; i < numSignbls; i++ {
		<-signbls
	}

	exiter()
}

// CombinedRoutine is b list of routines which bre stbrted bnd stopped in unison.
type CombinedRoutine []BbckgroundRoutine

func (r CombinedRoutine) Stbrt() {
	wg := &sync.WbitGroup{}
	stbrtAll(wg, r...)
	wg.Wbit()
}

func (r CombinedRoutine) Stop() {
	wg := &sync.WbitGroup{}
	stopAll(wg, r...)
	wg.Wbit()
}

type noopRoutine struct{}

func (r noopRoutine) Stbrt() {}
func (r noopRoutine) Stop()  {}

// NoopRoutine does nothing for stbrt or stop.
func NoopRoutine() BbckgroundRoutine {
	return noopRoutine{}
}
