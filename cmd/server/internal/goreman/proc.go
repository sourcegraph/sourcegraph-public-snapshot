package goreman

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var wg sync.WaitGroup

// stop specified proc.
func stopProc(proc string, kill bool) error {
	p, ok := procs[proc]
	if !ok {
		return errors.New("Unknown proc: " + proc)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil {
		return nil
	}

	p.quit = true

	if kill {
		err := p.cmd.Process.Kill()
		if err != nil {
			return err
		}
	} else {
		err := terminateProc(proc)
		if err != nil {
			return err
		}
	}

	p.cond.Wait()
	return nil
}

// start specified proc. if proc is started already, return nil.
func startProc(proc string) error {
	p, ok := procs[proc]
	if !ok {
		return errors.New("Unknown proc: " + proc)
	}

	p.mu.Lock()
	if procs[proc].cmd != nil {
		p.mu.Unlock()
		return nil
	}

	wg.Add(1)
	go func() {
		spawnProc(proc)
		wg.Done()
		p.mu.Unlock()
	}()
	return nil
}

// restart specified proc.
func restartProc(proc string) error {
	if _, ok := procs[proc]; !ok {
		return errors.New("Unknown proc: " + proc)
	}
	stopProc(proc, false)
	return startProc(proc)
}

// spawn all procs.
func startProcs() error {
	for proc := range procs {
		startProc(proc)
	}
	sc := make(chan os.Signal, 10)
	go func() {
		wg.Wait()
		sc <- syscall.SIGINT
	}()
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	<-sc

	stopped := make(chan struct{})
	go func() {
		stopProcs(false)
		close(stopped)
	}()

	// New signal chan to avoid built up buffered signals
	sc2 := make(chan os.Signal, 10)
	signal.Notify(sc2, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	select {
	case <-sc2:
		// Second signal received, do a hard exit
		stopProcs(true)
	case <-time.NewTimer(10 * time.Second).C:
		// 10 seconds has passed, kill
		stopProcs(true)
	case <-stopped:
		// Happy case, just continue
	}

	return nil
}

func stopProcs(kill bool) {
	// TODO we probably need a well defined order for shutting down, since
	// something may want to finish writing to postgres for example.
	var wg sync.WaitGroup
	for proc := range procs {
		wg.Add(1)
		go func(proc string) {
			defer wg.Done()
			stopProc(proc, kill)
		}(proc)
	}
	wg.Wait()
}
