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

	// Stop and start the processes. We do this with an artificially
	// incremented wg, so that the server does not shutdown when stopProcs
	// completes (the server shuts down when all processes are stopped).
	//
	// Note that we do not invoke waitProcs, as the original waitProcs
	// invocation is still running and is still valid.
	wg.Add(1)
	defer wg.Done()
	stopProcs(false)
	startProcs()

	return nil
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_532(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
