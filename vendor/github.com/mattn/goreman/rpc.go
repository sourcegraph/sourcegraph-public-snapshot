package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// Goreman is RPC server
type Goreman struct {
	rpcChan chan<- *rpcMessage
}

type rpcMessage struct {
	Msg  string
	Args []string
	// sending error (if any) when the task completes
	ErrCh chan error
}

// Start do start
func (r *Goreman) Start(args []string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	for _, arg := range args {
		if err = startProc(arg, nil, nil); err != nil {
			break
		}
	}
	return err
}

// Stop do stop
func (r *Goreman) Stop(args []string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	errChan := make(chan error, 1)
	r.rpcChan <- &rpcMessage{
		Msg:   "stop",
		Args:  args,
		ErrCh: errChan,
	}
	err = <-errChan
	return
}

// StopAll do stop all
func (r *Goreman) StopAll(args []string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	for proc := range procs {
		if err = stopProc(proc, nil); err != nil {
			break
		}
	}
	return err
}

// Restart do restart
func (r *Goreman) Restart(args []string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	for _, arg := range args {
		if err = restartProc(arg); err != nil {
			break
		}
	}
	return err
}

// RestartAll do restart all
func (r *Goreman) RestartAll(args []string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	for proc := range procs {
		if err = restartProc(proc); err != nil {
			break
		}
	}
	return err
}

// List do list
func (r *Goreman) List(args []string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	*ret = ""
	for proc := range procs {
		*ret += proc + "\n"
	}
	return err
}

// Status do status
func (r *Goreman) Status(args []string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	*ret = ""
	for proc := range procs {
		if procs[proc].cmd != nil {
			*ret += "*" + proc + "\n"
		} else {
			*ret += " " + proc + "\n"
		}
	}
	return err
}

// command: run.
func run(cmd string, args []string, serverPort uint) error {
	client, err := rpc.Dial("tcp", defaultServer(serverPort))
	if err != nil {
		return err
	}
	defer client.Close()
	var ret string
	switch cmd {
	case "start":
		return client.Call("Goreman.Start", args, &ret)
	case "stop":
		return client.Call("Goreman.Stop", args, &ret)
	case "stop-all":
		return client.Call("Goreman.StopAll", args, &ret)
	case "restart":
		return client.Call("Goreman.Restart", args, &ret)
	case "restart-all":
		return client.Call("Goreman.RestartAll", args, &ret)
	case "list":
		err := client.Call("Goreman.List", args, &ret)
		fmt.Print(ret)
		return err
	case "status":
		err := client.Call("Goreman.Status", args, &ret)
		fmt.Print(ret)
		return err
	}
	return errors.New("unknown command")
}

// start rpc server.
func startServer(ctx context.Context, rpcChan chan<- *rpcMessage, listenPort uint) error {
	gm := &Goreman{
		rpcChan: rpcChan,
	}
	rpc.Register(gm)
	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", defaultAddr(), listenPort))
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var acceptingConns = true
	for acceptingConns {
		conns := make(chan net.Conn, 1)
		go func() {
			conn, err := server.Accept()
			if err != nil {
				return
			}
			conns <- conn
		}()
		select {
		case <-ctx.Done():
			acceptingConns = false
			break
		case client := <-conns: // server is not canceled.
			wg.Add(1)
			go func() {
				defer wg.Done()
				rpc.ServeConn(client)
			}()
		}
	}
	done := make(chan struct{}, 1)
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	select {
	case <-done:
		return nil
	case <-time.After(10 * time.Second):
		return errors.New("RPC server did not shut down in 10 seconds, quitting")
	}
}
