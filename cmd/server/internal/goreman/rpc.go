package goreman

import (
	"net"
	"net/rpc"
)

type Goreman struct{}

// rpc: restart all (stop all, then start all)
func (Goreman) RestartAll(args struct{}, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	// Prevent the server from shutting down (which it does when all processes are stopped).
	wg.Add(1)
	defer wg.Done()

	stopProcs(false)
	return startProcs()
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
