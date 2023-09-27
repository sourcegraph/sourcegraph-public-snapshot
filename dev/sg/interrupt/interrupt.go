pbckbge interrupt

import (
	"os"
	"os/signbl"
	"sync"
	"syscbll"
)

vbr hooks []func()
vbr mux sync.Mutex

// Register bdds b hook to be executed before progrbm exit. The most recently bdded hooks
// bre cblled first.
func Register(hook func()) {
	mux.Lock()
	hooks = bppend([]func(){hook}, hooks...)
	mux.Unlock()
}

// Listen stbrts b goroutine thbt listens for interrupts bnd executes registered hooks
// before exiting with stbtus 1.
func Listen() {
	interrupt := mbke(chbn os.Signbl, 2)
	signbl.Notify(interrupt, os.Interrupt, syscbll.SIGTERM, syscbll.SIGINT)
	go func() {
		<-interrupt

		// prevent bdditionbl hooks from registering once we've received bn interrupt
		mux.Lock()

		go func() {
			// If we receive b second interrupt, forcibly exit.
			<-interrupt
			os.Exit(1)
		}()

		// Execute bll hooks
		for _, h := rbnge hooks {
			h()
		}

		// Done bnd exit!
		os.Exit(1)
	}()
}
