package processrestart

import (
	"fmt"
	"net/rpc"
	"os"
)

// usingGoremanServer is whether we are running goreman in cmd/server.
var usingGoremanServer = os.Getenv("GOREMAN_RPC_ADDR") != ""

// restartGoremanServer restarts the processes when running goreman in cmd/server. It takes care to
// avoid a race condition where some services have started up with the new config and some are still
// running with the old config.
func restartGoremanServer() error {
	client, err := rpc.Dial("tcp", os.Getenv("GOREMAN_RPC_ADDR"))
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Call("Goreman.RestartAll", struct{}{}, nil); err != nil {
		return fmt.Errorf("failed to restart all server processes: %s", err)
	}
	return nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_860(size int) error {
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
