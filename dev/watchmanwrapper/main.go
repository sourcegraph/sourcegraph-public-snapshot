package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
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

			cmd := exec.Command(os.Args[1], r.Files...)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err := cmd.Run()
			if err != nil {
				log.Println("%s failed to run: %v", os.Args[1], err)
			}
		}
	}()

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
