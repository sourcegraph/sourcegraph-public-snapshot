package userpasswd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricsAccountFailedSignInAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_frontend_account_failed_sign_in_attempts_total",
		Help: "Total number of failed sign-in attempts",
	})
	metricsAccountLockouts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_frontend_account_lockouts_total",
		Help: "Total number of account lockout",
	})
)
