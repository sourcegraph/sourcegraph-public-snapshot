// Command goremancmd exists for testing the internally vendored goreman that
// ./cmd/server uses.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/server/internal/goreman"
)

func do() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("USAGE: %s Procfile", os.Args[0])
	}

	procfile, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		return err
	}

	const goremanAddr = "127.0.0.1:5005"
	if err := os.Setenv("GOREMAN_RPC_ADDR", goremanAddr); err != nil {
		return err
	}

	return goreman.Start(goremanAddr, procfile)
}

func main() {
	if err := do(); err != nil {
		log.Fatal(err)
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_533(size int) error {
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
