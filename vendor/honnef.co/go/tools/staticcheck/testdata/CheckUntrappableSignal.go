package main

import (
	"os"
	"os/signal"
	"syscall"
)

func fn() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Ignore(os.Signal(syscall.SIGKILL)) // MATCH /cannot be trapped/
	signal.Ignore(os.Kill)                    // MATCH /cannot be trapped/
	signal.Notify(c, os.Kill)                 // MATCH /cannot be trapped/
	signal.Reset(os.Kill)                     // MATCH /cannot be trapped/
	signal.Ignore(syscall.SIGKILL)            // MATCH /cannot be trapped/
	signal.Notify(c, syscall.SIGKILL)         // MATCH /cannot be trapped/
	signal.Reset(syscall.SIGKILL)             // MATCH /cannot be trapped/
	signal.Ignore(syscall.SIGSTOP)            // MATCH /cannot be trapped/
	signal.Notify(c, syscall.SIGSTOP)         // MATCH /cannot be trapped/
	signal.Reset(syscall.SIGSTOP)             // MATCH /cannot be trapped/
}
