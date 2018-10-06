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
