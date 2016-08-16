package main

import (
	"errors"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/test/load"

	"context"
)

func main() {
	err := Main()
	if err != nil {
		log.Fatal(err)
	}
}

// Main is the entry point for the cli
func Main() error {
	var (
		lt       load.LoadTest
		endpoint string
		duration time.Duration
		err      error
	)
	flag.StringVar(&endpoint, "endpoint", "", "Endpoint to load test (eg https://staging.sourcegraph.com)")
	flag.StringVar(&lt.Username, "username", "", "Username to authenticate as")
	flag.StringVar(&lt.Password, "password", "", "Password for user")
	flag.BoolVar(&lt.Anonymous, "anonymous", false, "Do not login")
	flag.Uint64Var(&lt.Rate, "rate", 0, "Requests per second")
	flag.DurationVar(&duration, "duration", 0, "How long to run the load test for. 0 means forever")
	flag.DurationVar(&lt.ReportPeriod, "report-period", 10*time.Minute, "Rate at which to report partial metrics")
	flag.Parse()
	lt.TargetPaths = flag.Args()
	lt.Endpoint, err = url.Parse(endpoint)
	if err != nil {
		return err
	}
	if endpoint == "" || lt.Rate == 0 {
		return errors.New("-endpoint, -rate are required")
	}

	// Setup a context and signal listener so we can gracefully quit
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		s := <-sig
		log.Printf("Stopping load test after receiving signal %s", s)
		cancel()
	}()

	if duration > 0 {
		log.Printf("Duration: %s", duration)
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, time.Now().Add(duration))
		defer cancel()
	}

	for {
		err = lt.Run(ctx)
		if ctx.Err() != nil {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}
