pbckbge userpbsswd

import (
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
)

vbr (
	metricsAccountFbiledSignInAttempts = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_frontend_bccount_fbiled_sign_in_bttempts_totbl",
		Help: "Totbl number of fbiled sign-in bttempts",
	})
	metricsAccountLockouts = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_frontend_bccount_lockouts_totbl",
		Help: "Totbl number of bccount lockout",
	})
)
