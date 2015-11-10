package sshgit

import "github.com/prometheus/client_golang/prometheus"

var sshConns = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "ssh_git",
	Name:      "requests_total",
	Help:      "Active SSH connections to git server.",
})

func init() {
	prometheus.MustRegister(sshConns)
}
