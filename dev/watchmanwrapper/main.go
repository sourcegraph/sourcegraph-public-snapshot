// Command watchmanwrapper runs watchman subscribe and parses its output to
// trigger another command with the files that have changed. See
// dev/changewatch for how it is invocated.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
)

type Response struct {
	IsFreshInstance bool     `json:"is_fresh_instance"`
	Files           []string `json:"files"`
}

func main() {
	cmd := exec.Command("watchman", "-j", "--server-encoding=json", "-p")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	debug, _ := strconv.ParseBool(os.Getenv("WATCHMAN_DEBUG"))
	if debug {
		fmt.Fprintln(os.Stderr, "!!! WATCHMAN debugging enabled")
	}

	go func() {
		dec := json.NewDecoder(stdout)
		for {
			var r Response
			err := dec.Decode(&r)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Fatal(err)
			}
			if r.IsFreshInstance || len(r.Files) == 0 {
				continue
			}

			if debug {
				fmt.Fprintln(os.Stderr, "!!! WATCH EVENT", r.Files)
			}

			cmd := exec.Command(os.Args[1], r.Files...)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				log.Printf("%s failed to run: %v", os.Args[1], err)
			}
		}
	}()

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
