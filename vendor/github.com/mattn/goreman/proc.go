package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

// spawnProc starts the specified proc, and returns any error from running it.
func spawnProc(proc string, errCh chan<- error) {
	procObj := procs[proc]
	logger := createLogger(proc, procObj.colorIndex)

	cs := append(cmdStart, procObj.cmdline)
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger
	cmd.SysProcAttr = procAttrs

	if procObj.setPort {
		cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", procObj.port))
		fmt.Fprintf(logger, "Starting %s on port %d\n", proc, procObj.port)
	}

	if err := cmd.Start(); err != nil {
		select {
		case errCh <- err:
		default:
		}
		fmt.Fprintf(logger, "Failed to start %s: %s\n", proc, err)
		return
	}
	procObj.cmd = cmd
	procObj.stoppedBySupervisor = false
	procObj.mu.Unlock()
	err := cmd.Wait()
	procObj.mu.Lock()
	procObj.cond.Broadcast()
	if err != nil && procObj.stoppedBySupervisor == false {
		select {
		case errCh <- err:
		default:
		}
	}
	procObj.waitErr = err
	procObj.cmd = nil
	fmt.Fprintf(logger, "Terminating %s\n", proc)
}

// Stop the specified proc, issuing os.Kill if it does not terminate within 10
// seconds. If signal is nil, os.Interrupt is used.
func stopProc(proc string, signal os.Signal) error {
	if signal == nil {
		signal = os.Interrupt
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
	p.stoppedBySupervisor = true

	err := terminateProc(proc, signal)
	if err != nil {
		return err
	}

	timeout := time.AfterFunc(10*time.Second, func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p, ok := procs[proc]; ok && p.cmd != nil {
			err = killProc(p.cmd.Process)
		}
	})
	p.cond.Wait()
	timeout.Stop()
	return err
}

// start specified proc. if proc is started already, return nil.
func startProc(proc string, wg *sync.WaitGroup, errCh chan<- error) error {
	p, ok := procs[proc]
	if !ok || p == nil {
		return errors.New("unknown proc: " + proc)
	}

	p.mu.Lock()
	if procs[proc].cmd != nil {
		p.mu.Unlock()
		return nil
	}

	if wg != nil {
		wg.Add(1)
	}
	go func() {
		spawnProc(proc, errCh)
		if wg != nil {
			wg.Done()
		}
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

	stopProc(proc, nil)
	return startProc(proc, nil, nil)
}

// stopProcs attempts to stop every running process and returns any non-nil
// error, if one exists. stopProcs will wait until all procs have had an
// opportunity to stop.
func stopProcs(sig os.Signal) error {
	var err error
	for proc := range procs {
		stopErr := stopProc(proc, sig)
		if stopErr != nil {
			err = stopErr
		}
	}
	return err
}

// spawn all procs.
func startProcs(sc <-chan os.Signal, rpcCh <-chan *rpcMessage, exitOnError bool) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for proc := range procs {
		startProc(proc, &wg, errCh)
	}

	allProcsDone := make(chan struct{}, 1)
	go func() {
		wg.Wait()
		allProcsDone <- struct{}{}
	}()
	for {
		select {
		case rpcMsg := <-rpcCh:
			switch rpcMsg.Msg {
			// TODO: add more events here.
			case "stop":
				for _, proc := range rpcMsg.Args {
					if err := stopProc(proc, nil); err != nil {
						rpcMsg.ErrCh <- err
						break
					}
				}
				close(rpcMsg.ErrCh)
			default:
				panic("unimplemented rpc message type " + rpcMsg.Msg)
			}
		case err := <-errCh:
			if exitOnError {
				stopProcs(os.Interrupt)
				return err
			}
		case <-allProcsDone:
			return stopProcs(os.Interrupt)
		case sig := <-sc:
			return stopProcs(sig)
		}
	}
	return nil
}
