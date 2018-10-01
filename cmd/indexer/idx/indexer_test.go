package idx

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func Test_queueWithoutDuplicates(t *testing.T) {
	enqueue, dequeue := queueWithoutDuplicates(prometheus.NewGauge(prometheus.GaugeOpts{}))
	doDequeue := func() qitem {
		c := make(chan qitem)
		dequeue <- c
		return <-c
	}

	enqueue <- qitem{repo: "foo"}
	enqueue <- qitem{repo: "bar"}
	enqueue <- qitem{repo: "foo"}
	enqueue <- qitem{repo: "baz"}

	q := qitem{repo: "foo"}
	if doDequeue() != q {
		t.Fail()
	}
	q = qitem{repo: "bar"}
	if doDequeue() != q {
		t.Fail()
	}
	q = qitem{repo: "baz"}
	if doDequeue() != q {
		t.Fail()
	}
}
