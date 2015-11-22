package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func main() {
	label := []string{"git"}
	label = append(label, os.Args[1:]...)

	cmd := exec.Command("/usr/bin/git", os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()

	status, err := exitStatus(cmd.Run())
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile("/tmp/git-wrapper.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(f, "%s [%s]\n", label, time.Since(start))
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	os.Exit(status)
}

func exitStatus(err error) (int, error) {
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// There is no platform independent way to retrieve
			// the exit code, but the following will work on Unix
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), nil
			}
		}
		return 0, err
	}
	return 0, nil
}
