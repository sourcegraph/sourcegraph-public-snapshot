package gitcli

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	blockedCommandExecutedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_exec_blocked_command_received",
		Help: "Incremented each time a command not in the allowlist for gitserver is executed",
	})
)
