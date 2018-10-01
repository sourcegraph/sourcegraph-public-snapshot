package goreman

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	wg      sync.WaitGroup
	signals = make(chan os.Signal, 10)
)

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

	p.stopped = true

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
		stopped := spawnProc(proc)
		wg.Done()
		p.mu.Unlock()
		if !stopped {
			log.Printf("%s died. Shutting down...", proc)
			signals <- syscall.SIGINT
		}
	}()
	return nil
}

// startProcs starts the processes.
func startProcs() {
	for proc := range procs {
		startProc(proc)
	}
}

// waitProcs waits for processes to complete.
func waitProcs() error {
	go func() {
		wg.Wait()
		signals <- syscall.SIGINT
	}()
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	<-signals

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
