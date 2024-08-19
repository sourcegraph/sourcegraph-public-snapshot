package transmission

// Metrics is an interface that can be fulfilled by something like statsd
type Metrics interface {
	Gauge(string, interface{})
	Increment(string)
	Count(string, interface{})
}

type nullMetrics struct{}

func (nm *nullMetrics) Gauge(string, interface{}) {}
func (nm *nullMetrics) Increment(string)          {}
func (nm *nullMetrics) Count(string, interface{}) {}
