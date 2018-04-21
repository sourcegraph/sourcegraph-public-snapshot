package main

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
// if signal is nil, SIGTERM is used
func stopProc(proc string, quit bool, signal os.Signal) error {
	if signal == nil {
		signal = syscall.SIGTERM
	}
	p, ok := procs[proc]
	if !ok || p == nil {
		return errors.New("unknown proc: " + proc)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil {
		return nil
	}

	p.quit = quit
	err := terminateProc(proc, signal)
	if err != nil {
		return err
	}

	timeout := time.AfterFunc(10*time.Second, func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p, ok := procs[proc]; ok && p.cmd != nil {
			err = p.cmd.Process.Kill()
		}
	})
	p.cond.Wait()
	timeout.Stop()
	return err
}

// start specified proc. if proc is started already, return nil.
func startProc(proc string) error {
	p, ok := procs[proc]
	if !ok || p == nil {
		return errors.New("unknown proc: " + proc)
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
	p, ok := procs[proc]
	if !ok || p == nil {
		return errors.New("unknown proc: " + proc)
	}

	stopProc(proc, false, nil)
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
	signal := <-sc
	for proc := range procs {
		stopProc(proc, true, signal)
	}
	return nil
}
