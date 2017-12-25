package goreman

import (
	"net"
	"net/rpc"
)

type Goreman struct{}

// rpc: restart
func (Goreman) Restart(proc string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return restartProc(proc)
}

// start rpc server.
func startServer(addr string) error {
	rpc.Register(Goreman{})
	server, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	go func() {
		for {
			client, err := server.Accept()
			if err != nil {
				continue
			}
			rpc.ServeConn(client)
		}
	}()
	return nil
}
