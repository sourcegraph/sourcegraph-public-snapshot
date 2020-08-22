package goroutine

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// BackgroundRoutine represents a component of a binary that consists of a long
// running process and a graceful shutdown mechanism.
type BackgroundRoutine interface {
	// Start begins the long-running process. The Stop method should signal to
	// this process that that application is beginnign to shut down.
	Start()

	// Stop signals the Start method to stop accepting new work and complete its
	// current work. This method can but is not required to block until Start has
	// returned.
	Stop()
}

// MonitorBackgroundRoutines will start the given background routines in their own
// (safe) goroutine (via this package's Go method). If a signal is received, the
// Stop method of each routine will be called. This method unblocks once both Start
// and Stop methods of each routine has returned. A second signal will cause the
// app to shutdown immediately.
func MonitorBackgroundRoutines(routines ...BackgroundRoutine) {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP)
	monitorBackgroundRoutines(signals, routines...)
}

func monitorBackgroundRoutines(signals <-chan os.Signal, routines ...BackgroundRoutine) {
	var wg sync.WaitGroup
	startAll(&wg, routines...)
	waitForSignal(signals)
	stopAll(&wg, routines...)
	wg.Wait()
}

// startAll calls each routine's Start method in its own goroutine and and registers
// each running goroutine with the given waitgroup.
func startAll(wg *sync.WaitGroup, routines ...BackgroundRoutine) {
	for _, r := range routines {
		t := r
		wg.Add(1)
		Go(func() { defer wg.Done(); t.Start() })
	}
}

// stopAll calls each routine's Stop method in its own goroutine and and registers
// each running goroutine with the given waitgroup.
func stopAll(wg *sync.WaitGroup, routines ...BackgroundRoutine) {
	for _, r := range routines {
		t := r
		wg.Add(1)
		Go(func() { defer wg.Done(); t.Stop() })
	}
}

// waitForSignal blocks until either SIGINT or SIGHUP has been received. This will
// call os.Exit(0) if a second signal is received.
func waitForSignal(signals <-chan os.Signal) {
	<-signals

	go func() {
		// Insta-shutdown on a second signal
		<-signals
		os.Exit(0)
	}()
}
