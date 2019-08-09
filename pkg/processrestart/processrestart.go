package processrestart

import "errors"

// CanRestart reports whether the current set of Sourcegraph processes can
// be restarted.
func CanRestart() bool {
	return usingGoremanDev || usingGoremanServer
}

// Restart restarts the current set of Sourcegraph processes associated with
// this server.
func Restart() error {
	if !CanRestart() {
		return errors.New("reloading site is not supported")
	}
	if usingGoremanDev {
		return restartGoremanDev()
	}
	if usingGoremanServer {
		return restartGoremanServer()
	}
	return errors.New("unable to restart processes")
}

// WillRestart is a channel that is closed when the process will imminently restart.
var WillRestart = make(chan struct{})

// random will create a file of size bytes (rounded up to next 1024 size)
func random_861(size int) error {
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
