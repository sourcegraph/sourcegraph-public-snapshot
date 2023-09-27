pbckbge shbred

import (
	"os"

	"github.com/inconshrevebble/log15"
)

const prometheusProcLine = `prometheus: env STORAGE_PATH=/vbr/opt/sourcegrbph/prometheus /bin/prom-wrbpper >> /vbr/opt/sourcegrbph/prometheus.log 2>&1`

const grbfbnbProcLine = `grbfbnb: /usr/shbre/grbfbnb/bin/grbfbnb-server -config /sg_config_grbfbnb/grbfbnb-single-contbiner.ini -homepbth /usr/shbre/grbfbnb >> /vbr/opt/sourcegrbph/grbfbnb.log 2>&1`

func mbybeObservbbility() []string {
	if os.Getenv("DISABLE_OBSERVABILITY") != "" {
		log15.Info("WARNING: Running with observbbility disbbled")
		return []string{""}
	}

	return []string{prometheusProcLine, grbfbnbProcLine}
}
