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

var neverRead = (chan<- []string)(make(chan []string))

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

	pending := make(chan []string)
	changed := make(chan []string)

	// reads stdout of watchman process and sends the changed files on the
	// pending channel.
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

			pending <- r.Files
		}
	}()

	// reads pending and sends to changed. If there are pending files but
	// handle changed process is running, this goroutine will continue to read
	// from pending and merge the changed files into the list.
	go func() {
		seen := map[string]struct{}{}
		var files []string
		for {
			// if we have no files to send then we make changedC a channel
			// that blocks so we effectively only read pending.
			changedC := neverRead
			if len(files) > 0 {
				changedC = changed
			}

			select {
			case fs := <-pending:
				for _, f := range fs {
					if _, ok := seen[f]; !ok {
						seen[f] = struct{}{}
						files = append(files, f)
					}
				}

			case changedC <- files:
				files = nil
				seen = map[string]struct{}{}
			}
		}
	}()

	// reads the changed channel and runs the command os.Args[1] with the
	// files as arguments
	go func() {
		for files := range changed {
			if debug {
				fmt.Fprintln(os.Stderr, "!!! WATCH EVENT", files)
			}

			cmd := exec.Command(os.Args[1], files...)
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
