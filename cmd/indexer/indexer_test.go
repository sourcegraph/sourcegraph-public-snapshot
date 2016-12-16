package main

import "testing"
import "github.com/prometheus/client_golang/prometheus"

func TestQueue(t *testing.T) {
	enqueue, dequeue := queueWithoutDuplicates(prometheus.NewGauge(prometheus.GaugeOpts{}))
	doDequeue := func() string {
		c := make(chan string)
		dequeue <- c
		return <-c
	}

	enqueue <- "foo"
	enqueue <- "bar"
	enqueue <- "foo"
	enqueue <- "baz"

	if doDequeue() != "foo" {
		t.Fail()
	}
	if doDequeue() != "bar" {
		t.Fail()
	}
	if doDequeue() != "baz" {
		t.Fail()
	}
}
