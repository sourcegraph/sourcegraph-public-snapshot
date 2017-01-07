package main

import "testing"
import "github.com/prometheus/client_golang/prometheus"

func TestQueue(t *testing.T) {
	enqueue, dequeue := queueWithoutDuplicates(prometheus.NewGauge(prometheus.GaugeOpts{}))
	doDequeue := func() indexTask {
		c := make(chan indexTask)
		dequeue <- c
		return <-c
	}

	enqueue <- indexTask{repo: "foo"}
	enqueue <- indexTask{repo: "bar"}
	enqueue <- indexTask{repo: "foo"}
	enqueue <- indexTask{repo: "baz"}

	if doDequeue() != (indexTask{repo: "foo"}) {
		t.Fail()
	}
	if doDequeue() != (indexTask{repo: "bar"}) {
		t.Fail()
	}
	if doDequeue() != (indexTask{repo: "baz"}) {
		t.Fail()
	}
}
