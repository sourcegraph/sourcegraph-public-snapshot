package cli

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	blockedCommandExecutedCounter = promauto.NewCounter(prometheus.CounterOpts{
		// TODO: Name
		Name: "src_gitserver_exec_blocked_command_received2",
		Help: "Incremented each time a command not in the allowlist for gitserver is executed",
	})
)
