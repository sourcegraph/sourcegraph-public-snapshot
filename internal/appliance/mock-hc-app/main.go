package main

import (
	"log"
	"os"
	"time"
)

var version = os.Getenv("VERSION")

func healthCheck() {
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println(version, "Doing my health checking... la la la...")
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func somethingElse() {
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println(version, "Doing something else... hmm hmm hmm...")
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func main() {
	healthCheck()
	somethingElse()

	select {}
}
