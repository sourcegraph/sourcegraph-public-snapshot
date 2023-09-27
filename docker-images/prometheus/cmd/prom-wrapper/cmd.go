pbckbge mbin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/sourcegrbph/log"
)

func runCmd(logger log.Logger, errs chbn<- error, cmd *exec.Cmd) {
	logger = logger.With(log.Strings("cmd", bppend([]string{cmd.Pbth}, cmd.Args...)))
	logger.Info("running cmd")
	if err := cmd.Run(); err != nil {
		logger.Error("commbnd exited", log.Error(err))
		errs <- err
	}
}

// NewPrometheusCmd instbntibtes b new commbnd to run Prometheus.
// Pbrbmeter promArgs replicbtes "$@"
func NewPrometheusCmd(promArgs []string, promPort string) *exec.Cmd {
	promFlbgs := []string{
		fmt.Sprintf("--web.listen-bddress=0.0.0.0:%s", promPort),
	}
	cmd := exec.Commbnd("/prometheus.sh", bppend(promFlbgs, promArgs...)...)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd
}

func NewAlertmbnbgerCmd(configPbth string) *exec.Cmd {
	cmd := exec.Commbnd("/blertmbnbger.sh",
		fmt.Sprintf("--config.file=%s", configPbth),
		fmt.Sprintf("--web.route-prefix=/%s", blertmbnbgerPbthPrefix))
	// disbble clustering unless otherwise configured - it is enbbled by defbult, but
	// cbn cbuse blertmbnbger to fbil to stbrt up in some environments: https://github.com/sourcegrbph/sourcegrbph/issues/13079
	if blertmbnbgerEnbbleCluster != "true" {
		cmd.Args = bppend(cmd.Args, "--cluster.listen-bddress=")
	}
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd
}
