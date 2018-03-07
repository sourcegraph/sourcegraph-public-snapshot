package trace

// UserLatencyBuckets is a recommended list of buckets for use in prometheus
// histograms when measuring latency to users.
// Motivation: longer than 30s we don't care about. 2 is a general SLA we
// have. Otherwise rest is somewhat evenly spreadout to get good data
var UserLatencyBuckets = []float64{0.2, 0.5, 1, 2, 5, 10, 30}
